package types

const (
	EventTypePacket     = "receive_feechain_verification_packet"
	EventTypeEpochEnd   = "epoch_end"		// TODO: need to clean up (not use)
	EventTypeEpochStart = "epoch_start"

	AttributeKeyAckSuccess  = "success"
	AttributeKeyClientID    = "client_id"
	AttributeKeyAck         = "acknowledgement"
	AttributeKeyAckError    = "ack_error"
	AttributeEpochNumber    = "epoch_number"
	AttributeEpochStartTime = "start_time"
)
