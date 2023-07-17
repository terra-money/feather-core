package app

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Hardcoded chainID for simulation.
const (
	simulationAppChainID = "simulation-app"
)

var (
	// hacky workaround to restart a test with another seed
	simFailedDueToEmptyValSet = false
)

func init() {
	simcli.GetSimulatorFlags()
}

// Simulate app with invariant checks every operation. If a broken invariant is found, prints all failed invariants.
//
//	go test -run=TestFullAppSimulation ./app -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -v -timeout 24h
func TestFullAppSimulation(t *testing.T) {
	// this is a workaround for https://github.com/CosmWasm/wasmd/issues/1437
	// if simulation generates an empty validator set, restart the simulation with a new seed
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("%v", r)
			if !strings.Contains(err, "validator set is empty after InitGenesis") {
				panic(r)
			}
			t.Log("Simulation generated empty validator set - restarting simulation")
			simFailedDueToEmptyValSet = true
			TestFullAppSimulation(t)
		}
	}()

	if !simcli.FlagEnabledValue {
		t.Skip("skipping full application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.ChainID = simulationAppChainID

	// seems like cosmos sdk doesn't actually use these flags any longer, so they have no effect...
	config.OnOperation = true
	config.AllInvariants = true

	// if no seed is provided, generate a random one
	if config.Seed == simcli.DefaultSeedValue {
		config.Seed = rand.Int63()
	}

	// if simulation failed due to empty validator set, restart with new seed
	if simFailedDueToEmptyValSet {
		config.Seed = rand.Int63()
		simFailedDueToEmptyValSet = false
	}

	// if no period is provided, default to 1
	if simcli.FlagPeriodValue == 0 {
		simcli.FlagPeriodValue = 1
	}

	var logger log.Logger
	if simcli.FlagVerboseValue {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	db := dbm.NewMemDB()
	defer func() {
		require.NoError(t, db.Close())
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue
	appOptions[flags.FlagHome] = DefaultNodeHome

	app := New(
		logger,
		db,
		nil,
		true,
		nil,
		DefaultNodeHome,
		simcli.FlagPeriodValue,
		MakeEncodingConfig(),
		appOptions,
		baseapp.SetChainID(simulationAppChainID),
		fauxMerkleModeOpt,
	)
	require.Equal(t, AppName, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.BankKeeper.GetBlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before checking simulation error result
	err := simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

// Simulate app with 3 randomly generated seeds. For each seed, run 3 times each, and check that resulting app hashes are the same.
//
//	go test -v -run=TestAppStateDeterminism ./app -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h
//
// Run with one seed override (i.e. any seed that isn't simcli.DefaultSeedValue), 3 times:
//
//	go test -v -run=TestAppStateDeterminism ./app -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h -Seed=100
func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application state determinism simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.ChainID = simulationAppChainID

	// seems like cosmos sdk doesn't actually use these flags any longer, so they have no effect...
	config.OnOperation = false
	config.AllInvariants = false

	numSeeds := 3
	numTimesToRunPerSeed := 3
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	simulateWithSeedNTimes := func(seed int64, times int) {
		cfg := config
		cfg.Seed = seed

		for i := 0; i < times; i++ {
			var logger log.Logger
			if simcli.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			defer db.Close()

			appOptions := make(simtestutil.AppOptionsMap, 0)
			appOptions[flags.FlagHome] = DefaultNodeHome
			appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

			app := New(logger,
				db,
				nil,
				true,
				map[int64]bool{},
				DefaultNodeHome,
				simcli.FlagPeriodValue,
				MakeEncodingConfig(),
				appOptions,
				baseapp.SetChainID(simulationAppChainID),
				fauxMerkleModeOpt,
			)
			require.Equal(t, AppName, app.Name())

			fmt.Printf(
				"running non-determinism simulation; seed %d: attempt: %d/%d\n",
				seed, i+1, numTimesToRunPerSeed,
			)

			// run randomized simulation
			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
				simtypes.RandomAccounts,
				simtestutil.SimulationOperations(app, app.AppCodec(), cfg),
				app.BankKeeper.GetBlockedAddresses(),
				cfg,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if cfg.Commit {
				simtestutil.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[i] = appHash

			if i != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[i]),
					"non-determinism in seed %d: attempt: %d/%d\n", cfg.Seed, i+1, numTimesToRunPerSeed,
				)
			}
		}
	}

	switch config.Seed {
	case simcli.DefaultSeedValue:
		// no seed override, run simulation with multiple seeds
		for i := 0; i < numSeeds; i++ {
			seed := rand.Int63()
			fmt.Printf("running non-determinism simulation; seed %d: %d/%d\n", seed, i+1, numSeeds)
			simulateWithSeedNTimes(seed, numTimesToRunPerSeed)
		}
	default:
		// seed override provided - user may be debugging; run once
		fmt.Printf("running non-determinism simulation; seed %d\n", config.Seed)
		simulateWithSeedNTimes(config.Seed, numTimesToRunPerSeed)
	}
}

// Performs the following steps:
// 1. Runs a simulation with a randomly generated seed.
// 2. Exports the resulting application state as a genesis file (same as CLI command `feather-cored export`)
// 3. Imports the genesis file.
// 4. Verifies that the resulting application state is the same as step 1.
// 5. Runs another simulation, but with the same seed as step 1.
//
//	go test -v -run=TestAppImportExport ./app -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h
func TestAppImportExport(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application import/export simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.ChainID = simulationAppChainID

	// seems like cosmos sdk doesn't actually use these flags any longer, so they have no effect...
	config.OnOperation = false
	config.AllInvariants = false

	// if no seed is provided, generate a random one
	if config.Seed == simcli.DefaultSeedValue {
		config.Seed = rand.Int63()
	}

	var logger log.Logger
	if simcli.FlagVerboseValue {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	db := dbm.NewMemDB()
	defer func() {
		require.NoError(t, db.Close())
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue
	appOptions[flags.FlagHome] = DefaultNodeHome

	app := New(
		logger,
		db,
		nil,
		true,
		nil,
		DefaultNodeHome,
		simcli.FlagPeriodValue,
		MakeEncodingConfig(),
		appOptions,
		baseapp.SetChainID(simulationAppChainID),
		fauxMerkleModeOpt,
	)
	require.Equal(t, AppName, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.BankKeeper.GetBlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before checking simulation error result
	err := simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	db2 := dbm.NewMemDB()
	defer func() {
		require.NoError(t, db2.Close())
	}()

	app2 := New(
		logger,
		db2,
		nil,
		true,
		nil,
		DefaultNodeHome,
		simcli.FlagPeriodValue,
		MakeEncodingConfig(),
		appOptions,
		baseapp.SetChainID(simulationAppChainID),
		fauxMerkleModeOpt,
	)
	require.Equal(t, AppName, app2.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("%v", r)
			if !strings.Contains(err, "validator set is empty after InitGenesis") {
				panic(r)
			}
			logger.Info("Skipping simulation as all validators have been unbonded")
			logger.Info("err", err, "stacktrace", string(debug.Stack()))
		}
	}()

	ctxA := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	ctxB := app2.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	app2.ModuleManager.InitGenesis(ctxB, app.AppCodec(), genesisState)
	app2.StoreConsensusParams(ctxB, exported.ConsensusParams)

	fmt.Printf("comparing stores...\n")

	type StoreKeysPrefixes struct {
		A        storetypes.StoreKey
		B        storetypes.StoreKey
		Prefixes [][]byte
	}

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.UnsafeGetKey(authtypes.StoreKey), app2.UnsafeGetKey(authtypes.StoreKey), [][]byte{}},
		{
			app.UnsafeGetKey(stakingtypes.StoreKey), app2.UnsafeGetKey(stakingtypes.StoreKey),
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingIDKey, stakingtypes.UnbondingIndexKey, stakingtypes.UnbondingTypeKey, stakingtypes.ValidatorUpdatesKey,
			},
		}, // ordering may change but it doesn't matter
		{app.UnsafeGetKey(slashingtypes.StoreKey), app2.UnsafeGetKey(slashingtypes.StoreKey), [][]byte{}},
		{app.UnsafeGetKey(minttypes.StoreKey), app2.UnsafeGetKey(minttypes.StoreKey), [][]byte{}},
		{app.UnsafeGetKey(distrtypes.StoreKey), app2.UnsafeGetKey(distrtypes.StoreKey), [][]byte{}},
		{app.UnsafeGetKey(banktypes.StoreKey), app2.UnsafeGetKey(banktypes.StoreKey), [][]byte{banktypes.BalancesPrefix}},
		{app.UnsafeGetKey(paramtypes.StoreKey), app2.UnsafeGetKey(paramtypes.StoreKey), [][]byte{}},
		{app.UnsafeGetKey(govtypes.StoreKey), app2.UnsafeGetKey(govtypes.StoreKey), [][]byte{}},
		{app.UnsafeGetKey(evidencetypes.StoreKey), app2.UnsafeGetKey(evidencetypes.StoreKey), [][]byte{}},
		{app.UnsafeGetKey(capabilitytypes.StoreKey), app2.UnsafeGetKey(capabilitytypes.StoreKey), [][]byte{}},
		{app.UnsafeGetKey(authzkeeper.StoreKey), app2.UnsafeGetKey(authzkeeper.StoreKey), [][]byte{authzkeeper.GrantKey, authzkeeper.GrantQueuePrefix}},
		// TODO add keys for non-default cosmos sdk modules
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, 0, len(failedKVAs), simtestutil.GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}

	// simulation with the exact same parameters should pass
	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.BankKeeper.GetBlockedAddresses(),
		config,
		app.AppCodec(),
	)
	require.NoError(t, err)
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}
