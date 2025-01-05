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

var levelNames = [5][]byte{
	nil,
	[]byte("DEBUG"),
	[]byte("INFO"),
	[]byte("WARN"),
	[]byte("ERROR"),
}

func extractLevel(msg []byte) Level {
	if len(msg) >= 2 && msg[0] == '[' {
		i := bytes.IndexByte(msg, ']')
		s := msg[1:i]

		if i != -1 {
			switch {
			case bytes.Equal(s, levelNames[LevelDebug]):
				return LevelDebug
			case bytes.Equal(s, levelNames[LevelInfo]):
				return LevelInfo
			case bytes.Equal(s, levelNames[LevelWarn]):
				return LevelWarn
			case bytes.Equal(s, levelNames[LevelError]):
				return LevelError
			}
		}
	}

	return LevelError
}

type LevelFilter struct {
	output io.Writer
	level  Level
}

func New(output io.Writer, level Level) *LevelFilter {
	return &LevelFilter{
		output: output,
		level:  level,
	}
}

func (f *LevelFilter) Write(p []byte) (n int, err error) {
	if f.level == LevelNone {
		return len(p), nil
	}

	if extractLevel(p) >= f.level {
		return f.output.Write(p)
	}

	return len(p), nil
}
