package ante_test

import (
	"testing"

	transferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	"github.com/stretchr/testify/require"
	ubermock "go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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
	anteHandler    sdk.AnteHandler
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

	// Setup feeabs keeper
	suite.feeabsKeeper = feeabskeeper.NewKeeper(suite.encCfg.Codec, key, subspace, suite.stakingKeeper, suite.accountKeeper, nil, transferkeeper.Keeper{}, suite.channelKeeper, suite.portKeeper, suite.scopedKeeper)
	suite.clientCtx = client.Context{}.
		WithTxConfig(suite.encCfg.TxConfig)
	require.NoError(t, err)

	// setup txBuilder
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	return suite
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	_        string
	_        func(*AnteTestSuite) TestCaseArgs
	simulate bool
	expPass  bool
	expErr   error
}

type TestCaseArgs struct {
	chainID   string
	accNums   []uint64
	accSeqs   []uint64
	feeAmount sdk.Coins
	gasLimit  uint64
	msgs      []sdk.Msg
	privs     []cryptotypes.PrivKey
}

// DeliverMsgs constructs a tx and runs it through the ante handler. This is used to set the context for a test case, for
// example to test for replay protection.
func (suite *AnteTestSuite) DeliverMsgs(t *testing.T, privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, simulate bool) (sdk.Context, error) {
	t.Helper()
	require.NoError(t, suite.txBuilder.SetMsgs(msgs...))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	tx, txErr := suite.CreateTestTx(privs, accNums, accSeqs, chainID)
	require.NoError(t, txErr)
	return suite.anteHandler(suite.ctx, tx, simulate)
}

func (suite *AnteTestSuite) RunTestCase(t *testing.T, tc TestCase, args TestCaseArgs) {
	t.Helper()
	require.NoError(t, suite.txBuilder.SetMsgs(args.msgs...))
	suite.txBuilder.SetFeeAmount(args.feeAmount)
	suite.txBuilder.SetGasLimit(args.gasLimit)
	// Theoretically speaking, ante handler unit tests should only test
	// ante handlers, but here we sometimes also test the tx creation
	// process.
	tx, txErr := suite.CreateTestTx(args.privs, args.accNums, args.accSeqs, args.chainID)
	newCtx, anteErr := suite.anteHandler(suite.ctx, tx, tc.simulate)

	if tc.expPass {
		require.NoError(t, txErr)
		require.NoError(t, anteErr)
		require.NotNil(t, newCtx)

		suite.ctx = newCtx
	} else {
		switch {
		case txErr != nil:
			require.Error(t, txErr)
			require.ErrorIs(t, txErr, tc.expErr)

		case anteErr != nil:
			require.Error(t, anteErr)
			require.ErrorIs(t, anteErr, tc.expErr)

		default:
			t.Fatal("expected one of txErr, anteErr to be an error")
		}
	}
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

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
}