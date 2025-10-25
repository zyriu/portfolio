package backup_grist

import (
	"context"
	"fmt"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/settings"
)

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Loading settings...")
	settingsData, err := settings.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %v", err)
	}

	// Validate backup path is configured
	if settingsData.Grist.BackupPath == "" {
		return fmt.Errorf("grist backup path not configured in settings")
	}

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Creating Grist document backup...")
	if err := g.BackupDocument(ctx, settingsData.Grist.BackupPath); err != nil {
		return err
	}

	updateStatus("✓ Backup completed successfully")
	return nil
}
