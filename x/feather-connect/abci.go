package feather_connect

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/terra-money/feather-core/x/feather-connect/keeper"
	"github.com/terra-money/feather-core/x/feather-connect/types"
)

func EndBlock(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	conf := types.NewVerifierConfig()

	if ctx.BlockHeight() != conf.BlockHeight {
		return []abci.ValidatorUpdate{}
	}

	channels := k.IbcKeeper.ChannelKeeper.GetAllChannels(ctx)
	for _, channel := range channels {

		if channel.PortId == ibctransfertypes.ModuleName {
			clientState, _ := k.IbcKeeper.ClientKeeper.GetClientState(ctx, channel.ConnectionHops[0])

			if clientState.GetLatestHeight().GetRevisionHeight() == uint64(conf.BlockHeight) {
				denomTraces := k.IbcTransferKeeper.GetAllDenomTraces(ctx)

				for _, denomTrace := range denomTraces {
					if denomTrace.BaseDenom == conf.BaseDenom {
						conf.SetAllianceDenom(denomTrace.IBCDenom())
						k.AllianceKeeper.CreateAlliance(ctx, &conf.Alliance)
						return []abci.ValidatorUpdate{}
					}
				}
			}
		}
	}
	panic(
		fmt.Sprintf(
			"No IBC channels with port 'transfer' and denom '%v' at height '%v'",
			conf.BaseDenom,
			conf.BlockHeight,
		),
	)
}
