package jobstatus

import (
	"context"
	"testing"
)

func TestWithStatusUpdater(t *testing.T) {
	called := false
	updater := func(status string) {
		called = true
		if status != "update" {
			t.Fatalf("unexpected status: %s", status)
		}
	}

	ctx := WithStatusUpdater(context.Background(), updater)
	got := GetStatusUpdater(ctx)
	got("update")

	if !called {
		t.Fatalf("expected updater to be called")
	}
}

func TestGetStatusUpdaterNoop(t *testing.T) {
	updater := GetStatusUpdater(context.Background())
	updater("ignored")
}
