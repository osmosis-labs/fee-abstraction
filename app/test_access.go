package app

import (
	"testing"

	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"

	"github.com/osmosis-labs/fee-abstraction/v8/app/params"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/keeper"
)

type TestSupport struct {
	t   testing.TB
	app *FeeApp
}

func NewTestSupport(tb testing.TB, app *FeeApp) *TestSupport {
	tb.Helper()
	return &TestSupport{t: tb, app: app}
}

func (s TestSupport) IBCKeeper() *ibckeeper.Keeper {
	return s.app.IBCKeeper
}

func (s TestSupport) AppCodec() codec.Codec {
	return s.app.appCodec
}

func (s TestSupport) ScopeIBCKeeper() capabilitykeeper.ScopedKeeper {
	return s.app.ScopedIBCKeeper
}

func (s TestSupport) ScopedTransferKeeper() capabilitykeeper.ScopedKeeper {
	return s.app.ScopedTransferKeeper
}

func (s TestSupport) StakingKeeper() stakingkeeper.Keeper {
	return *s.app.StakingKeeper
}

func (s TestSupport) BankKeeper() bankkeeper.Keeper {
	return s.app.BankKeeper
}

func (s TestSupport) TransferKeeper() ibctransferkeeper.Keeper {
	return s.app.TransferKeeper
}

func (s TestSupport) FeeAbsKeeper() keeper.Keeper {
	return s.app.FeeabsKeeper
}

func (s TestSupport) GetBaseApp() *baseapp.BaseApp {
	return s.app.BaseApp
}

func (TestSupport) GetTxConfig() client.TxConfig {
	return params.MakeEncodingConfig().TxConfig
}
