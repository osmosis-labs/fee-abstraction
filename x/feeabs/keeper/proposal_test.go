package keeper_test

import (
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	apphelpers "github.com/osmosis-labs/fee-abstraction/v7/app/helpers"
	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

func (suite *KeeperTestSuite) TestAddHostZoneProposal() {
	suite.SetupTest()
	addrs := simtestutil.AddTestAddrs(suite.feeAbsApp.BankKeeper, suite.feeAbsApp.StakingKeeper, suite.ctx, 10, valTokens)

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

			legacyProposal, err := govv1types.NewLegacyContent(proposal, authtypes.NewModuleAddress(govtypes.ModuleName).String())
			suite.Require().NoError(err)

			// store proposal
			_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{legacyProposal}, "", "", "", addrs[0])
			suite.Require().NoError(err)

			// execute proposal
			handler := suite.govKeeper.LegacyRouter().GetRoute(proposal.ProposalRoute())
			err = handler(suite.ctx, proposal)
			suite.Require().NoError(err)

			hostChainConfig, err := suite.feeAbsKeeper.GetHostZoneConfig(suite.ctx, tc.hostChainConfig.IbcDenom)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.hostChainConfig, hostChainConfig)

			// store proposal again and it should error
			_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{legacyProposal}, "", "", "", addrs[0])
			suite.Require().Error(err)
		})
	}
}
