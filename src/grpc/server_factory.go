package grpc

import (
	"log"

	"github.com/yhonda-ohishi/etc_meisai/src/services"
)

// defaultLogger wraps log.Logger to implement LoggerInterface
type defaultLogger struct {
	logger *log.Logger
}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

func (l *defaultLogger) Println(v ...interface{}) {
	l.logger.Println(v...)
}

func (l *defaultLogger) Print(v ...interface{}) {
	l.logger.Print(v...)
}

func (l *defaultLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}

func (l *defaultLogger) Fatal(v ...interface{}) {
	l.logger.Fatal(v...)
}

func (l *defaultLogger) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}

func (l *defaultLogger) Panic(v ...interface{}) {
	l.logger.Panic(v...)
}

// NewETCMeisaiServerWithConcreteServices creates a server with concrete service implementations
// This is a factory function for backward compatibility
func NewETCMeisaiServerWithConcreteServices(
	etcMeisaiService *services.ETCMeisaiService,
	etcMappingService *services.ETCMappingService,
	importService *services.ImportService,
	statisticsService *services.StatisticsService,
	logger *log.Logger,
) *ETCMeisaiServer {
	// Convert concrete logger to interface
	var loggerInterface LoggerInterface
	if logger != nil {
		loggerInterface = &defaultLogger{logger: logger}
	}

	// Cast concrete services to interfaces (Go's implicit interface implementation)
	return NewETCMeisaiServer(
		etcMeisaiService,  // implements ETCMeisaiServiceInterface
		etcMappingService, // implements ETCMappingServiceInterface
		importService,     // implements ImportServiceInterface
		statisticsService, // implements StatisticsServiceInterface
		loggerInterface,
	)
}