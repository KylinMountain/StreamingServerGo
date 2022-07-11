package rtmp

const (
	MessageTypeChunkSize                = 1
	MessageTypeAbort                    = 2
	MessageTypeAcknowledge              = 3
	MessageTypeAcknowledgeWindowSize    = 5
	MessageTypeSetAcknowledgeWindowSize = 6

	MessageTypeUserControl = 4

	MessageTypeDataAMF0         = 18
	MessageTypeDataAMF3         = 15
	MessageTypeSharedObjectAMF0 = 19
	MessageTypeSharedObjectAMF3 = 16

	MessageTypeAudio     = 8
	MessageTypeVideo     = 9
	MessageTypeAggregate = 22

	MessageTypeCommandAMF0 = 20
	MessageTypeCommandAMF3 = 17
)
