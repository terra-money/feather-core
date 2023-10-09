package block

import (
	tm "github.com/cometbft/cometbft/types"
	"github.com/terra-money/feather-core/mantlemint/lib"
)

var prefix = []byte("block/height:")
var getKey = func(height uint64) []byte {
	return lib.ConcatBytes(prefix, lib.UintToBigEndian(height))
}

type BlockRecord struct {
	BlockID *tm.BlockID `json:"block_id"`
	Block   *tm.Block   `json:"block"`
}
