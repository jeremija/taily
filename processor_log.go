package guardlog

import (
	"fmt"
	"os"
)

type ProcessorLog struct{}

func NewProcessorLog() *ProcessorLog {
	return &ProcessorLog{}
}

func (p ProcessorLog) ProcessMessage(message Message) error {
	fmt.Fprintf(os.Stdout, "%s %s %s", message.Timestamp, message.WatcherID, message.Fields["message"])

	return nil
}
