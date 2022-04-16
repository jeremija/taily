package guardlog

type Processor interface {
	ProcessMessage(Message) error
}
