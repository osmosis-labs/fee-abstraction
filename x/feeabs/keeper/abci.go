package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// BeginBlocker of epochs module.
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		logger := k.Logger(ctx)

		// If blocktime < initial epoch start time, return
		if ctx.BlockTime().Before(epochInfo.StartTime) {
			return
		}
		// if epoch counting hasn't started, signal we need to start.
		shouldInitialEpochStart := !epochInfo.EpochCountingStarted

		epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		shouldEpochStart := (ctx.BlockTime().After(epochEndTime)) || shouldInitialEpochStart

		if !shouldEpochStart {
			return false
		}
		epochInfo.CurrentEpochStartHeight = ctx.BlockHeight()
		// TODO: need create function to this

		if shouldInitialEpochStart {
			epochInfo.EpochCountingStarted = true
			epochInfo.CurrentEpoch = 1
			epochInfo.CurrentEpochStartTime = epochInfo.StartTime
			logger.Info(fmt.Sprintf("Starting new epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
		} else {
			k.executeAllHostChainTWAPQuery(ctx)
			k.executeAllHostChainSwap(ctx)
		}

		// emit new epoch start event, set epoch info, and run BeforeEpochStart hook
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEpochStart,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
				sdk.NewAttribute(types.AttributeEpochStartTime, fmt.Sprintf("%d", epochInfo.CurrentEpochStartTime.Unix())),
			),
		)
		k.setEpochInfo(ctx, epochInfo)

		return false
	})
}

// executeAllHostChainTWAPQuery will iterate all hostZone and send the IBC Query Packet to Osmosis chain.
func (k Keeper) executeAllHostChainTWAPQuery(ctx sdk.Context) {
	err := k.handleOsmosisIbcQuery(ctx)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error executeAllHostChainTWAPQuery %s", err.Error()))
	}
}

// executeAllHostChainTWAPSwap will iterate all hostZone and execute swap over chain.
func (k Keeper) executeAllHostChainSwap(ctx sdk.Context) {
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		var err error

		if hostZoneConfig.Frozen {
			return false
		}

		if hostZoneConfig.IsOsmosis {
			err = k.transferIBCTokenToOsmosisChainWithIBCHookMemo(ctx, hostZoneConfig)
		} else {
			err = k.transferIBCTokenToHostChainWithMiddlewareMemo(ctx, hostZoneConfig)
		}

		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to transfer IBC token %s", err.Error()))
			err = k.FronzenHostZoneByIBCDenom(ctx, hostZoneConfig.IbcDenom)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Failed to frozem host zone %s", err.Error()))
			}
		}

		return false
	})
}
