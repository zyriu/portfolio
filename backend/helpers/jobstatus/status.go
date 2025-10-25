package jobstatus

import "context"

// StatusUpdater is a function that jobs can call to update their status
type StatusUpdater func(status string)

type contextKey string

const statusUpdaterKey contextKey = "statusUpdater"

// GetStatusUpdater extracts the status updater from the context
func GetStatusUpdater(ctx context.Context) StatusUpdater {
	if updater, ok := ctx.Value(statusUpdaterKey).(StatusUpdater); ok {
		return updater
	}
	// Return a no-op function if no updater is found
	return func(status string) {}
}

// WithStatusUpdater adds a status updater to the context
func WithStatusUpdater(ctx context.Context, updater StatusUpdater) context.Context {
	return context.WithValue(ctx, statusUpdaterKey, updater)
}
