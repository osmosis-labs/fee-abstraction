package ante

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errorstypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper"
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

type FeeAbstractionDeductFeeDecorate struct {
	accountKeeper  AccountKeeper
	bankKeeper     BankKeeper
	feeabsKeeper   feeabskeeper.Keeper
	feegrantKeeper FeegrantKeeper
}

func NewFeeAbstractionDeductFeeDecorate(ak AccountKeeper, bk BankKeeper, feeabsKeeper feeabskeeper.Keeper, fk FeegrantKeeper) FeeAbstractionDeductFeeDecorate {
	return FeeAbstractionDeductFeeDecorate{
		accountKeeper:  ak,
		bankKeeper:     bk,
		feeabsKeeper:   feeabsKeeper,
		feegrantKeeper: fk,
	}
}

func (fadfd FeeAbstractionDeductFeeDecorate) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(errorstypes.ErrTxDecode, "Tx must be a FeeTx")
	}

	fee := feeTx.GetFee()
	if len(fee) == 0 {
		return fadfd.normalDeductFeeAnteHandle(ctx, tx, simulate, next, feeTx)
	}

	feeDenom := fee.GetDenomByIndex(0)
	hasHostChainConfig := fadfd.feeabsKeeper.HasHostZoneConfig(ctx, feeDenom)
	if !hasHostChainConfig {
		return fadfd.normalDeductFeeAnteHandle(ctx, tx, simulate, next, feeTx)
	}

	hostChainConfig, _ := fadfd.feeabsKeeper.GetHostZoneConfig(ctx, feeDenom)
	return fadfd.abstractionDeductFeeHandler(ctx, tx, simulate, next, feeTx, hostChainConfig)
}

