package keeper_test

import (
	"fmt"
	"testing"

	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/osmosis-labs/fee-abstraction/v4/app"
	apphelpers "github.com/osmosis-labs/fee-abstraction/v4/app/helpers"
	"github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/keeper"
	"github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx          sdk.Context
	feeAbsApp    *app.FeeAbs
	feeAbsKeeper keeper.Keeper
	govKeeper    govkeeper.Keeper
	queryClient  types.QueryClient
	msgServer    types.MsgServer
}

const (
	SourcePort      = "feeabs"
	SourceChannel   = "channel-0"
	IBCDenom        = "ibc/1"
	OsmosisIBCDenom = "ibc/2"
)

var valTokens = sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)

func (suite *KeeperTestSuite) SetupTest() {
	suite.feeAbsApp = apphelpers.Setup(suite.T(), false, 1)
	suite.ctx = suite.feeAbsApp.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})
	suite.feeAbsKeeper = suite.feeAbsApp.FeeabsKeeper
	suite.govKeeper = suite.feeAbsApp.GovKeeper

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.feeAbsApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(suite.feeAbsKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)

	suite.msgServer = keeper.NewMsgServerImpl(suite.feeAbsKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
