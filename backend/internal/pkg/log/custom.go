package log

import (
	"context"
	"log/slog"

	"vpainless/internal/pkg/authz"
	"vpainless/pkg/middleware"
)

type CustomHandler struct {
	Handler slog.Handler
}

func (c *CustomHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return c.Handler.Enabled(ctx, lvl)
}

func (c *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	var attrs []slog.Attr
	rid, err := middleware.GetRequestID(ctx)
	if err == nil {
		attrs = append(attrs, slog.String("request_id", rid.String()))
	}

	principal, err := authz.GetPrincipal(ctx)
	if err == nil {
		attrs = append(attrs, slog.Any("principal", principal))
	}

	return c.Handler.WithAttrs(attrs).Handle(ctx, r)
}

func (c *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{Handler: c.Handler.WithAttrs(attrs)}
}

func (c *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{Handler: c.Handler.WithGroup(name)}
}
