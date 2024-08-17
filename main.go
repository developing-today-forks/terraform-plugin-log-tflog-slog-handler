package tflogsloghandler

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TFLogSlogHandler struct {
	fields map[string]any
	groups []string
	mutex  sync.Mutex
}

var _ slog.Handler = (*TFLogSlogHandler)(nil)

func NewSlogHandler() *TFLogSlogHandler {
	return &TFLogSlogHandler{
		fields: make(map[string]any),
	}
}

func (*TFLogSlogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *TFLogSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	fields := h.renderAdditionalFields(record)
	subsystem := strings.Join(h.groups, ".")

	if subsystem != "" {
		switch record.Level {
		case slog.LevelDebug:
			tflog.SubsystemDebug(ctx, subsystem, record.Message, fields)
		case slog.LevelInfo:
			tflog.SubsystemInfo(ctx, subsystem, record.Message, fields)
		case slog.LevelWarn:
			tflog.SubsystemWarn(ctx, subsystem, record.Message, fields)
		case slog.LevelError:
			tflog.SubsystemError(ctx, subsystem, record.Message, fields)
		default:
			tflog.SubsystemInfo(ctx, subsystem, record.Message, fields)
		}
	} else {
		switch record.Level {
		case slog.LevelDebug:
			tflog.Debug(ctx, record.Message, fields)
		case slog.LevelInfo:
			tflog.Info(ctx, record.Message, fields)
		case slog.LevelWarn:
			tflog.Warn(ctx, record.Message, fields)
		case slog.LevelError:
			tflog.Error(ctx, record.Message, fields)
		default:
			tflog.Info(ctx, record.Message, fields)
		}
	}
	return nil
}

func (h *TFLogSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	newHandler := &TFLogSlogHandler{
		fields: make(map[string]any),
		groups: append([]string{}, h.groups...),
	}
	for k, v := range h.fields {
		newHandler.fields[k] = v
	}
	addAttrsToMap(attrs, newHandler.fields, newHandler.groups)
	return newHandler
}

func (h *TFLogSlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h.mutex.Lock()
	defer h.mutex.Unlock()
	newHandler := &TFLogSlogHandler{
		fields: make(map[string]any),
		groups: append(append([]string{}, h.groups...), name),
	}
	for k, v := range h.fields {
		newHandler.fields[k] = v
	}
	return newHandler
}

func (h *TFLogSlogHandler) renderAdditionalFields(record slog.Record) map[string]any {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	result := make(map[string]any, len(h.fields)+record.NumAttrs())
	for k, v := range h.fields {
		result[k] = v
	}
	record.Attrs(func(attr slog.Attr) bool {
		addAttrToMap(attr, result, h.groups)
		return true
	})
	return result
}

func addAttrsToMap(attrs []slog.Attr, fields map[string]any, groups []string) {
	for _, a := range attrs {
		addAttrToMap(a, fields, groups)
	}
}

func addAttrToMap(attr slog.Attr, fields map[string]any, groups []string) {
	if attr.Equal(slog.Attr{}) {
		return
	}
	val := attr.Value.Resolve()
	key := attr.Key

	current := fields
	for _, group := range groups {
		if _, ok := current[group]; !ok {
			current[group] = make(map[string]any)
		}
		current = current[group].(map[string]any)
	}

	if val.Kind() == slog.KindGroup {
		attrs := val.Group()
		if len(attrs) == 0 {
			return
		}
		if key == "" {
			addAttrsToMap(attrs, current, nil)
			return
		}
		group := make(map[string]any, len(attrs))
		addAttrsToMap(attrs, group, nil)
		current[key] = group
		return
	}
	current[key] = val.Any()
}
