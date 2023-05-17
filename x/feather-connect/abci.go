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

func EndBlock(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	params := k.GetParams(ctx)

	if !params.HaltIfNoChannel {
		return []abci.ValidatorUpdate{}
	}

	if ctx.BlockHeight() != params.AllianceBondHeight {
		return []abci.ValidatorUpdate{}
	}

	channels := k.IbcKeeper.ChannelKeeper.GetAllChannels(ctx)
	for _, channel := range channels {
		if channel.PortId == ibctransfertypes.ModuleName {
			denomTraces := k.IbcTransferKeeper.GetAllDenomTraces(ctx)

			if channel.State == channeltypes.OPEN {
				_, clientState, err := k.IbcKeeper.ChannelKeeper.GetChannelClientState(ctx, channel.PortId, channel.ChannelId)
				if err != nil {
					panic(err)
				}
				ibctm, _ := clientState.(*ibctm.ClientState)

				if ibctm.ChainId == params.BaseChainId {
					for _, denomTrace := range denomTraces {
						if denomTrace.BaseDenom == params.BaseDenom {
							params.Alliance.Denom = denomTrace.IBCDenom()

							k.AllianceKeeper.CreateAlliance(ctx, &params.Alliance)
							k.SetParams(ctx, params)

							return []abci.ValidatorUpdate{}
						}
					}
				}
			}
		}
	}
	panic(
		fmt.Sprintf(
			"No IBC channels with port 'transfer' and denom '%v' at height '%v'",
			params.BaseDenom,
			params.AllianceBondHeight,
		),
	)
}
