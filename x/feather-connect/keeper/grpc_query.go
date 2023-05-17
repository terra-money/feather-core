package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/feather-core/x/feather-connect/types"
)

type QueryServer struct {
	Keeper
}

var _ types.QueryServer = QueryServer{}

func (k QueryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	// Define a variable that will store the params
	var params types.Params

	// Get context with the information about the environment
	ctx := sdk.UnwrapSDKContext(c)

	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &QueryServer{Keeper: keeper}
}
