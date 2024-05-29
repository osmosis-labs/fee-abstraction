package keeper_test

import (
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func (s *KeeperTestSuite) TestMsgUpdateParams() {
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
				Authority: s.feeAbsKeeper.GetAuthority(),
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.input)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
