package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/feather-core/x/feather/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, g types.GenesisState) []abci.ValidatorUpdate {
	k.SetParams(ctx, g.Params)
	return []abci.ValidatorUpdate{}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
