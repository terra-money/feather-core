package featherbase_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "github.com/terra-money/feather-base/testutil/keeper"
	"github.com/terra-money/feather-base/testutil/nullify"
	"github.com/terra-money/feather-base/x/featherbase"
	"github.com/terra-money/feather-base/x/featherbase/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.FeatherbaseKeeper(t)
	featherbase.InitGenesis(ctx, *k, genesisState)
	got := featherbase.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
