package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/terra-money/feather-core/app"
)

func initSDKConfig() {
	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.AccountAddressPrefix, app.AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(app.ValidatorAddressPrefix, app.ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(app.ConsensusNodeAddressPrefix, app.ConsensusNodePubKeyPrefix)
	config.Seal()
}
