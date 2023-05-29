package types

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
)

var (
	HaltIfNoChannel    = []byte("HaltIfNoChannel")
	BaseDenom          = []byte("BaseDenom")
	BaseChainId        = []byte("BaseChainId")
	AllianceBondHeight = []byte("AllianceBondHeight")
	Alliance           = []byte("Alliance")
)

var _ paramtypes.ParamSet = (*Params)(nil)

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(HaltIfNoChannel, &p.HaltIfNoChannel, validateHaltIfNoChannel),
		paramtypes.NewParamSetPair(BaseDenom, &p.BaseDenom, validateBaseDenom),
		paramtypes.NewParamSetPair(BaseChainId, &p.BaseChainId, validateBaseChainId),
		paramtypes.NewParamSetPair(AllianceBondHeight, &p.AllianceBondHeight, validateAllianceBondHeight),
		paramtypes.NewParamSetPair(Alliance, &p.Alliance, validateAlliance),
	}
}

func validateHaltIfNoChannel(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBaseDenom(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
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
		HaltIfNoChannel:    false,
		BaseDenom:          "uluna",
		BaseChainId:        "phoenix-1",
		AllianceBondHeight: 1000,
		Alliance: alliancetypes.MsgCreateAllianceProposal{
			Title:                "Alliance with Terra",
			Description:          "Asset uluna creates an alliance with the chain to increase the economical security of the chain.",
			RewardWeight:         sdkmath.LegacyMustNewDecFromStr("0.1"),
			TakeRate:             sdkmath.LegacyNewDec(0),
			RewardChangeRate:     sdkmath.LegacyNewDec(1),
			RewardChangeInterval: time.Duration(0),
			RewardWeightRange: alliancetypes.RewardWeightRange{
				Min: sdkmath.LegacyMustNewDecFromStr("0.1"),
				Max: sdkmath.LegacyMustNewDecFromStr("0.1"),
			},
		},
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}
