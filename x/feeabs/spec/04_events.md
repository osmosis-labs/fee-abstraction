# Events

The `feeabs` module emits the following events:

## BeginBlocker

| Type        | Attribute Key | Attribute Value         |
| ----------- | ------------- | ----------------------- |
| epoch_start | epoch_number  | {currentEpoch}          |
| epoch_start | start_time    | {currentEpochStartTime} |

## IBC

| Type                                 | Attribute Key   | Attribute Value |
| ------------------------------------ | --------------- | --------------- |
| receive_feechain_verification_packet | module          | {moduleName}    |
| receive_feechain_verification_packet | acknowledgement | {ack}           |
| receive_feechain_verification_packet | success         | {ack.Result}    |
| receive_feechain_verification_packet | ack_error       | {ack.Error◊}     |
