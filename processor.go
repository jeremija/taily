package guardlog

// Processor is a message processor.
type Processor interface {
	// ProcessMessage processes the message read.
	ProcessMessage(Message) error
}
