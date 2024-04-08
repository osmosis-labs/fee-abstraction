package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/golang/mock/gomock"
	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/ante"
	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
	"github.com/stretchr/testify/require"
)

func TestMempoolDecorator(t *testing.T) {
	gasLimit := uint64(200000)
	testCases := []struct {
		name        string
		feeAmount   sdk.Coins
		minGasPrice sdk.DecCoins
		malleate    func(*AnteTestSuite)
		isErr       bool
		expErr      error
	}{
		{
			"empty fee, should fail",
			sdk.Coins{},
			sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 100))...),
			func(suite *AnteTestSuite) {
			},
			true,
			sdkerrors.ErrInsufficientFee,
		},
		{
			"not enough native fee, should fail",
			sdk.NewCoins(sdk.NewInt64Coin("native", 100)),
			sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...),
			func(suite *AnteTestSuite) {},
			true,
			sdkerrors.ErrInsufficientFee,
		},
		{
			"enough native fee, should pass",
			sdk.NewCoins(sdk.NewInt64Coin("native", 1000*int64(gasLimit))),
			sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...),
			func(suite *AnteTestSuite) {},
			false,
			nil,
		},
		{
			"unknown ibc fee denom, should fail",
			sdk.NewCoins(sdk.NewInt64Coin("ibcfee", 1000*int64(gasLimit))),
			sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...),
			func(suite *AnteTestSuite) {},
			true,
			sdkerrors.ErrInvalidCoins,
		},
		{
			"not enough ibc fee, should fail",
			sdk.NewCoins(sdk.NewInt64Coin("ibcfee", 999*int64(gasLimit))),
			sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...),
			func(suite *AnteTestSuite) {
				err := suite.feeabsKeeper.SetHostZoneConfig(suite.ctx, types.HostChainFeeAbsConfig{
					IbcDenom:                "ibcfee",
					OsmosisPoolTokenDenomIn: "osmosis",
					PoolId:                  1,
					Status:                  types.HostChainFeeAbsStatus_UPDATED,
					MinSwapAmount:           0,
				})
				require.NoError(t, err)
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", sdk.NewDec(1))
				suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("native").MinTimes(1)
			},
			true,
			sdkerrors.ErrInsufficientFee,
		},

		{
			"enough ibc fee, should pass",
			sdk.NewCoins(sdk.NewInt64Coin("ibcfee", 1000*int64(gasLimit))),
			sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...),
			func(suite *AnteTestSuite) {
				err := suite.feeabsKeeper.SetHostZoneConfig(suite.ctx, types.HostChainFeeAbsConfig{
					IbcDenom:                "ibcfee",
					OsmosisPoolTokenDenomIn: "osmosis",
					PoolId:                  1,
					Status:                  types.HostChainFeeAbsStatus_UPDATED,
					MinSwapAmount:           0,
				})
				require.NoError(t, err)
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", sdk.NewDec(1))
				suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("native").MinTimes(1)
			},
			false,
			nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			// Setup test context
			tc.malleate(suite)
			suite.txBuilder.SetGasLimit(gasLimit)
			suite.txBuilder.SetFeeAmount(tc.feeAmount)
			suite.ctx = suite.ctx.WithIsCheckTx(true)
			suite.ctx = suite.ctx.WithMinGasPrices(tc.minGasPrice)

			// Construct tx and run through mempool decorator
			tx := suite.txBuilder.GetTx()
			mempoolDecorator := ante.NewFeeAbstrationMempoolFeeDecorator(suite.feeabsKeeper)
			antehandler := sdk.ChainAnteDecorators(mempoolDecorator)

			// Run the ante handler
			_, err := antehandler(suite.ctx, tx, false)

			if tc.isErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}

}
