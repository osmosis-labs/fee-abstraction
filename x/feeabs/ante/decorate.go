package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	feeabskeeper "github.com/notional-labs/feeabstraction/v1/x/feeabs/keeper"
	feeabstypes "github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
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
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
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
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
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
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
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
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := fadfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())

			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		feeAbstractionPayer = feeGranter
	}

	//fee abstraction deduct logic
	deductFeesFrom := fadfd.feeabsKeeper.GetFeeAbsModuleAddress()
	deductFeesFromAcc := fadfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee abstraction didn't set : %s does not exist", deductFeesFrom)
	}

	// calculate the native token can be swapped from ibc token
	ibcFees := feeTx.GetFee()
	if len(ibcFees) != 1 {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid ibc token: %s", ibcFees)
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
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, accAddress, types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
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
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !simulate {
		minGasPrices := ctx.MinGasPrices()
		if minGasPrices.IsZero() {
			return next(ctx, tx, simulate)
		}
		feeCoinsLen := feeCoins.Len()
		if feeCoinsLen == 0 {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees")
		}

		hostChainConfig, err := famfd.feeabsKeeper.GetHostZoneConfig(ctx, feeCoins[0].Denom)
		if err != nil && feeCoinsLen == 1 {
			ibcFees := feeTx.GetFee()
			nativeCoinsFees, err := famfd.feeabsKeeper.CalculateNativeFromIBCCoins(ctx, ibcFees, hostChainConfig)
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees")

			}
			feeCoins = nativeCoinsFees
		}

		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}

	}

	return next(ctx, tx, simulate)
}
