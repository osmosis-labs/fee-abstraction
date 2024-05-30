package ante_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/ante"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func TestMempoolDecorator(t *testing.T) {
	gasLimit := uint64(200000)
	// mockHostZoneConfig is used to mock the host zone config, with ibcfee as the ibc fee denom to be used as alternative fee
	mockHostZoneConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                "ibcfee",
		OsmosisPoolTokenDenomIn: "osmosis",
		PoolId:                  1,
		Status:                  types.HostChainFeeAbsStatus_UPDATED,
	}
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
				err := suite.feeabsKeeper.SetHostZoneConfig(suite.ctx, mockHostZoneConfig)
				require.NoError(t, err)
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", math.LegacyNewDec(1))
				suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("native", nil).MinTimes(1)
			},
			true,
			sdkerrors.ErrInsufficientFee,
		},

		{
			"enough ibc fee, should pass",
			sdk.NewCoins(sdk.NewInt64Coin("ibcfee", 1000*int64(gasLimit))),
			sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...),
			func(suite *AnteTestSuite) {
				err := suite.feeabsKeeper.SetHostZoneConfig(suite.ctx, mockHostZoneConfig)
				require.NoError(t, err)
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", math.LegacyNewDec(1))
				suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("native", nil).MinTimes(1)
			},
			false,
			nil,
		},
		// TODO: Add support for multiple denom fees(--fees 50ibc,50native)
		// {
		// 	"half native fee, half ibc fee, should pass",
		// 	sdk.NewCoins(sdk.NewInt64Coin("native", 500*int64(gasLimit)), sdk.NewInt64Coin("ibcfee", 500*int64(gasLimit))),
		// 	sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...),
		// 	func(suite *AnteTestSuite) {
		// 		err := suite.feeabsKeeper.SetHostZoneConfig(suite.ctx, types.HostChainFeeAbsConfig{
		// 			IbcDenom:                "ibcfee",
		// 			OsmosisPoolTokenDenomIn: "osmosis",
		// 			PoolId:                  1,
		// 			Status:                  types.HostChainFeeAbsStatus_UPDATED,
		// 			MinSwapAmount:           0,
		// 		})
		// 		require.NoError(t, err)
		// 		suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", math.LegacyNewDec(1))
		// 		suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("native").MinTimes(1)
		// 	},
		// 	false,
		// 	nil,
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := SetupTestSuite(t, true)

			tc.malleate(suite)
			suite.txBuilder.SetGasLimit(gasLimit)
			suite.txBuilder.SetFeeAmount(tc.feeAmount)
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

func TestDeductFeeDecorator(t *testing.T) {
	gasLimit := uint64(200000)
	minGasPrice := sdk.NewDecCoinsFromCoins(sdk.NewCoins(sdk.NewInt64Coin("native", 1000))...)
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin("native", 1000*int64(gasLimit)))
	ibcFeeAmount := sdk.NewCoins(sdk.NewInt64Coin("ibcfee", 1000*int64(gasLimit)))
	// mockHostZoneConfig is used to mock the host zone config, with ibcfee as the ibc fee denom to be used as alternative fee
	mockHostZoneConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                "ibcfee",
		OsmosisPoolTokenDenomIn: "osmosis",
		PoolId:                  1,
		Status:                  types.HostChainFeeAbsStatus_UPDATED,
	}
	testCases := []struct {
		name     string
		malleate func(*AnteTestSuite)
		isErr    bool
		expErr   error
	}{
		{
			"not enough native fee in balance, should fail",
			func(suite *AnteTestSuite) {
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", math.LegacyNewDec(1))
				// suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), types.ModuleName, feeAmount).Return(sdkerrors.ErrInsufficientFee).MinTimes(1)
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), authtypes.FeeCollectorName, feeAmount).Return(sdkerrors.ErrInsufficientFee).MinTimes(1)
			},
			true,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"enough native fee in balance, should pass",
			func(suite *AnteTestSuite) {
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", math.LegacyNewDec(1))
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), authtypes.FeeCollectorName, feeAmount).Return(nil).MinTimes(1)
			},
			false,
			nil,
		},
		{
			"not enough ibc fee in balance, should fail",
			func(suite *AnteTestSuite) {
				err := suite.feeabsKeeper.SetHostZoneConfig(suite.ctx, mockHostZoneConfig)
				require.NoError(t, err)
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", math.LegacyNewDec(1))
				suite.txBuilder.SetFeeAmount(ibcFeeAmount)
				suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("native", nil).MinTimes(1)
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), types.ModuleName, ibcFeeAmount).Return(sdkerrors.ErrInsufficientFunds).MinTimes(1)
			},
			true,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"enough ibc fee in balance, should pass",
			func(suite *AnteTestSuite) {
				err := suite.feeabsKeeper.SetHostZoneConfig(suite.ctx, mockHostZoneConfig)
				require.NoError(t, err)
				suite.feeabsKeeper.SetTwapRate(suite.ctx, "ibcfee", math.LegacyNewDec(1))
				suite.txBuilder.SetFeeAmount(ibcFeeAmount)
				suite.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("native", nil).MinTimes(1)
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), types.ModuleName, ibcFeeAmount).Return(nil).MinTimes(1)
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), authtypes.FeeCollectorName, feeAmount).Return(nil).MinTimes(1)
			},
			false,
			nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			acc := suite.CreateTestAccounts(1)[0]
			// default value for gasLimit, feeAmount, feePayer. Use native token fee as default
			suite.txBuilder.SetGasLimit(gasLimit)
			suite.txBuilder.SetFeeAmount(feeAmount)
			suite.txBuilder.SetFeePayer(acc.acc.GetAddress())
			suite.ctx = suite.ctx.WithMinGasPrices(minGasPrice)

			// mallate the test case, e.g. setup to pay fee in IBC token
			tc.malleate(suite)

			// Construct tx and run through mempool decorator
			tx := suite.txBuilder.GetTx()
			deductFeeDecorator := ante.NewFeeAbstractionDeductFeeDecorate(suite.accountKeeper, suite.bankKeeper, suite.feeabsKeeper, suite.feeGrantKeeper)
			antehandler := sdk.ChainAnteDecorators(deductFeeDecorator)
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
