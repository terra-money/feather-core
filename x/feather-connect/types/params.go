package types

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
)

var (
	BlockHeight = []byte("blockHeight")
	BaseDenom   = []byte("baseDenom")
	Alliance    = []byte("alliance")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(BlockHeight, &p.BlockHeight, validateBlockHeight),
		paramtypes.NewParamSetPair(BaseDenom, &p.BaseDenom, validateBaseDenom),
		paramtypes.NewParamSetPair(Alliance, &p.Alliance, validateAlliance),
	}
}

func validateBlockHeight(i interface{}) error {
	blockHeight, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if blockHeight < 0 {
		return fmt.Errorf("blockHeight must be positive: %d", blockHeight)
	}
	return nil
}

func validateBaseDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := sdk.ValidateDenom(denom); err != nil {
		return fmt.Errorf("invalid denom: %s", err)
	}

	return nil
}

func validateAlliance(i interface{}) error {
	_, ok := i.(alliancetypes.MsgCreateAllianceProposal)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		BlockHeight: 1000,
		BaseDenom:   "uluna",
		Alliance: alliancetypes.MsgCreateAllianceProposal{
			Title:                "Alliance with Terra",
			Description:          "Asset uluna creates an alliance with the chain to increase the economical security of the chain.",
			RewardWeight:         sdkmath.LegacyNewDec(1),
			TakeRate:             sdkmath.LegacyNewDec(1),
			RewardChangeRate:     sdkmath.LegacyNewDec(1),
			RewardChangeInterval: time.Duration(0),
			RewardWeightRange: alliancetypes.RewardWeightRange{
				Min: sdkmath.LegacyNewDec(0),
				Max: sdkmath.LegacyNewDec(0),
			},
		},
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}
