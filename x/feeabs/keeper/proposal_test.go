package keeper_test

import (
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	apphelpers "github.com/osmosis-labs/fee-abstraction/v8/app"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func (s *KeeperTestSuite) TestAddHostZoneProposal() {
	s.SetupTest()
	addrs := simtestutil.AddTestAddrs(s.feeAbsApp.BankKeeper, s.feeAbsApp.StakingKeeper, s.ctx, 10, valTokens)

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
				Status:                  types.HostChainFeeAbsStatus_UPDATED,
			},
			shouldErr: false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			proposal := apphelpers.AddHostZoneProposalFixture(func(p *types.AddHostZoneProposal) {
				p.HostChainConfig = &tc.hostChainConfig
			})

			legacyProposal, err := govv1types.NewLegacyContent(proposal, authtypes.NewModuleAddress(govtypes.ModuleName).String())
			s.Require().NoError(err)

			// store proposal
			_, err = s.govKeeper.SubmitProposal(s.ctx, []sdk.Msg{legacyProposal}, "", "", "", addrs[0])
			s.Require().NoError(err)

			// execute proposal
			handler := s.govKeeper.LegacyRouter().GetRoute(proposal.ProposalRoute())
			err = handler(s.ctx, proposal)
			s.Require().NoError(err)

			hostChainConfig, found := s.feeAbsKeeper.GetHostZoneConfig(s.ctx, tc.hostChainConfig.IbcDenom)
			s.Require().True(found)
			s.Require().Equal(tc.hostChainConfig, hostChainConfig)

			// store proposal again and it should error
			_, err = s.govKeeper.SubmitProposal(s.ctx, []sdk.Msg{legacyProposal}, "", "", "", addrs[0])
			s.Require().Error(err)
		})
	}
}

func (s *KeeperTestSuite) TestDeleteHostZoneProposal() {
	s.SetupTest()
	addrs := simtestutil.AddTestAddrs(s.feeAbsApp.BankKeeper, s.feeAbsApp.StakingKeeper, s.ctx, 10, valTokens)

	hostChainConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
		OsmosisPoolTokenDenomIn: "ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E",
		PoolId:                  1,
		Status:                  types.HostChainFeeAbsStatus_UPDATED,
	}

	addProposal := &types.AddHostZoneProposal{
		Title:           "AddHostZoneProposal Title",
		Description:     "AddHostZoneProposal Description",
		HostChainConfig: &hostChainConfig,
	}

	legacyProposal, err := govv1types.NewLegacyContent(addProposal, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	s.Require().NoError(err)

	// Store proposal
	_, err = s.govKeeper.SubmitProposal(s.ctx, []sdk.Msg{legacyProposal}, "", "", "", addrs[0])
	s.Require().NoError(err)

	// Execute proposal
	handler := s.govKeeper.LegacyRouter().GetRoute(addProposal.ProposalRoute())
	err = handler(s.ctx, addProposal)
	s.Require().NoError(err)

	hostChainConfig, found := s.feeAbsKeeper.GetHostZoneConfig(s.ctx, hostChainConfig.IbcDenom)
	s.Require().True(found)
	s.Require().Equal(hostChainConfig, hostChainConfig)

	testCases := []struct {
		desc           string
		deleteProposal *types.DeleteHostZoneProposal
		shouldError    bool
	}{
		{
			desc: "should success when delete an exists host zone config.",
			deleteProposal: &types.DeleteHostZoneProposal{
				Title:       "DeleteHostZoneProposal Title",
				Description: "DeleteHostZoneProposal Description",
				IbcDenom:    "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
			},
			shouldError: false,
		},
		{
			deleteProposal: &types.DeleteHostZoneProposal{
				Title:       "DeleteHostZoneProposal Title",
				Description: "DeleteHostZoneProposal Description",
				IbcDenom:    "ibc/00000",
			},
			desc:        "should error when delete a not exists host zone config.",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.desc, func() {
			legacyProposal, err := govv1types.NewLegacyContent(tc.deleteProposal, authtypes.NewModuleAddress(govtypes.ModuleName).String())
			s.Require().NoError(err)

			// Store proposal
			_, err = s.govKeeper.SubmitProposal(s.ctx, []sdk.Msg{legacyProposal}, "", "", "", addrs[0])
			if !tc.shouldError {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}

			// Execute proposal
			handler = s.govKeeper.LegacyRouter().GetRoute(addProposal.ProposalRoute())
			err = handler(s.ctx, tc.deleteProposal)
			s.Require().NoError(err)
		})
	}
}
