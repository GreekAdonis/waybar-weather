// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func NewLogger(level slog.Level) *Logger {
	output := os.Stderr
	return &Logger{slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{Level: level}))}
}

func Err(err error) slog.Attr {
	return slog.Any("error", err)
}
