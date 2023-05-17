package keeper

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	alliancekeeper "github.com/terra-money/alliance/x/alliance/keeper"
	"github.com/terra-money/feather-core/x/feather-connect/types"
)

type Keeper struct {
	paramSpace        paramtypes.Subspace
	IbcKeeper         ibckeeper.Keeper
	IbcTransferKeeper ibctransferkeeper.Keeper
	AllianceKeeper    alliancekeeper.Keeper
}

func NewKeeper(
	paramSpace paramtypes.Subspace,
	ibcKeeper ibckeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
	allianceKeeper alliancekeeper.Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		kt := paramtypes.NewKeyTable().RegisterParamSet(&types.Params{})
		paramSpace = paramSpace.WithKeyTable(kt)
	}
	return Keeper{
		paramSpace:        paramSpace,
		IbcKeeper:         ibcKeeper,
		IbcTransferKeeper: ibcTransferKeeper,
		AllianceKeeper:    allianceKeeper,
	}
}
