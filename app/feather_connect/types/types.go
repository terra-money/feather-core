package types

import (
	"time"

	sdkmath "cosmossdk.io/math"

	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
)

const ModuleName = "feather_connect"

type VerifierConfig struct {
	BlockHeight int64
	BaseDenom   string
	Alliance    alliancetypes.MsgCreateAllianceProposal
}

func NewVerifierConfig() VerifierConfig {

	return VerifierConfig{
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

func (v VerifierConfig) SetAllianceDenom(denom string) {
	v.Alliance.Denom = denom
}
