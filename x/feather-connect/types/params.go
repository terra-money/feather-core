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
	HaltIfNoChannel    = []byte("haltIfNoChannel")
	BaseDenom          = []byte("baseDenom")
	BaseChainId        = []byte("baseChainId")
	AllianceBondHeight = []byte("allianceBondHeight")
	Alliance           = []byte("alliance")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(BaseDenom, &p.BaseDenom, validateBaseDenom),
		paramtypes.NewParamSetPair(BaseChainId, &p.BaseChainId, validateBaseChainId),
		paramtypes.NewParamSetPair(AllianceBondHeight, &p.AllianceBondHeight, validateAllianceBondHeight),
		paramtypes.NewParamSetPair(Alliance, &p.Alliance, validateAlliance),
	}
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

func validateBaseChainId(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateAllianceBondHeight(i interface{}) error {
	AllianceBondHeight, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if AllianceBondHeight < 0 {
		return fmt.Errorf("AllianceBondHeight must be positive: %d", AllianceBondHeight)
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
		HaltIfNoChannel:    true,
		BaseDenom:          "uluna",
		BaseChainId:        "phoenix-1",
		AllianceBondHeight: 1000,
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
