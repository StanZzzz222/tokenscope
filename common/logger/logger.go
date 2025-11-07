package logger

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"io"
	"os"
	"time"
)

/*
   Created by zyx
   Date Time: 2025/9/11
   File: logger.go
*/

var logger *TokenScopeLogger

type TokenScopeLogger struct {
	*log.Logger
}

type TeeWriter struct {
	file io.Writer
}

func InitLogger() {
	if logger == nil {
		opts := log.NewWithOptions(os.Stdout, log.Options{
			ReportCaller:    false,
			ReportTimestamp: true,
			TimeFormat:      "[2006-01-02 15:04:05]",
			Prefix:          "",
		})
		outFileName := viper.GetString("logger.name")
		enalbedOutFile := viper.GetBool("logger.file-enabled")
		if enalbedOutFile {
			logFile, err := os.OpenFile(outFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
			if err != nil {
				panic("open log file falied: " + err.Error())
			}
			writer := &TeeWriter{
				file: logFile,
			}
			opts.SetOutput(writer)
		}
		logger = &TokenScopeLogger{opts}
	}
}

func Logger() *TokenScopeLogger {
	return logger
}

func TimeTrack(name string, callback func()) {
	start := time.Now()
	callback()
	elapsed := time.Since(start)
	time.Sleep(time.Millisecond * 300)
	logger.Infof("%v loaded, took %d ms", name, elapsed.Milliseconds())
}

func (tw *TeeWriter) Write(p []byte) (int, error) {
	n, err := os.Stdout.Write(p)
	if err != nil {
		return n, err
	}
	_, _ = tw.file.Write(p)
	return n, err
}

func (t *TokenScopeLogger) Errorf(format string, args ...any) {
	t.Logger.Errorf(format, args...)
}
