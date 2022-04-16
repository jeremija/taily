package guardlog

import "github.com/juju/errors"

type Processors []Processor

func (p Processors) ProcessMessage(message Message) error {
	for _, proc := range p {
		if err := proc.ProcessMessage(message); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
