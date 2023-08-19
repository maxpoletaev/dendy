package loglevel

import (
	"bytes"
	"io"
)

type Level uint8

const (
	LevelNone Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

var levels = map[string]Level{
	"DEBUG": LevelDebug,
	"INFO":  LevelInfo,
	"WARN":  LevelWarn,
	"ERROR": LevelError,
}

func extractLevel(msg []byte) Level {
	if len(msg) >= 2 && msg[0] == '[' {
		i := bytes.IndexByte(msg, ']')

		if i > 0 {
			l := string(msg[1:i])
			return levels[l]
		}
	}

	return LevelError
}

type LevelFilter struct {
	Output io.Writer
	Level  Level
}

func (f *LevelFilter) Write(p []byte) (n int, err error) {
	if f.Level == LevelNone {
		return len(p), nil
	}

	if extractLevel(p) >= f.Level {
		return f.Output.Write(p)
	}

	return len(p), nil
}
