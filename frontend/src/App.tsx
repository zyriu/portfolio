import React, { useEffect, useState } from "react";
import { Layout, Section } from "./components/Menu";
import Dashboard from "./panels/Dashboard";
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
  const [section, setSection] = useState<Section>("dashboard");
  const [jobs, setJobs] = useState<JobState[]>([]);
  const [selectedErrorJob, setSelectedErrorJob] = useState<string | undefined>(undefined);
  const [acknowledgedErrors, setAcknowledgedErrors] = useState<Set<string>>(new Set());
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

  const handleNavigateToError = (jobName: string) => {
    setSelectedErrorJob(jobName);
    setSection("logs");
    // Mark this error as acknowledged
    setAcknowledgedErrors(prev => new Set(prev).add(jobName));
  };

  // Clear selected error job when navigating away from logs
  useEffect(() => {
    if (section !== "logs") {
      setSelectedErrorJob(undefined);
    }
  }, [section]);

  // Clear acknowledged errors when job succeeds or error changes
  useEffect(() => {
    jobs.forEach(job => {
      // If job no longer has an error, remove it from acknowledged set
      if (!job.err && acknowledgedErrors.has(job.name)) {
        setAcknowledgedErrors(prev => {
          const newSet = new Set(prev);
          newSet.delete(job.name);
          return newSet;
        });
      }
    });
  }, [jobs, acknowledgedErrors]);

  return (
    <Layout section={section} setSection={setSection}>
      {section === "dashboard" && (
        <Dashboard 
          jobs={jobs} 
          onRefreshJobs={fetchJobs}
          onNavigateToError={handleNavigateToError}
          acknowledgedErrors={acknowledgedErrors}
        />
      )}
      {section === "logs" && <Logs initialJobName={selectedErrorJob} />}
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
