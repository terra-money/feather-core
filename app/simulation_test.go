package app

import (
	"os"
	"testing"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
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

// Running as a go test:
// `go test -run=TestFullAppSimulation ./app -NumBlocks 200 -BlockSize 50 -Commit true Verbose true -Enabled true“
func TestFullAppSimulation(t *testing.T) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = simulationAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(
		config,
		simulationDirPrefix,
		simulationDbName,
		simcli.FlagVerboseValue,
		simcli.FlagEnabledValue,
	)
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue // how often to check for broken invariants in x/crisis

	app := New(logger,
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		uint(1),
		MakeEncodingConfig(),
		appOptions,
		baseapp.SetChainID(simulationAppChainID),
	)
	require.Equal(t, Name, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		map[string]bool{},
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulatino error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}
