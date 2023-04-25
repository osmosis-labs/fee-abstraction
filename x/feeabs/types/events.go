package types

const (
	EventTypePacket     = "receive_feechain_verification_packet"
	EventTypeEpochStart = "epoch_start"

	AttributeKeyAckSuccess  = "success"
	AttributeKeyAck         = "acknowledgement"
	AttributeKeyAckError    = "ack_error"
	AttributeEpochNumber    = "epoch_number"
	AttributeEpochStartTime = "start_time"
)
