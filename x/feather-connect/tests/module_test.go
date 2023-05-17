package tests

import (
	"testing"

	acbitypes "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/terra-money/feather-core/app"
	test_utils "github.com/terra-money/feather-core/x/feather-connect/tests/utils"
)

func TestEndBlock(t *testing.T) {
	app := app.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 2})

	app.FeatherConnectKeeper.InitGenesis(ctx, test_utils.MockedParams)
	app.IBCKeeper.ChannelKeeper.SetChannel(ctx, "transfer", "channel-0", test_utils.MockedChannel)
	app.TransferKeeper.SetDenomTrace(ctx, test_utils.MockedDenomTrace)
	app.IBCKeeper.ClientKeeper.SetClientState(ctx, "transfer", test_utils.MockedClientState)

	app.EndBlock(acbitypes.RequestEndBlock{Height: 2})
}
