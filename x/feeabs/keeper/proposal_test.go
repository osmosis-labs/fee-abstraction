package keeper_test

import (
	apphelpers "github.com/notional-labs/feeabstraction/v2/app/helpers"
	"github.com/notional-labs/feeabstraction/v2/x/feeabs/types"
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
				IbcDenom:                   "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
				OsmosisPoolTokenDenomIn:    "ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E",
				MiddlewareAddress:          "cosmos1alc8mjana7ssgeyffvlfza08gu6rtav8rmj6nv",
				IbcTransferChannel:         "channel-2",
				HostZoneIbcTransferChannel: "channel-1",
				CrosschainSwapAddress:      "osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r",
				PoolId:                     1,
				IsOsmosis:                  false,
				Frozen:                     false,
				OsmosisQueryChannel:        "channel-1",
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
