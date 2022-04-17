package processor

import "github.com/jeremija/taily/types"

type Factory func() (types.Processor, error)
