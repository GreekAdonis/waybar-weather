package main

import (
	"log/slog"
	"os"
)

type logger struct {
	*slog.Logger
}

func newLogger() *logger {
	var output = os.Stderr
	log := slog.New(
		slog.NewTextHandler(output,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)
	return &logger{log}
}

func logError(err error) slog.Attr {
	return slog.Any("error", err)
}
