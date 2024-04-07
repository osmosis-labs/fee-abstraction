package keeper_test

import (
	"fmt"
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

func (s *KeeperTestSuite) TestGetDecTWAPFromBytes() {
	s.SetupTest()

	data := []byte{10, 19, 50, 49, 52, 50, 56, 53, 55, 49, 52, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48}
	twap, err := s.feeAbsKeeper.GetDecTWAPFromBytes(data)
	require.NoError(s.T(), err)
	require.Equal(s.T(), sdkmath.LegacyMustNewDecFromStr("2.142857140000000000"), twap)
}

// test successful ibc ack
// go test -v -run TestKeeperTestSuite/TestSuccessfulTwapAck  github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper
func (s *KeeperTestSuite) TestSuccessfulTwapAck() {
	s.SetupTest()

	// construct ack packet
	ack := s.generateAckPacket([]abci.ResponseQuery{{
		Code: 0,
		Key:  []byte{10, 19, 50, 49, 52, 50, 56, 53, 55, 49, 52, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48},
	}})
	abciQuery := s.generateQueryRequest()

	// setup env
	err := s.feeAbsKeeper.SetHostZoneConfig(s.ctx, types.HostChainFeeAbsConfig{
		IbcDenom:                IBCDenom,
		OsmosisPoolTokenDenomIn: OsmosisIBCDenom,
		Status:                  types.HostChainFeeAbsStatus_UPDATED,
	})
	require.NoError(s.T(), err)

	err = s.feeAbsKeeper.OnAcknowledgementPacket(s.ctx, ack, abciQuery)
	require.NoError(s.T(), err)
	dec, err := s.feeAbsKeeper.GetTwapRate(s.ctx, IBCDenom)
	require.NoError(s.T(), err)
	require.Equal(s.T(), sdkmath.LegacyMustNewDecFromStr("2.142857140000000000"), dec)
}

// test failed ibc ack
// should increase fallback count
// go test -v -run TestKeeperTestSuite/TestFailedTwapAck  github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper
func (s *KeeperTestSuite) TestFailedTwapAck() {
	s.SetupTest()

	// construct ack packet
	ack := s.generateAckPacket([]abci.ResponseQuery{{
		Code: 1,
	}})
	abciQuery := s.generateQueryRequest()

	// setup env
	hostZoneConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                IBCDenom,
		OsmosisPoolTokenDenomIn: OsmosisIBCDenom,
		Status:                  types.HostChainFeeAbsStatus_UPDATED,
	}
	err := s.feeAbsKeeper.SetHostZoneConfig(s.ctx, hostZoneConfig)
	require.NoError(s.T(), err)

	s.feeAbsKeeper.SetEpochInfo(s.ctx, types.EpochInfo{
		Identifier:   types.DefaultQueryEpochIdentifier,
		Duration:     types.DefaultQueryPeriod,
		CurrentEpoch: 1,
	})
	res, exist := s.feeAbsKeeper.GetEpochInfo(s.ctx, types.DefaultQueryEpochIdentifier)
	require.True(s.T(), exist)
	require.Equal(s.T(), int64(1), res.CurrentEpoch)

	// simulate receiving failed ack packet
	err = s.feeAbsKeeper.OnAcknowledgementPacket(s.ctx, ack, abciQuery)
	require.NoError(s.T(), err)
	exp := s.feeAbsKeeper.GetBlockDelayToQuery(s.ctx, IBCDenom)
	require.Equal(s.T(), int64(2), exp.Jump)
	require.Equal(s.T(), int64(3), exp.FutureEpoch)

	// test query twap after failed, filter should be true, thus not allowing twap query
	s.feeAbsKeeper.SetEpochInfo(s.ctx, types.EpochInfo{
		Identifier:   types.DefaultQueryEpochIdentifier,
		Duration:     types.DefaultQueryPeriod,
		CurrentEpoch: 2,
	})
	reqs, err := s.feeAbsKeeper.HandleOsmosisIbcQuery(s.ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), reqs, 0)

	// test query twap allowed after epoch reaches time, filter should be false
	epochInfo := types.EpochInfo{
		Identifier:   types.DefaultQueryEpochIdentifier,
		Duration:     types.DefaultQueryPeriod,
		CurrentEpoch: 3,
	}
	s.feeAbsKeeper.SetEpochInfo(s.ctx, epochInfo)

	filter := s.feeAbsKeeper.IbcQueryHostZoneFilter(s.ctx, hostZoneConfig, epochInfo)
	require.False(s.T(), filter)
}

