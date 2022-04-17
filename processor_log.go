package guardlog

import (
	"fmt"
	"os"
	"time"
)

type ProcessorLog struct{}

func NewProcessorLog() *ProcessorLog {
	return &ProcessorLog{}
}

func (p ProcessorLog) ProcessMessage(message Message) error {
	fmt.Fprintf(os.Stdout, "%s %s %s\n", message.Timestamp.Format(time.RFC3339Nano), message.ReaderID, message.Fields)

	return nil
}
