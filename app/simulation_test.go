package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	"github.com/stretchr/testify/require"
)

// Hardcoded chainID for simulation.
const (
	simulationAppChainID = "simulation-app"
	simulationDirPrefix  = "leveldb-app-sim"
	simulationDbName     = "Simulation"
)

func init() {
	simcli.GetSimulatorFlags()
}

// Run with 3 randomly generated seeds, 5 times each:
// 
// 	go test -v -run=TestAppStateDeterminism ./app -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h
// 
// Run with one seed override (i.e. any seed that isn't simcli.DefaultSeedValue), 5 times:
// 
// 	go test -v -run=TestAppStateDeterminism ./app -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h -Seed=100
func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.ChainID = simulationAppChainID
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false

	numSeeds := 3
	numTimesToRunPerSeed := 5
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

			app := New(logger,
				db,
				nil,
				true,
				map[int64]bool{},
				DefaultNodeHome,
				simcli.FlagPeriodValue,
				MakeEncodingConfig(),
				simtestutil.EmptyAppOptions{},
				baseapp.SetChainID(simulationAppChainID),
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
