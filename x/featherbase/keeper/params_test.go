package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/terra-money/feather-base/testutil/keeper"
	"github.com/terra-money/feather-base/x/featherbase/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.FeatherbaseKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
