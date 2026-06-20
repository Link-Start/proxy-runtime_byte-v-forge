package settingscore

import (
	"context"
	"errors"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

// WriteRuntimeSecret stores raw under secretID via writer and returns the
// resulting secret reference. The writer is required.
func WriteRuntimeSecret(ctx context.Context, writer secretref.Writer, raw string, secretID string, purpose string) (*commonv1.SecretRef, error) {
	if writer == nil {
		return nil, errors.New("proxy-runtime secret store is required")
	}
	return writer.WriteSecret(ctx, secretref.WriteRequest{
		SecretID: secretID,
		Provider: "proxy-runtime",
		Purpose:  purpose,
		Value:    raw,
	})
}
