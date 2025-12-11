package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/fatih/color"
)

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	// opts PrettyHandlerOptions
	slog.Handler

	out         io.Writer
	attrs       []slog.Attr
	withContext bool
}

func (opts PrettyHandlerOptions) NewPrettyHandler(
	withContext bool,
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler:     slog.NewJSONHandler(out, opts.SlogOpts),
		out:         out,
		withContext: withContext,
	}

	return h
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]any, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	timeStr := r.Time.Format("[15:04:05.000]")
	msg := color.CyanString(r.Message)

	output := strings.Join([]string{
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	}, " ")

	_, err = fmt.Fprintln(h.out, output)
	return err
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.withContext {
		return &PrettyHandler{
			Handler:     h.Handler,
			out:         h.out,
			attrs:       attrs,
			withContext: h.withContext,
		}
	}
	return &PrettyHandler{
		Handler:     h.Handler,
		out:         h.out,
		withContext: h.withContext,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		Handler:     h.Handler.WithGroup(name),
		out:         h.out,
		withContext: h.withContext,
	}
}
