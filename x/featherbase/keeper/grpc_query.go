package keeper

import (
	"github.com/terra-money/feather-base/x/featherbase/types"
)

var _ types.QueryServer = Keeper{}
