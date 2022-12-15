package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/terra-money/feather-base/testutil/keeper"
	"github.com/terra-money/feather-base/x/featherbase/keeper"
	"github.com/terra-money/feather-base/x/featherbase/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.FeatherbaseKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