// test correct setting of OUTDATED status
// go test -v -run TestKeeperTestSuite/TestOutdatedStatus  github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper
func (s *KeeperTestSuite) TestOutdatedStatus() {
	s.SetupTest()

	// construct ack packet
	ack := s.generateAckPacket([]abci.ResponseQuery{{
		Code: 1,
	}})
	abciQuery := s.generateQueryRequest()

	// setup env
	hostZoneConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                IBCDenom,
		OsmosisPoolTokenDenomIn: OsmosisIBCDenom,
		Status:                  types.HostChainFeeAbsStatus_UPDATED,
	}
	err := s.feeAbsKeeper.SetHostZoneConfig(s.ctx, hostZoneConfig)
	require.NoError(s.T(), err)

	s.feeAbsKeeper.SetEpochInfo(s.ctx, types.EpochInfo{
		Identifier:   types.DefaultQueryEpochIdentifier,
		Duration:     types.DefaultQueryPeriod,
		CurrentEpoch: 1,
	})
	res, exist := s.feeAbsKeeper.GetEpochInfo(s.ctx, types.DefaultQueryEpochIdentifier)
	require.True(s.T(), exist)
	require.Equal(s.T(), int64(1), res.CurrentEpoch)

	// simulate receiving failed ack packet, exponential backoff jump = 2
	err = s.feeAbsKeeper.OnAcknowledgementPacket(s.ctx, ack, abciQuery)
	require.NoError(s.T(), err)
	exp := s.feeAbsKeeper.GetBlockDelayToQuery(s.ctx, IBCDenom)
	require.Equal(s.T(), int64(2), exp.Jump)
	require.Equal(s.T(), int64(3), exp.FutureEpoch)

	// simulate receiving failed ack packet again at epoch 3, exponential backoff jump = 4
	epochInfo := types.EpochInfo{
		Identifier:   types.DefaultQueryEpochIdentifier,
		Duration:     types.DefaultQueryPeriod,
		CurrentEpoch: 3,
	}
	s.feeAbsKeeper.SetEpochInfo(s.ctx, epochInfo)

	filter := s.feeAbsKeeper.IbcQueryHostZoneFilter(s.ctx, hostZoneConfig, epochInfo)
	require.False(s.T(), filter)

	err = s.feeAbsKeeper.OnAcknowledgementPacket(s.ctx, ack, abciQuery)
	require.NoError(s.T(), err)
	exp = s.feeAbsKeeper.GetBlockDelayToQuery(s.ctx, IBCDenom)
	require.Equal(s.T(), int64(4), exp.Jump)
	require.Equal(s.T(), int64(7), exp.FutureEpoch)

	// as current epoch is 7, connection should be set to outdated
	epochInfo = types.EpochInfo{
		Identifier:   types.DefaultQueryEpochIdentifier,
		Duration:     types.DefaultQueryPeriod,
		CurrentEpoch: 7,
	}
	s.feeAbsKeeper.SetEpochInfo(s.ctx, epochInfo)
	filter = s.feeAbsKeeper.IbcQueryHostZoneFilter(s.ctx, hostZoneConfig, epochInfo)
	require.False(s.T(), filter)

	config, found := s.feeAbsKeeper.GetHostZoneConfig(s.ctx, IBCDenom)
	require.True(s.T(), found)
	require.Equal(s.T(), types.HostChainFeeAbsStatus_OUTDATED, config.Status)

	// assume that the last query is successful, status should be updated
	// Exponential backoff is reset
	ack = s.generateAckPacket([]abci.ResponseQuery{{
		Code: 0,
		Key:  []byte{10, 19, 50, 49, 52, 50, 56, 53, 55, 49, 52, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48},
	}})

	err = s.feeAbsKeeper.OnAcknowledgementPacket(s.ctx, ack, abciQuery)
	require.NoError(s.T(), err)
	config, found = s.feeAbsKeeper.GetHostZoneConfig(s.ctx, IBCDenom)
	require.True(s.T(), found)
	require.Equal(s.T(), types.HostChainFeeAbsStatus_UPDATED, config.Status)
	exp = s.feeAbsKeeper.GetBlockDelayToQuery(s.ctx, IBCDenom)
	require.Equal(s.T(), int64(1), exp.Jump)
	require.Equal(s.T(), int64(0), exp.FutureEpoch)
}