func (fadfd FeeAbstractionDeductFeeDecorate) normalDeductFeeAnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler, feeTx sdk.FeeTx) (newCtx sdk.Context, err error) {
	if addr := fadfd.accountKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if fadfd.feegrantKeeper == nil {
			return ctx, sdkerrors.Wrap(errorstypes.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := fadfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())
			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := fadfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(errorstypes.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		err = DeductFees(fadfd.bankKeeper, ctx, deductFeesFrom, feeTx.GetFee())
		if err != nil {
			return ctx, err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

func (fadfd FeeAbstractionDeductFeeDecorate) abstractionDeductFeeHandler(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler, feeTx sdk.FeeTx, hostChainConfig feeabstypes.HostChainFeeAbsConfig) (newCtx sdk.Context, err error) {
	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	feeAbstractionPayer := feePayer
	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if fadfd.feegrantKeeper == nil {
			return ctx, sdkerrors.Wrap(errorstypes.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := fadfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())
			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		feeAbstractionPayer = feeGranter
	}

	deductFeesFrom := fadfd.feeabsKeeper.GetFeeAbsModuleAddress()
	deductFeesFromAcc := fadfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(errorstypes.ErrUnknownAddress, "fee abstraction didn't set : %s does not exist", deductFeesFrom)
	}

	// calculate the native token can be swapped from ibc token
	ibcFees := feeTx.GetFee()
	if len(ibcFees) != 1 {
		return ctx, sdkerrors.Wrapf(errorstypes.ErrInvalidCoins, "invalid ibc token: %s", ibcFees)
	}

	nativeFees, err := fadfd.feeabsKeeper.CalculateNativeFromIBCCoins(ctx, ibcFees, hostChainConfig)
	if err != nil {
		return ctx, err
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		err = fadfd.bankKeeper.SendCoinsFromAccountToModule(ctx, feeAbstractionPayer, feeabstypes.ModuleName, ibcFees)
		if err != nil {
			return ctx, err
		}

		err = DeductFees(fadfd.bankKeeper, ctx, deductFeesFrom, nativeFees)
		if err != nil {
			return ctx, err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, accAddress sdk.AccAddress, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, accAddress, types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(errorstypes.ErrInsufficientFunds, err.Error())
	}

	return nil
}

// MempoolFeeDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeDecorator
type FeeAbstrationMempoolFeeDecorator struct {
	feeabsKeeper feeabskeeper.Keeper
}

func NewFeeAbstrationMempoolFeeDecorator(feeabsKeeper feeabskeeper.Keeper) FeeAbstrationMempoolFeeDecorator {
	return FeeAbstrationMempoolFeeDecorator{
		feeabsKeeper: feeabsKeeper,
	}
}

func (famfd FeeAbstrationMempoolFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(errorstypes.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Do not check minimum-gas-prices and global fees during simulations
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Check if this is bypass msg or bypass but not exceed gas useage
	var byPass, byPassNotExceedMaxGasUsage bool
	goCtx := ctx.Context()
	bp := goCtx.Value(feeabstypes.ByPassMsgKey{})
	bpnemgu := goCtx.Value(feeabstypes.ByPassNotExceedMaxGasUsageKey{})
	if bp != nil {
		if bpb, ok := bp.(bool); ok {
			byPass = bpb
		}
	}
	if bpnemgu != nil {
		if bpnemgub, ok := bpnemgu.(bool); ok {
			byPassNotExceedMaxGasUsage = bpnemgub
		}
	}
	if byPass {
		return next(ctx, tx, simulate)
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() {
		feeRequired := GetTxFeeRequired(ctx, int64(gas))
		if feeRequired.IsZero() {
			return next(ctx, tx, simulate)
		}

		// split feeRequired into zero and non-zero coins(nonZeroCoinFeesReq, zeroCoinFeesDenomReq), split feeCoins according to
		// nonZeroCoinFeesReq, zeroCoinFeesDenomReq,
		// so that feeCoins can be checked separately against them.
		nonZeroCoinFeesReq, zeroCoinFeesDenomReq := getNonZeroFees(feeRequired)

		// feeCoinsNonZeroDenom contains non-zero denominations from the feeRequired
		// feeCoinsNonZeroDenom is used to check if the fees meets the requirement imposed by nonZeroCoinFeesReq
		// when feeCoins does not contain zero coins' denoms in feeRequired
		_, feeCoinsZeroDenom := splitCoinsByDenoms(feeCoins, zeroCoinFeesDenomReq)

		feeCoinsLen := feeCoins.Len()
		// if the msg does not satisfy bypass condition and the feeCoins denoms are subset of fezeRequired,
		// check the feeCoins amount against feeRequired
		//
		// when feeCoins=[]
		// special case: and there is zero coin in fee requirement, pass,
		// otherwise, err
		if feeCoinsLen == 0 {
			if len(zeroCoinFeesDenomReq) != 0 {
				return next(ctx, tx, simulate)
			}
			return ctx, sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, feeRequired)
		}

		// when feeCoins != []
		// special case: if TX has at least one of the zeroCoinFeesDenomReq, then it should pass
		if len(feeCoinsZeroDenom) > 0 {
			return next(ctx, tx, simulate)
		}

		// Check if feeDenom is defined in feeabs
		// If so, replace the amount of feeDenom in feeCoins with the
		// corresponding amount of native denom that allow to pay fee
		// TODO: Support more fee token in feeRequired for fee-abstraction
		feeDenom := feeCoins.GetDenomByIndex(0)
		hasHostChainConfig := famfd.feeabsKeeper.HasHostZoneConfig(ctx, feeDenom)
		if hasHostChainConfig && feeCoinsLen == 1 {
			hostChainConfig, _ := famfd.feeabsKeeper.GetHostZoneConfig(ctx, feeDenom)
			nativeCoinsFees, err := famfd.feeabsKeeper.CalculateNativeFromIBCCoins(ctx, feeCoins, hostChainConfig)
			if err != nil {
				return ctx, sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "insufficient fees")

			}
			feeCoins = nativeCoinsFees
		}

		// After all the checks, the tx is confirmed:
		// feeCoins denoms subset off feeRequired (or replaced with fee-abstraction)
		// Not bypass
		// feeCoins != []
		// Not contain zeroCoinFeesDenomReq's denoms
		//
		// check if the feeCoins has coins' amount higher/equal to nonZeroCoinFeesReq
		if !feeCoins.IsAnyGTE(nonZeroCoinFeesReq) {
			err := sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, feeRequired)
			if byPassNotExceedMaxGasUsage {
				err = sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "Insufficient fees; bypass-min-fee-msg-types with gas consumption exceeds the maximum allowed gas value.")
			}
			return ctx, err
		}

	}

	return next(ctx, tx, simulate)
}

// GetTxFeeRequired returns the required fees for the given FeeTx.
func GetTxFeeRequired(ctx sdk.Context, gasLimit int64) sdk.Coins {
	minGasPrices := ctx.MinGasPrices()
	// special case: if minGasPrices=[], requiredFees=[]
	if minGasPrices.IsZero() {
		return sdk.Coins{}
	}

	requiredFees := make(sdk.Coins, len(minGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(gasLimit)
	for i, gp := range minGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredFees.Sort()
}
