package keeper_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

var govAcc = authtypes.NewEmptyModuleAccount(govtypes.ModuleName, authtypes.Minter)

func (suite *KeeperTestSuite) TestMsgUpdateParams() {
	// default params
	params := feeabstypes.DefaultParams()

	testCases := []struct {
		name      string
		input     *feeabstypes.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &feeabstypes.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "all good",
			input: &feeabstypes.MsgUpdateParams{
				Authority: suite.feeAbsKeeper.GetAuthority(),
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateParams(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