func (s *KeeperTestSuite) generateAckPacket(queryResponses []abci.ResponseQuery) channeltypes.Acknowledgement {
	data := types.CosmosResponse{
		Responses: queryResponses,
	}
	dataBz, err := data.Marshal()
	require.NoError(s.T(), err)
	ackQueryPacket := types.InterchainQueryPacketAck{
		Data: dataBz,
	}
	ackPacketResBz, err := types.ModuleCdc.MarshalJSON(&ackQueryPacket)
	require.NoError(s.T(), err)
	return channeltypes.NewResultAcknowledgement(ackPacketResBz)
}

func (s *KeeperTestSuite) generateQueryRequest() []abci.RequestQuery {
	icqReqData := types.QueryArithmeticTwapToNowRequest{
		BaseAsset:  IBCDenom,
		QuoteAsset: OsmosisIBCDenom,
	}
	icqReqDataBz, err := icqReqData.Marshal()
	require.NoError(s.T(), err)
	return []abci.RequestQuery{{
		Data: icqReqDataBz,
	}}
}

// test unsuccessful osmosisCrooschainSwap due to below swap threshold
func (s *KeeperTestSuite) TestTransferOsmosisCrosschainSwap() {
	tests := []struct {
		name             string
		AmountToTransfer uint64
		hostConfig       types.HostChainFeeAbsConfig
		wantErr          bool
	}{
		{
			"swap zero token",
			0,
			types.HostChainFeeAbsConfig{
				IbcDenom:                IBCDenom,
				OsmosisPoolTokenDenomIn: OsmosisIBCDenom,
				Status:                  types.HostChainFeeAbsStatus_UPDATED,
				PoolId:                  1,
				MinSwapAmount:           1000,
			},
			true,
		},
		{
			"swap amount below min",
			500,
			types.HostChainFeeAbsConfig{
				IbcDenom:                IBCDenom,
				OsmosisPoolTokenDenomIn: OsmosisIBCDenom,
				Status:                  types.HostChainFeeAbsStatus_UPDATED,
				PoolId:                  1,
				MinSwapAmount:           1000,
			},
			true,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			s.FundFeeAbsModuleAccount(s.ctx, tc.AmountToTransfer)
			err := s.feeAbsKeeper.TransferOsmosisCrosschainSwap(s.ctx, tc.hostConfig)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), fmt.Sprintf("invalid amount to transfer, expect minimum %v, got %v", tc.hostConfig.MinSwapAmount, tc.AmountToTransfer))

			} else {
				require.NoError(t, err)
			}
		})
	}
}

// helper function to fund fee abs module account for testing
func (s *KeeperTestSuite) FundFeeAbsModuleAccount(ctx sdk.Context, amount uint64) {
	s.T().Helper()
	err := s.feeAbsApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(IBCDenom, sdk.NewIntFromUint64(amount))))
	require.NoError(s.T(), err)
	err = s.feeAbsApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(IBCDenom, sdk.NewIntFromUint64(amount))))
	require.NoError(s.T(), err)
}
