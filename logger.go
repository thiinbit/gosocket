package gosocket

import (
	"log"
	"os"
	"sync"
)

type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
}

var (
	DefaultDebugLogger = log.New(os.Stderr, "[Gosocket-Debug]", log.LstdFlags)
	DefaultLogger      = log.New(os.Stderr, "[Gosocket]", log.LstdFlags)
)

type DebugLogger struct {
	isDebugMode bool
	logger      Logger
	mu          sync.Mutex
}

func (d *DebugLogger) SetDebugMode(on bool) {
	d.mu.Lock()
	d.isDebugMode = on
	d.mu.Unlock()
}

func (d *DebugLogger) Print(v ...interface{}) {
	if d.isDebugMode {
		d.logger.Print(v...)
	}
}
func (d *DebugLogger) Printf(format string, v ...interface{}) {
	if d.isDebugMode {
		d.logger.Printf(format, v...)
	}
}
func (d *DebugLogger) Println(v ...interface{}) {
	if d.isDebugMode {
		d.logger.Println(v...)
	}
}
func (d *DebugLogger) Fatal(v ...interface{}) {
	if d.isDebugMode {
		d.logger.Fatal(v...)
	}
}
func (d *DebugLogger) Fatalf(format string, v ...interface{}) {
	if d.isDebugMode {
		d.logger.Fatalf(format, v...)
	}
}
func (d *DebugLogger) Fatalln(v ...interface{}) {
	if d.isDebugMode {
		d.logger.Fatalln(v...)
	}
}
func (d *DebugLogger) Panic(v ...interface{}) {
	if d.isDebugMode {
		d.logger.Panic(v...)
	}
}
func (d *DebugLogger) Panicf(format string, v ...interface{}) {
	if d.isDebugMode {
		d.logger.Panicf(format, v...)
	}
}
func (d *DebugLogger) Panicln(v ...interface{}) {
	if d.isDebugMode {
		d.logger.Panicln(v...)
	}
}
