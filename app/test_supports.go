package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/keeper"
)

func (app *FeeApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

func (app *FeeApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

func (app *FeeApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

func (app *FeeApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

func (app *FeeApp) GetStakingKeeper() *stakingkeeper.Keeper {
	return app.StakingKeeper
}

func (app *FeeApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

func (app *FeeApp) GetFeeabsKeeper() feeabskeeper.Keeper {
	return app.FeeabsKeeper
}
