package guardlog

import "github.com/peer-calls/log"

func LoggerWithReaderID(logger log.Logger, readerID ReaderID) log.Logger {
	return logger.WithCtx(log.Ctx{
		"reader_id": readerID,
	})
}
