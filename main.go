package tflogsloghandler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TFLogSlogHandler struct {
	fields map[string]any
	name   string
	mutex  sync.Mutex
}

var _ slog.Handler = (*TFLogSlogHandler)(nil)

func NewSlogHandler() *TFLogSlogHandler {
	return &TFLogSlogHandler{}
}

func (*TFLogSlogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *TFLogSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	switch record.Level {
	case slog.LevelDebug:
		tflog.Debug(ctx, record.Message, h.renderAdditionalFields(record))
	case slog.LevelInfo:
		tflog.Info(ctx, record.Message, h.renderAdditionalFields(record))
	case slog.LevelWarn:
		tflog.Warn(ctx, record.Message, h.renderAdditionalFields(record))
	case slog.LevelError:
		tflog.Error(ctx, record.Message, h.renderAdditionalFields(record))
	default:
		tflog.Info(ctx, record.Message, h.renderAdditionalFields(record))
	}
	return nil
}

func (h *TFLogSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	addAttrsToMap(attrs, h.fields)
	return h
}

func (h *TFLogSlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := TFLogSlogHandler{
		name: name,
	}
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h2.fields = make(map[string]any, len(h.fields)+1)
	for k, v := range h.fields {
		h2.fields[k] = v
	}
	return &h2
}

func (h *TFLogSlogHandler) renderAdditionalFields(record slog.Record) map[string]any {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	result := make(map[string]any, len(h.fields)+record.NumAttrs())
	for k, v := range h.fields {
		result[k] = v
	}
	record.Attrs(func(attr slog.Attr) bool {
		addAttrToMap(attr, result)
		return true
	})
	return result
}

func addAttrsToMap(attrs []slog.Attr, fields map[string]any) {
	for _, a := range attrs {
		addAttrToMap(a, fields)
	}
}

func addAttrToMap(attr slog.Attr, fields map[string]any) {
	if attr.Equal(slog.Attr{}) {
		return
	}
	val := attr.Value.Resolve()
	if val.Kind() == slog.KindGroup {
		attrs := val.Group()
		if len(attrs) == 0 {
			return
		}
		if attr.Key == "" {
			addAttrsToMap(attrs, fields)
			return
		}
		group := make(map[string]any, len(attrs))
		addAttrsToMap(attrs, group)
		fields[attr.Key] = group
		return
	}
	fields[attr.Key] = val.Any()
}
