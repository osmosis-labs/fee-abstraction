package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/osmosis-labs/fee-abstraction/v7/app"
	apphelpers "github.com/osmosis-labs/fee-abstraction/v7/app/helpers"
	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper"
	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
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

func (s *KeeperTestSuite) SetupTest() {
	s.feeAbsApp = apphelpers.Setup(s.T(), false, 1)
	s.ctx = s.feeAbsApp.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})
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
