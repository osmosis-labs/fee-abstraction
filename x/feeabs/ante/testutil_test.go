package ante_test

import (
	"testing"

	transferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	"github.com/stretchr/testify/require"
	ubermock "go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper"
	feeabstestutil "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/testutil"
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  authtypes.AccountI
	priv cryptotypes.PrivKey
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	ctx            sdk.Context
	clientCtx      client.Context
	txBuilder      client.TxBuilder
	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     *feeabstestutil.MockBankKeeper
	feeGrantKeeper *feeabstestutil.MockFeegrantKeeper
	stakingKeeper  *feeabstestutil.MockStakingKeeper
	feeabsKeeper   feeabskeeper.Keeper
	channelKeeper  *feeabstestutil.MockChannelKeeper
	portKeeper     *feeabstestutil.MockPortKeeper
	scopedKeeper   *feeabstestutil.MockScopedKeeper
	encCfg         moduletestutil.TestEncodingConfig
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func SetupTestSuite(t *testing.T, isCheckTx bool) *AnteTestSuite {
	t.Helper()
	suite := &AnteTestSuite{}
	ctrl := ubermock.NewController(t)

	// Setup mock keepers
	suite.bankKeeper = feeabstestutil.NewMockBankKeeper(ctrl)
	suite.stakingKeeper = feeabstestutil.NewMockStakingKeeper(ctrl)
	suite.feeGrantKeeper = feeabstestutil.NewMockFeegrantKeeper(ctrl)
	suite.channelKeeper = feeabstestutil.NewMockChannelKeeper(ctrl)
	suite.portKeeper = feeabstestutil.NewMockPortKeeper(ctrl)
	suite.scopedKeeper = feeabstestutil.NewMockScopedKeeper(ctrl)

	// setup necessary params for Account Keeper
	key := sdk.NewKVStoreKey(feeabstypes.StoreKey)
	authKey := sdk.NewKVStoreKey(authtypes.StoreKey)
	subspace := paramtypes.NewSubspace(nil, nil, nil, nil, "feeabs")
	subspace = subspace.WithKeyTable(feeabstypes.ParamKeyTable())
	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		"multiPerm":              {"burner", "minter", "staking"},
		"random":                 {"random"},
		"feeabs":                 nil,
	}

	// setup context for Account Keeper
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	testCtx.CMS.MountStoreWithDB(authKey, storetypes.StoreTypeIAVL, testCtx.DB)
	testCtx.CMS.MountStoreWithDB(sdk.NewTransientStoreKey("transient_test2"), storetypes.StoreTypeTransient, testCtx.DB)
	err := testCtx.CMS.LoadLatestVersion()
	require.NoError(t, err)
	suite.ctx = testCtx.Ctx.WithIsCheckTx(isCheckTx).WithBlockHeight(1) // app.BaseApp.NewContext(isCheckTx, tmproto.Header{}).WithBlockHeight(1)

	suite.encCfg = moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	suite.encCfg.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(suite.encCfg.InterfaceRegistry)
	suite.accountKeeper = authkeeper.NewAccountKeeper(
		suite.encCfg.Codec, authKey, authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix, authtypes.NewModuleAddress("gov").String(),
	)
	suite.accountKeeper.SetModuleAccount(suite.ctx, authtypes.NewEmptyModuleAccount(feeabstypes.ModuleName))
	// Setup feeabs keeper
	suite.feeabsKeeper = feeabskeeper.NewKeeper(suite.encCfg.Codec, key, subspace, suite.stakingKeeper, suite.accountKeeper, nil, transferkeeper.Keeper{}, suite.channelKeeper, suite.portKeeper, suite.scopedKeeper)
	suite.clientCtx = client.Context{}.
		WithTxConfig(suite.encCfg.TxConfig)
	require.NoError(t, err)

	// setup txBuilder
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	return suite
}

func (suite *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		if err != nil {
			panic(err)
		}
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}
