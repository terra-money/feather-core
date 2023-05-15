package keeper

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	alliancekeeper "github.com/terra-money/alliance/x/alliance/keeper"
	"github.com/terra-money/feather-core/app/feather_connect/types"
)

type Keeper struct {
	IbcKeeper         ibckeeper.Keeper
	IbcTransferKeeper ibctransferkeeper.Keeper
	AllianceKeeper    alliancekeeper.Keeper
}

func NewKeeper(
	ibcKeeper ibckeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
	allianceKeeper alliancekeeper.Keeper,
) Keeper {
	return Keeper{
		IbcKeeper:         ibcKeeper,
		IbcTransferKeeper: ibcTransferKeeper,
		AllianceKeeper:    allianceKeeper,
	}
}

func EndBlock(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	conf := types.NewVerifierConfig()

	if ctx.BlockHeight() != conf.BlockHeight {
		return []abci.ValidatorUpdate{}
	}

	channels := k.IbcKeeper.ChannelKeeper.GetAllChannels(ctx)
	for _, channel := range channels {

		if channel.PortId == ibctransfer.ModuleName {
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
