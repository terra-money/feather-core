package mantlemint

import (
	"github.com/cometbft/cometbft/state"
	tendermint "github.com/cometbft/cometbft/types"
)

type Mantlemint interface {
	Inject(*tendermint.Block) error
	Init(*tendermint.GenesisDoc) error
	LoadInitialState() error
	GetCurrentHeight() int64
	GetCurrentBlock() *tendermint.Block
	GetCurrentState() state.State
	GetCurrentEventCollector() *EventCollector
	SetBlockExecutor(executor Executor)
}

type Executor interface {
	ApplyBlock(state.State, tendermint.BlockID, *tendermint.Block) (state.State, int64, error)
	SetEventBus(publisher tendermint.BlockEventPublisher)
}

type MantlemintCallbackBefore func(block *tendermint.Block) error
type MantlemintCallbackAfter func(block *tendermint.Block, events *EventCollector) error

// --- internal types
