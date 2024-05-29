package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/osmosis-labs/fee-abstraction/v8/app"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/keeper"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx          sdk.Context
	feeAbsApp    *app.FeeApp
	feeAbsKeeper keeper.Keeper
	govKeeper    govkeeper.Keeper
	queryClient  types.QueryClient
	msgServer    types.MsgServer
}

const (
	SourcePort                       = "feeabs"
	SourceChannel                    = "channel-0"
	IBCDenom                         = "ibc/1"
	OsmosisIBCDenom                  = "ibc/2"
	TestOsmosisQueryTwapPath         = "/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow"
	TestNativeIbcedInOsmosis         = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
	TestChainName                    = "feeappd-t1"
	TestIbcTransferChannel           = "channel-0"
	TestIbcQueryIcqChannel           = "channel-3"
	TestOsmosisCrosschainSwapAddress = "osmo1abc123"
)

var valTokens = sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)

func (s *KeeperTestSuite) SetupTest() {
	s.feeAbsApp = app.Setup(s.T())
	s.ctx = s.feeAbsApp.NewContextLegacy(true, cmtproto.Header{Height: 1})

	s.feeAbsKeeper = s.feeAbsApp.FeeabsKeeper
	s.govKeeper = s.feeAbsApp.GovKeeper

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.feeAbsApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(s.feeAbsKeeper))
	s.queryClient = types.NewQueryClient(queryHelper)

	s.msgServer = keeper.NewMsgServerImpl(s.feeAbsKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetParams() {
	params := types.Params{
		OsmosisQueryTwapPath: TestOsmosisQueryTwapPath,
		ChainName:            TestChainName,
		NativeIbcedInOsmosis: TestNativeIbcedInOsmosis,
		IbcTransferChannel:   TestIbcTransferChannel,
		IbcQueryIcqChannel:   TestIbcQueryIcqChannel,
	}

	suite.feeAbsKeeper.SetParams(suite.ctx, params)

	suite.Run("stored params are as expected", func() {
		actual := suite.feeAbsKeeper.GetParams(suite.ctx)
		suite.Equal(actual.OsmosisQueryTwapPath, TestOsmosisQueryTwapPath)
		suite.Equal(actual.ChainName, TestChainName)
		suite.Equal(actual.NativeIbcedInOsmosis, TestNativeIbcedInOsmosis)
		suite.Equal(actual.IbcTransferChannel, TestIbcTransferChannel)
		suite.Equal(actual.IbcQueryIcqChannel, TestIbcQueryIcqChannel)
	})
}
