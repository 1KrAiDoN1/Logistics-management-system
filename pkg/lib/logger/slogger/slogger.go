package slogger

import (
	"context"
	"encoding/json"
	"io"
	stdLog "log"
	"log/slog"
	"time"

	"github.com/fatih/color"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "err",
		Value: slog.StringValue(err.Error()),
	}
}

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
	TimeZone *time.Location
}

type PrettyHandler struct {
	slog.Handler
	l        *stdLog.Logger
	attrs    []slog.Attr
	timeZone *time.Location
}

func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	timezone := opts.TimeZone
	if timezone == nil {
		timezone = time.Local
	}
	h := &PrettyHandler{
		Handler:  slog.NewJSONHandler(out, opts.SlogOpts),
		l:        stdLog.New(out, "", 0),
		timeZone: timezone,
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

	fields := make(map[string]interface{}, r.NumAttrs())

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

	localTime := r.Time.In(h.timeZone)
	timeStr := localTime.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	)

	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
	}
}
