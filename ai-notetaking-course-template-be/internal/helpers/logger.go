package helpers

import (
	"ai-notetaking-be/internal/loggers"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func SetupLogger() {
	log := logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	log.Info("Logger initiated using logrus")
	Logger = log
}

func LogExecution(
	logger loggers.Logger,
	start time.Time,
	err *error,
	memBefore uint64,
	message string,
	fields map[string]interface{},
) {
	duration := time.Since(start).Milliseconds()

	if fields == nil {
		fields = map[string]interface{}{}
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsed := m.Alloc - memBefore
	fields["mem_before"] = memBefore
	fields["mem_after"] = m.Alloc
	fields["mem_used"] = memUsed
	fields["duration"] = duration
	fields["duration_ms"] = duration
	fields["memory_bytes"] = memUsed
	fields["memory_kb"] = memUsed / 1024
	fields["memory_mb"] = memUsed / (1024 * 1024)

	if *err != nil {
		fields["error"] = (*err).Error()
		logger.Warn(message+" FAILED", fields)
	} else {
		if duration > 200 {
			logger.Warn(message+" SLOW", fields)
			return
		}
		logger.Info(message+" SUCCESS", fields)
	}
}
