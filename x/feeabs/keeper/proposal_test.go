package keeper_test

import (
	apphelpers "github.com/notional-labs/fee-abstraction/v2/app/helpers"
	"github.com/notional-labs/fee-abstraction/v2/x/feeabs/types"
)

func (suite *KeeperTestSuite) TestAddHostZoneProposal() {
	suite.SetupTest()

	for _, tc := range []struct {
		desc            string
		hostChainConfig types.HostChainFeeAbsConfig
		shouldErr       bool
	}{
		{
			desc: "Success",
			hostChainConfig: types.HostChainFeeAbsConfig{
				IbcDenom:                "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
				OsmosisPoolTokenDenomIn: "ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E",
				PoolId:                  1,
				Frozen:                  false,
			},
			shouldErr: false,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			proposal := apphelpers.AddHostZoneProposalFixture(func(p *types.AddHostZoneProposal) {
				p.HostChainConfig = &tc.hostChainConfig
			})

			// store proposal
			storedProposal, err := suite.govKeeper.SubmitProposal(suite.ctx, proposal)
			suite.Require().NoError(err)

			// execute proposal
			handler := suite.govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
			err = handler(suite.ctx, storedProposal.GetContent())
			suite.Require().NoError(err)

			hostChainConfig, err := suite.feeAbsKeeper.GetHostZoneConfig(suite.ctx, tc.hostChainConfig.IbcDenom)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.hostChainConfig, hostChainConfig)

			// store proposal again and it should error
			_, err = suite.govKeeper.SubmitProposal(suite.ctx, proposal)
			suite.Require().Error(err)
		})
	}
}
