package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/terra-money/feather-core/x/feather/types"
)

func genAllianceBondHeight(r *rand.Rand) int64 {
	return int64(simulation.RandIntBetween(r, 1, 99))
}

// RandomizedGenState generates a random GenesisState for staking
func RandomizedGenState(simState *module.SimulationState) {
	params := types.DefaultParams()

	params.AllianceBondHeight = genAllianceBondHeight(simState.Rand)

	featherGenesis := types.GenesisState{
		Params: params,
	}

	bz, err := json.MarshalIndent(&featherGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Select randomly generated feather parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&featherGenesis)
}
