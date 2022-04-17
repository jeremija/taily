package guardlog

import "github.com/peer-calls/log"

func LoggerWithWatcherID(logger log.Logger, watcherID WatcherID) log.Logger {
	return logger.WithCtx(log.Ctx{
		"watcher_id": watcherID,
	})
}
