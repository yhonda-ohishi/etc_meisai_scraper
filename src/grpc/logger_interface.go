package grpc

// LoggerInterface defines the interface for logging
type LoggerInterface interface {
	// Printf formats according to a format specifier and writes to the logger
	Printf(format string, v ...interface{})

	// Println writes to the logger with a newline
	Println(v ...interface{})

	// Print writes to the logger
	Print(v ...interface{})

	// Fatalf is equivalent to Printf() followed by a call to os.Exit(1)
	Fatalf(format string, v ...interface{})

	// Fatal is equivalent to Print() followed by a call to os.Exit(1)
	Fatal(v ...interface{})

	// Panicf is equivalent to Printf() followed by a call to panic()
	Panicf(format string, v ...interface{})

	// Panic is equivalent to Print() followed by a call to panic()
	Panic(v ...interface{})
}