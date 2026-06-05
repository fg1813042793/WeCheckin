package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *log.Logger

func Init(logDir, level string, maxAge int, compress bool) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	accessLog := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, fmt.Sprintf("access-%s.log", time.Now().Format("2006-01-02"))),
		MaxAge:     maxAge,
		Compress:   compress,
		LocalTime:  true,
	}

	errorLog := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, fmt.Sprintf("error-%s.log", time.Now().Format("2006-01-02"))),
		MaxAge:     maxAge,
		Compress:   compress,
		LocalTime:  true,
	}

	Logger = log.New(io.MultiWriter(os.Stdout, accessLog), "", log.LstdFlags|log.Lshortfile)
	log.SetOutput(io.MultiWriter(os.Stdout, errorLog))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	go rotateDaily(logDir, accessLog, errorLog)

	return nil
}

func rotateDaily(logDir string, accessLog, errorLog *lumberjack.Logger) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		time.Sleep(time.Until(next))
		accessLog.Filename = filepath.Join(logDir, fmt.Sprintf("access-%s.log", time.Now().Format("2006-01-02")))
		errorLog.Filename = filepath.Join(logDir, fmt.Sprintf("error-%s.log", time.Now().Format("2006-01-02")))
		accessLog.Rotate()
		errorLog.Rotate()
	}
}
