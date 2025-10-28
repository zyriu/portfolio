import React, { useEffect, useState } from "react";
import { Layout, Section } from "./components/Menu";
import Jobs from "./panels/Jobs";
import Logs from "./panels/Logs";
import { backendJobs } from "./backend";
import { JobExecutionProvider, useJobExecution } from "./contexts/JobExecutionContext";

interface JobState {
  name: string;
  interval: number;
  running: boolean;
  lastRunUnix: number;
  nextRunUnix: number;
  err: string;
  isExecuting: boolean;
  currentStatus: string;
}

function AppContent() {
  const [section, setSection] = useState<Section>("jobs");
  const [jobs, setJobs] = useState<JobState[]>([]);
  const [errorJobName, setErrorJobName] = useState<string | null>(null);
  const { setExecutions, executions } = useJobExecution();

  const fetchJobs = async () => {
    try {
      const result = await backendJobs.Jobs();
      const sortedJobs = (result || []).sort((a, b) => a.name.localeCompare(b.name));
      setJobs(sortedJobs);
    } catch (err) {
      // Handle error silently
    }
  };

  const fetchExecutions = async () => {
    try {
      const result = await backendJobs.GetExecutions();
      // Map backend executions to frontend format
      const executions = (result || []).map((exec: any) => ({
        id: exec.id,
        jobName: exec.jobName,
        startTime: exec.startTime,
        endTime: exec.endTime,
        status: exec.status,
        logs: (exec.logs || []).map((log: any) => ({
          timestamp: log.timestamp,
          message: log.message,
          level: log.level
        }))
      }));
      setExecutions(executions);
    } catch (err) {
      // Handle error silently
    }
  };

  const handleErrorClick = async (jobName: string) => {
    try {
      // Clear the error first
      await backendJobs.ClearError(jobName);
      
      // Set the job name for auto-expanding the log
      setErrorJobName(jobName);
      
      // Switch to logs panel
      setSection("logs");
      
      // Refresh jobs to update the UI
      await fetchJobs();
    } catch (err) {
      console.error("Failed to clear error:", err);
    }
  };

  // Poll jobs and executions continuously at the app level
  useEffect(() => {
    fetchJobs();
    fetchExecutions();
    const interval = setInterval(() => {
      fetchJobs();
      fetchExecutions();
    }, 1000);
    return () => clearInterval(interval);
  }, [setExecutions]);

  return (
    <Layout section={section} setSection={setSection}>
      {section === "jobs" && (
        <Jobs 
          jobs={jobs} 
          onRefreshJobs={fetchJobs}
          onErrorClick={handleErrorClick}
        />
      )}
      {section === "logs" && (
        <Logs 
          errorJobName={errorJobName}
          onErrorLogExpanded={() => setErrorJobName(null)}
        />
      )}
    </Layout>
  );
}

export default function App() {
  return (
    <JobExecutionProvider>
      <AppContent />
    </JobExecutionProvider>
  );
}
