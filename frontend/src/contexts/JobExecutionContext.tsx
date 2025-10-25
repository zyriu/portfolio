import React, { createContext, useContext, useState, useCallback } from 'react';

export interface JobLog {
  timestamp: string;
  message: string;
  level: "info" | "error" | "success";
}

export interface JobExecution {
  id: string;
  jobName: string;
  startTime: string;
  endTime?: string;
  status: "running" | "completed" | "failed";
  logs: JobLog[];
}

interface JobExecutionContextType {
  executions: JobExecution[];
  setExecutions: (executions: JobExecution[]) => void;
}

const JobExecutionContext = createContext<JobExecutionContextType | undefined>(undefined);

export function JobExecutionProvider({ children }: { children: React.ReactNode }) {
  const [executions, setExecutions] = useState<JobExecution[]>([]);

  return (
    <JobExecutionContext.Provider value={{ executions, setExecutions }}>
      {children}
    </JobExecutionContext.Provider>
  );
}

export function useJobExecution() {
  const context = useContext(JobExecutionContext);
  if (!context) {
    throw new Error('useJobExecution must be used within a JobExecutionProvider');
  }
  return context;
}

