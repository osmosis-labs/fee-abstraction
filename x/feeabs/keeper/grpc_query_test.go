package keeper_test

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (suite *KeeperTestSuite) TestOsmosisArithmeticTwap() {
	suite.SetupTest()
	twapPrice := sdk.NewDec(1)
	suite.feeAbsKeeper.SetTwapRate(suite.ctx, "denom", twapPrice)

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
		suite.Run(tc.desc, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			if !tc.shouldErr {
				res, err := suite.queryClient.OsmosisArithmeticTwap(goCtx, tc.req)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.res, res)
			} else {
				_, err := suite.queryClient.OsmosisArithmeticTwap(goCtx, tc.req)
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestHostChainConfig() {
	suite.SetupTest()

	chainConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                   randStringRunes(10),
		OsmosisPoolTokenDenomIn:    randStringRunes(10),
		MiddlewareAddress:          randStringRunes(10),
		IbcTransferChannel:         randStringRunes(10),
		HostZoneIbcTransferChannel: randStringRunes(10),
		CrosschainSwapAddress:      randStringRunes(10),
		PoolId:                     randUint64Num(),
	}

	err := suite.feeAbsKeeper.SetHostZoneConfig(suite.ctx, chainConfig.IbcDenom, chainConfig)
	suite.Require().NoError(err)

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
		suite.Run(tc.desc, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			if !tc.shouldErr {
				res, err := suite.queryClient.HostChainConfig(goCtx, tc.req)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.res, res)
			} else {
				_, err := suite.queryClient.HostChainConfig(goCtx, tc.req)
				suite.Require().NoError(err)
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
