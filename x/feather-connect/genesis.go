package feather_connect

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/feather-core/x/feather-connect/types"
)

// ValidateGenesis
func ValidateGenesis(data *types.GenesisState) error {
	if data.Params.BlockHeight < 1 {
		panic(fmt.Errorf("blockHeight cannot be less than 1 on genesis state"))
	}

	if err := sdk.ValidateDenom(data.Params.BaseDenom); err != nil {
		panic(fmt.Errorf("invalid denom on genesis state: %s", err))
	}

	params := data.Params.Alliance

	if len(params.Title) == 0 {
		panic(fmt.Errorf("title is empty on genesis state"))
	}

	if len(params.Description) == 0 {
		panic(fmt.Errorf("description is empty on genesis state"))
	}

	if err := sdk.ValidateDenom(params.Denom); err != nil {
		panic(fmt.Errorf("invalid denom on genesis state: %s", err))
	}

	if params.RewardWeight.IsNil() || params.RewardWeight.IsNegative() {
		panic(fmt.Errorf("rewardWeight cannot be negative nor nil on genesis state"))
	}

	if params.TakeRate.IsNil() || params.TakeRate.IsNegative() {
		panic(fmt.Errorf("takeRate cannot be negative nor nil on genesis state"))
	}

	if params.RewardChangeRate.IsNil() || params.RewardChangeRate.IsNegative() {
		panic(fmt.Errorf("rewardChangeRate cannot be negative nor nil on genesis state"))
	}

	if params.RewardChangeInterval < 0 {
		panic(fmt.Errorf("rewardChangeInterval cannot be negative nor nil on genesis state"))
	}

	if params.RewardWeightRange.Min.IsNegative() || params.RewardWeightRange.Max.IsNegative() {
		panic(fmt.Errorf("rewardWeightRange Min or Max cannot be negative on genesis state"))
	}

	return nil
}

func DefaultGenesisState() *types.GenesisState {
	return &types.GenesisState{
		Params: types.DefaultParams(),
	}
}
