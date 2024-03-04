package ante

import (
	"errors"
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

func NewFeeAbstractionDeductFeeDecorate(
	ak AccountKeeper,
	bk BankKeeper,
	feeabsKeeper feeabskeeper.Keeper,
	fk FeegrantKeeper,
) FeeAbstractionDeductFeeDecorate {
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

	if addr := fadfd.accountKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	fee := feeTx.GetFee()
	if len(fee) == 0 {
		return fadfd.normalDeductFeeAnteHandle(ctx, tx, simulate, next, feeTx)
	}

	feeDenom := fee.GetDenomByIndex(0)
	hostChainConfig, found := fadfd.feeabsKeeper.GetHostZoneConfig(ctx, feeDenom)
	if !found {
		return fadfd.normalDeductFeeAnteHandle(ctx, tx, simulate, next, feeTx)
	}

	return fadfd.abstractionDeductFeeHandler(ctx, tx, simulate, next, feeTx, hostChainConfig)
}

// normalDeductFeeAnteHandle deducts the fee from fee payer or fee granter (if set) and ensure
// the fee collector module account is set
func (fadfd FeeAbstractionDeductFeeDecorate) normalDeductFeeAnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
	feeTx sdk.FeeTx,
) (newCtx sdk.Context, err error) {
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
	if !fee.IsZero() {
		err = DeductFees(fadfd.bankKeeper, ctx, deductFeesFrom, fee)
		if err != nil {
			return ctx, err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// abstractionDeductFeeHandler calculates the equivalent native tokens from
// IBC tokens and deducts the fees accordingly if the transaction involves IBC tokens
// and the host chain configuration is set.
func (fadfd FeeAbstractionDeductFeeDecorate) abstractionDeductFeeHandler(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler, feeTx sdk.FeeTx, hostChainConfig feeabstypes.HostChainFeeAbsConfig) (newCtx sdk.Context, err error) {
	if hostChainConfig.Frozen {
		return ctx, sdkerrors.Wrap(feeabstypes.ErrHostZoneFrozen, "cannot deduct fee as host zone is frozen")
	}
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
	if len(fee) != 1 {
		return ctx, sdkerrors.Wrapf(errorstypes.ErrInvalidCoins, "invalid ibc token: %s", fee)
	}

	nativeFees, err := fadfd.feeabsKeeper.CalculateNativeFromIBCCoins(ctx, fee, hostChainConfig)
	if err != nil {
		return ctx, err
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		err = fadfd.bankKeeper.SendCoinsFromAccountToModule(ctx, feeAbstractionPayer, feeabstypes.ModuleName, fee)
		if err != nil {
			return ctx, err
		}

		err = DeductFees(fadfd.bankKeeper, ctx, deductFeesFrom, nativeFees)
		if err != nil {
			return ctx, err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, accAddress sdk.AccAddress, fees sdk.Coins) error {
	if err := fees.Validate(); err != nil {
		return sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	if err := bankKeeper.SendCoinsFromAccountToModule(ctx, accAddress, types.FeeCollectorName, fees); err != nil {
		return sdkerrors.Wrapf(errorstypes.ErrInsufficientFunds, err.Error())
	}

	return nil
}

// FeeAbstrationMempoolFeeDecorator will check if the transaction's fee is at least as large
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

	// Check if this is bypass msg or bypass but not exceed gas usage
	var byPass, byPassExceedMaxGasUsage, isGlobalFee bool
	goCtx := ctx.Context()
	bp := goCtx.Value(feeabstypes.ByPassMsgKey{})
	bpemgu := goCtx.Value(feeabstypes.ByPassExceedMaxGasUsageKey{})
	iglbf := goCtx.Value(feeabstypes.GlobalFeeKey{})
	if bp != nil {
		if bpb, ok := bp.(bool); ok {
			byPass = bpb
		}
	}
	if bpemgu != nil {
		if bpemgub, ok := bpemgu.(bool); ok {
			byPassExceedMaxGasUsage = bpemgub
		}
	}
	if iglbf != nil {
		if iglbfb, ok := iglbf.(bool); ok {
			isGlobalFee = iglbfb
		}
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() || isGlobalFee {
		feeRequired, err := famfd.GetTxFeeRequired(ctx, int64(gas))
		if err != nil {
			return ctx, err
		}

		// split feeRequired into zero and non-zero coins(nonZeroCoinFeesReq, zeroCoinFeesDenomReq)
		// split feeCoins according to nonZeroCoinFeesReq, zeroCoinFeesDenomReq,
		// so that feeCoins can be checked separately against them.
		nonZeroCoinFeesReq, zeroCoinFeesDenomReq := getNonZeroFees(feeRequired)

		// feeCoinsNonZeroDenom contains non-zero denominations from the feeRequired
		// feeCoinsNonZeroDenom is used to check if the fees meets the requirement imposed by nonZeroCoinFeesReq
		// when feeCoins does not contain zero coins' denoms in feeRequired
		feeCoinsNonZeroDenom, feeCoinsZeroDenom := splitCoinsByDenoms(feeCoins, zeroCoinFeesDenomReq)

		feeCoinsLen := feeCoins.Len()

		// Check if feeDenom is defined in feeabs
		// If so, replace the amount of feeDenom in feeCoins with the
		// corresponding amount of native denom that allow to pay fee
		// TODO: Support more fee token in feeRequired for fee-abstraction
		if feeCoinsNonZeroDenom.Len() == 1 {
			feeDenom := feeCoinsNonZeroDenom.GetDenomByIndex(0)
			hostChainConfig, found := famfd.feeabsKeeper.GetHostZoneConfig(ctx, feeDenom)
			if found {
				if hostChainConfig.Frozen {
					return ctx, sdkerrors.Wrapf(feeabstypes.ErrHostZoneFrozen, "cannot deduct fee as host zone is frozen")
				}
				nativeCoinsFees, err := famfd.feeabsKeeper.CalculateNativeFromIBCCoins(ctx, feeCoinsNonZeroDenom, hostChainConfig)
				if err != nil {
					return ctx, sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "insufficient fees")
				}
				feeCoinsNonZeroDenom = nativeCoinsFees
			}
		}

		// After replace the feeCoinsNonZeroDenom, feeCoinsNonZeroDenom must be in denom subset of nonZeroCoinFeesReq
		if !feeCoinsNonZeroDenom.DenomsSubsetOf(nonZeroCoinFeesReq) {
			return ctx, sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "fee is not a subset of required fees; got %s, required: %s", feeCoins.String(), feeRequired.String())
		}

		// if the msg does not satisfy bypass condition and the feeCoins denoms are subset of fezeRequired,
		// check the feeCoins amount against feeRequired
		//
		// when feeCoins=[]
		// special case: and there is zero coin in fee requirement, pass, otherwise, err
		// when feeCoins != []
		// special case: if TX has at least one of the zeroCoinFeesDenomReq, then it should pass
		if byPass || (feeCoinsLen == 0 && len(zeroCoinFeesDenomReq) != 0) || len(feeCoinsZeroDenom) > 0 {
			return next(ctx, tx, simulate)
		}

		if feeCoinsLen == 0 {
			return ctx, sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, feeRequired)
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
			if byPassExceedMaxGasUsage {
				err = sdkerrors.Wrapf(errorstypes.ErrInsufficientFee, "Insufficient fees; bypass-min-fee-msg-types with gas consumption exceeds the maximum allowed gas value.")
			}
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (famfd FeeAbstrationMempoolFeeDecorator) DefaultZeroFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom := famfd.feeabsKeeper.GetDefaultBondDenom(ctx)
	if bondDenom == "" {
		return nil, errors.New("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdk.NewDec(0))}, nil
}

// GetTxFeeRequired returns the required fees for the given FeeTx.
func (famfd FeeAbstrationMempoolFeeDecorator) GetTxFeeRequired(ctx sdk.Context, gasLimit int64) (sdk.Coins, error) {
	var (
		minGasPrices sdk.DecCoins
		err          error
	)

	minGasPrices = ctx.MinGasPrices()
	// if min_gas_prices is empty set, set to 0(bond_denom)
	if len(minGasPrices) == 0 {
		minGasPrices, err = famfd.DefaultZeroFee(ctx)
		if err != nil {
			return sdk.Coins{}, err
		}
	}

	requiredFees := make(sdk.Coins, len(minGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(gasLimit)
	for i, gp := range minGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredFees.Sort(), nil
}
