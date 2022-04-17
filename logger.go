package guardlog

import "github.com/peer-calls/log"

// LoggerWithReaderID adds the reader_id meta to the logger.
func LoggerWithReaderID(logger log.Logger, readerID ReaderID) log.Logger {
	return logger.WithCtx(log.Ctx{
		"reader_id": readerID,
	})
}
