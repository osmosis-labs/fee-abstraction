package keeper_test

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (s *KeeperTestSuite) TestOsmosisArithmeticTwap() {
	s.SetupTest()
	twapPrice := sdk.NewDec(1)
	s.feeAbsKeeper.SetTwapRate(s.ctx, "denom", twapPrice)

	for _, tc := range []struct {
		desc      string
		req       *types.QueryOsmosisArithmeticTwapRequest
		res       *types.QueryOsmosisArithmeticTwapResponse
		shouldErr bool
	}{
		{
			desc: "Success",
			req: &types.QueryOsmosisArithmeticTwapRequest{
				IbcDenom: "denom",
			},
			res: &types.QueryOsmosisArithmeticTwapResponse{
				ArithmeticTwap: twapPrice,
			},
			shouldErr: false,
		},
		{
			desc: "Invalid denom",
			req: &types.QueryOsmosisArithmeticTwapRequest{
				IbcDenom: "invalid",
			},
			shouldErr: true,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			goCtx := sdk.WrapSDKContext(s.ctx)
			if !tc.shouldErr {
				res, err := s.queryClient.OsmosisArithmeticTwap(goCtx, tc.req)
				s.Require().NoError(err)
				s.Require().Equal(tc.res, res)
			} else {
				_, err := s.queryClient.OsmosisArithmeticTwap(goCtx, tc.req)
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestHostChainConfig() {
	s.SetupTest()

	chainConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                randStringRunes(10),
		OsmosisPoolTokenDenomIn: randStringRunes(10),
		PoolId:                  randUint64Num(),
	}

	err := s.feeAbsKeeper.SetHostZoneConfig(s.ctx, chainConfig.IbcDenom, chainConfig)
	s.Require().NoError(err)

	for _, tc := range []struct {
		desc      string
		req       *types.QueryHostChainConfigRequest
		res       *types.QueryHostChainConfigRespone
		shouldErr bool
	}{
		{
			desc: "Success",
			req: &types.QueryHostChainConfigRequest{
				IbcDenom: chainConfig.IbcDenom,
			},
			res: &types.QueryHostChainConfigRespone{
				HostChainConfig: chainConfig,
			},
			shouldErr: false,
		},
		{
			desc: "Invalid denom",
			req: &types.QueryHostChainConfigRequest{
				IbcDenom: "Invalid",
			},
			res: &types.QueryHostChainConfigRespone{
				HostChainConfig: chainConfig,
			},
			shouldErr: true,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			goCtx := sdk.WrapSDKContext(s.ctx)
			if !tc.shouldErr {
				res, err := s.queryClient.HostChainConfig(goCtx, tc.req)
				s.Require().NoError(err)
				s.Require().Equal(tc.res, res)
			} else {
				_, err := s.queryClient.HostChainConfig(goCtx, tc.req)
				s.Require().NoError(err)
			}
		})
	}
}

func randStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func randUint64Num() uint64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint64()
}
