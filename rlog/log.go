package rlog

import "log"

type ILogger interface {
	Trace(args ...interface{})
	Tracef(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type SLogger struct {
}

func (p *SLogger) Trace(args ...interface{})                 { log.Println(args...) }
func (p *SLogger) Tracef(format string, args ...interface{}) { log.Printf(format, args...) }

func (p *SLogger) Debug(args ...interface{})                 { log.Println(args...) }
func (p *SLogger) Debugf(format string, args ...interface{}) { log.Printf(format, args...) }

func (p *SLogger) Info(args ...interface{})                 { log.Println(args...) }
func (p *SLogger) Infof(format string, args ...interface{}) { log.Printf(format, args...) }

func (p *SLogger) Warn(args ...interface{})                 { log.Println(args...) }
func (p *SLogger) Warnf(format string, args ...interface{}) { log.Printf(format, args...) }

func (p *SLogger) Error(args ...interface{})                 { log.Println(args...) }
func (p *SLogger) Errorf(format string, args ...interface{}) { log.Printf(format, args...) }

func (p *SLogger) Panic(args ...interface{})                 { log.Panicln(args...) }
func (p *SLogger) Panicf(format string, args ...interface{}) { log.Panicf(format, args...) }

func (p *SLogger) Fatal(args ...interface{})                 { log.Fatalln(args...) }
func (p *SLogger) Fatalf(format string, args ...interface{}) { log.Fatalf(format, args...) }

var Logger ILogger

func Trace(args ...interface{})                 { Logger.Trace(args...) }
func Tracef(format string, args ...interface{}) { Logger.Tracef(format, args...) }

func Debug(args ...interface{})                 { Logger.Debug(args...) }
func Debugf(format string, args ...interface{}) { Logger.Debugf(format, args...) }

func Info(args ...interface{})                 { Logger.Info(args...) }
func Infof(format string, args ...interface{}) { Logger.Infof(format, args...) }

func Warn(args ...interface{})                 { Logger.Warn(args...) }
func Warnf(format string, args ...interface{}) { Logger.Warnf(format, args...) }

func Error(args ...interface{})                 { Logger.Error(args...) }
func Errorf(format string, args ...interface{}) { Logger.Errorf(format, args...) }

func Panic(args ...interface{})                 { Logger.Panic(args...) }
func Panicf(format string, args ...interface{}) { Logger.Panicf(format, args...) }

func Fatal(args ...interface{})                 { Logger.Fatal(args...) }
func Fatalf(format string, args ...interface{}) { Logger.Fatalf(format, args...) }

func init() {
	Logger = &SLogger{}
}
