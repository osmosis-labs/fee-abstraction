package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/feeabstraction/v1/app"
	apphelpers "github.com/notional-labs/feeabstraction/v1/app/helpers"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/keeper"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx          sdk.Context
	feeAbsApp    *app.FeeAbs
	feeAbsKeeper keeper.Keeper
	queryClient  types.QueryClient
	msgServer    types.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.feeAbsApp = apphelpers.Setup(suite.T(), false, 1)
	suite.ctx = suite.feeAbsApp.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})
	suite.feeAbsKeeper = suite.feeAbsApp.FeeabsKeeper

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.feeAbsApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(suite.feeAbsKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)

	suite.msgServer = keeper.NewMsgServerImpl(suite.feeAbsKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
