package prober

import (
	"bytes"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type scrapeLogger struct {
	next         log.Logger
	buffer       bytes.Buffer
	bufferLogger log.Logger
	logLevel     level.Option
}

func newScrapeLogger(logger log.Logger, module string, logLevel level.Option) *scrapeLogger {
	logger = log.With(logger, "module", module)
	sl := &scrapeLogger{
		next:     logger,
		buffer:   bytes.Buffer{},
		logLevel: logLevel,
	}
	bl := log.NewLogfmtLogger(&sl.buffer)
	sl.bufferLogger = log.With(bl, "ts", log.DefaultTimestampUTC, "caller", log.Caller(6), "module", module)
	return sl
}

func (sl scrapeLogger) Log(keyvals ...interface{}) error {
	sl.bufferLogger.Log(keyvals...)

	return level.NewFilter(sl.next, sl.logLevel).Log(keyvals...)
}
