package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
)

func Setup() {
	w := os.Stderr

	slog.SetDefault(slog.New(tint.NewHandler(colorable.NewColorable(w), &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.DateTime,
	})))
}
