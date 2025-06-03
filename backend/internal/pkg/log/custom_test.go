package log_test

import (
	"context"
	"log/slog"
	"testing"

	"vpainless/internal/pkg/log"
)

func TestCustomHandler_Handle(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		r       slog.Record
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var c log.CustomHandler
			gotErr := c.Handle(context.Background(), tt.r)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Handle() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Handle() succeeded unexpectedly")
			}
		})
	}
}
