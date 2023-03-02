# State 

## OsmosisTwapExchangeRate

The exchange rate of an ibc denom to Osmosis: `0x01<ibc_denom_bytes> -> sdk.Dec`

When we send the QueryArithmeticTwapToNowRequest to the Osmosis contract via IBC, the contract will send an acknowledgement with price data to the fee abstraction chain. The OsmosisTwapExchangeRate will then be updated based on this value.
This exchange rate is then used to calculate transaction fees in the appropriate IBC denom. By updating the exchange rate based on the most recent price data, we can ensure that transaction fees accurately reflect the current market conditions on Osmosis.

It's important to note that the exchange rate will fluctuate over time, as it is based on the time-weighted average price (TWAP) of the IBC denom on Osmosis. This means that the exchange rate will reflect the average price of the IBC denom over a certain time period, rather than an instantaneous price.


## Modified Fee Deduct Antehandler

When making a transaction, usually users need to pay fees in the native token, but "fee abstraction" allows them to pay fees in other tokens.

To allow for this, we use modified versions of `MempoolFeeDecorator` and `DeductFeeDecorate`. In these ante handlers, IBC tokens are swapped to the native token before the next fee handler logic is executed.

If a blockchain uses the Fee Abstraction module, it is necessary to replace the MempoolFeeDecorator and `DeductFeeDecorate` with the `FeeAbstrationMempoolFeeDecorator` and `FeeAbstractionDeductFeeDecorate`, respectively.


Example :

```
anteDecorators := []sdk.AnteDecorator{
  ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
  ante.NewRejectExtensionOptionsDecorator(),
  feeabsante.NewFeeAbstrationMempoolFeeDecorator(options.FeeAbskeeper),
  ante.NewValidateBasicDecorator(),
  ante.NewTxTimeoutHeightDecorator(),
  ante.NewValidateMemoDecorator(options.AccountKeeper),
  ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
  feeabsante.NewFeeAbstractionDeductFeeDecorate(options.AccountKeeper, options.BankKeeper, options.FeeAbskeeper, options.FeegrantKeeper),
  // SetPubKeyDecorator must be called before all signature verification decorators
  ante.NewSetPubKeyDecorator(options.AccountKeeper),
  ante.NewValidateSigCountDecorator(options.AccountKeeper),
  ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
  ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
  ante.NewIncrementSequenceDecorator(options.AccountKeeper),
  ibcante.NewAnteDecorator(options.IBCKeeper),
 }


