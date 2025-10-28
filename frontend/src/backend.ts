import { defaultSettings, Settings } from "./types";
// existing Wails Manager bindings for jobs:
import { Jobs, SetInterval, Pause, Resume, Trigger, LoadSettings, SaveSettings, SyncJobsWithSettingsPublic, GetExecutions, ClearError } from "../wailsjs/go/backend/Manager";

export async function loadSettings(): Promise<Settings> {
  try {
    const settingsJSON = await LoadSettings();
    const parsed = JSON.parse(settingsJSON);
    return { ...defaultSettings, ...parsed }; // shallow merge defaults
  } catch (error) {
    return defaultSettings;
  }
}

export async function saveSettings(s: Settings): Promise<void> {
  try {
    await SaveSettings(JSON.stringify(s));
    // Sync jobs with the new settings (stop disabled jobs, start enabled ones)
    await SyncJobsWithSettingsPublic();
  } catch (error) {
    throw error;
  }
}

// Job control passthroughs (these already exist on backend.Manager)
export const backendJobs = { Jobs, SetInterval, Pause, Resume, Trigger, SyncJobsWithSettings: SyncJobsWithSettingsPublic, GetExecutions, ClearError };
