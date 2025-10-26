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
  const { setExecutions } = useJobExecution();

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
        />
      )}
      {section === "logs" && <Logs />}
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
