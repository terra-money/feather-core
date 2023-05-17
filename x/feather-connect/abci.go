package feather_connect

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/terra-money/feather-core/x/feather-connect/keeper"
)

// Each end block, check if the current block height is the same as the alliance bond height.
// If so, check if there is a channel with port 'transfer' and specified denom in the module params.
// If so, create an alliance with the specified params.
func EndBlock(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	params := k.GetParams(ctx)

	// if HaltIfNoChannel parameter is false, do nothing.
	if !params.HaltIfNoChannel {
		return []abci.ValidatorUpdate{}
	}

	// if the block height is not yet the specified in theparameter, do nothing.
	if ctx.BlockHeight() != params.AllianceBondHeight {
		return []abci.ValidatorUpdate{}
	}

	// get all IBC channels and iterate over them ...
	channels := k.IbcKeeper.ChannelKeeper.GetAllChannels(ctx)
	for _, channel := range channels {
		// ... if the channel is not 'transfer' channel, continue to next channel.
		if channel.PortId != ibctransfertypes.ModuleName {
			continue
		}
		// ... if the channel is not opened, continue to next channel.
		if channel.State != channeltypes.OPEN {
			continue
		}

		// Rtreive all denom traces ...
		denomTraces := k.IbcTransferKeeper.GetAllDenomTraces(ctx)
		// ... if there are not denom traces, continue to next channel.
		if len(denomTraces) == 0 {
			continue
		}

		// Retreive the channel client state ...
		_, clientState, err := k.IbcKeeper.ChannelKeeper.GetChannelClientState(ctx, channel.PortId, channel.ChannelId)
		if err != nil {
			panic(err)
		}
		ibctm, _ := clientState.(*ibctm.ClientState)
		// If the channel client state's chain id is not the
		// one defined in the module params continue to next channel.
		if ibctm.ChainId != params.BaseChainId {
			continue
		}

		// Iterate over all denom traces and check if there is a denom trace
		// with the same base denom as the one defined in the module params,
		// if so, create an alliance with the specified params, update
		// the module params and return.
		for _, denomTrace := range denomTraces {
			if denomTrace.BaseDenom == params.BaseDenom {
				params.Alliance.Denom = denomTrace.IBCDenom()

				k.AllianceKeeper.CreateAlliance(ctx, &params.Alliance)
				k.SetParams(ctx, params)

				return []abci.ValidatorUpdate{}
			}
		}
	}

	// If none of the previous conditions are met halt the chain
	panic(
		fmt.Sprintf(
			"No IBC channels with port 'transfer' and denom '%v' at height '%v'",
			params.BaseDenom,
			params.AllianceBondHeight,
		),
	)
}
