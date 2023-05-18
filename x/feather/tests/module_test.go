package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	acbitypes "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/terra-money/feather-core/app"
	test_utils "github.com/terra-money/feather-core/x/feather/tests/utils"
	"github.com/terra-money/feather-core/x/feather/types"
)

func TestEndBlockHappyPath(t *testing.T) {
	// GIVEN application configuration ...
	app := app.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 2})

	// ... setup a mocked genesis for feather module ...
	app.FeatherKeeper.InitGenesis(ctx, test_utils.MockedGenesis)

	// ... and ibc module with the necessary data to detect it on EndBlocker.
	app.IBCKeeper.ChannelKeeper.SetChannel(ctx, "transfer", "channel-0", test_utils.MockedChannel)
	app.IBCKeeper.ConnectionKeeper.SetConnection(ctx, "connection-0", test_utils.MockedConnection)
	app.TransferKeeper.SetDenomTrace(ctx, test_utils.MockedDenomTrace)
	app.IBCKeeper.ClientKeeper.SetClientState(ctx, "transfer", test_utils.MockedClientState)

	// WHEN the end block is executed.
	app.EndBlock(acbitypes.RequestEndBlock{Height: 2})

	// THEN validate that
	// ... the alliance has been created...
	alliance, found := app.AllianceKeeper.GetAssetByDenom(ctx, test_utils.IBC_DENOM)
	require.True(t, found)
	require.NotEmpty(t, alliance.Denom)

	// ... the module params have been updated with the correct denom.
	params := app.FeatherKeeper.GetParams(ctx)
	require.Equal(t, params.Alliance.Denom, test_utils.IBC_DENOM)
}

func TestEndBlockWithoutHalting(t *testing.T) {
	// GIVEN application configuration ...
	app := app.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 2})

	// ... setup a mocked genesis for feather module ...
	app.FeatherKeeper.InitGenesis(ctx, types.GenesisState{
		Params: types.Params{
			HaltIfNoChannel:    false,
			AllianceBondHeight: 2,
		},
	})

	// WHEN the end block is executed.
	res := app.EndBlock(acbitypes.RequestEndBlock{Height: 2})

	// THEN validate that it did not throw an error...
	require.NotNil(t, res)
}

func TestEndBlockHalting(t *testing.T) {
	// THEN validate that it throw an error.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but no panic occurred")
		}
	}()

	// GIVEN application configuration ...
	app := app.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 2})
	// ... setup a mocked genesis for feather module ...
	app.FeatherKeeper.InitGenesis(ctx, test_utils.MockedGenesis)

	// WHEN the end block is executed.
	app.EndBlock(acbitypes.RequestEndBlock{Height: 2})

}
