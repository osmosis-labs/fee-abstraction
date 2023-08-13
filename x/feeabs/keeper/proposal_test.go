package keeper_test

import (
	apphelpers "github.com/osmosis-labs/fee-abstraction/v2/app/helpers"
	"github.com/osmosis-labs/fee-abstraction/v2/x/feeabs/types"
)

func (s *KeeperTestSuite) TestAddHostZoneProposal() {
	s.SetupTest()

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
		s.Run(tc.desc, func() {
			proposal := apphelpers.AddHostZoneProposalFixture(func(p *types.AddHostZoneProposal) {
				p.HostChainConfig = &tc.hostChainConfig
			})

			// store proposal
			storedProposal, err := s.govKeeper.SubmitProposal(s.ctx, proposal)
			s.Require().NoError(err)

			// execute proposal
			handler := s.govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
			err = handler(s.ctx, storedProposal.GetContent())
			s.Require().NoError(err)

			hostChainConfig, err := s.feeAbsKeeper.GetHostZoneConfig(s.ctx, tc.hostChainConfig.IbcDenom)
			s.Require().NoError(err)
			s.Require().Equal(tc.hostChainConfig, hostChainConfig)

			// store proposal again and it should error
			_, err = s.govKeeper.SubmitProposal(s.ctx, proposal)
			s.Require().Error(err)
		})
	}
}
