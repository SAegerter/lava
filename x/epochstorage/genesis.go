package epochstorage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lavanet/lava/x/epochstorage/keeper"
	"github.com/lavanet/lava/x/epochstorage/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the stakeStorage
	for _, elem := range genState.StakeStorageList {
		k.SetStakeStorage(ctx, elem)
	}
	// Set if defined
	if genState.EpochDetails != nil {
		k.SetEpochDetails(ctx, *genState.EpochDetails)
	}
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.StakeStorageList = k.GetAllStakeStorage(ctx)
	// Get all epochDetails
	epochDetails, found := k.GetEpochDetails(ctx)
	if found {
		genesis.EpochDetails = &epochDetails
	}
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
