package block

import (
	"fmt"

	tmjson "github.com/cometbft/cometbft/libs/json"
	tm "github.com/cometbft/cometbft/types"
	"github.com/terra-money/feather-core/app"
	"github.com/terra-money/feather-core/mantlemint/db/safe_batch"
	"github.com/terra-money/feather-core/mantlemint/indexer"
	"github.com/terra-money/feather-core/mantlemint/mantlemint"
)

var IndexBlock = indexer.CreateIndexer(func(indexerDB safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, _ *mantlemint.EventCollector, _ *app.App) error {
	defer fmt.Printf("[indexer/block] indexing done for height %d\n", block.Height)
	record := BlockRecord{
		Block:   block,
		BlockID: blockID,
	}

	recordJSON, recordErr := tmjson.Marshal(record)
	if recordErr != nil {
		return recordErr
	}

	return indexerDB.Set(getKey(uint64(block.Height)), recordJSON)
})
