import React, { useEffect, useState } from "react";
import { backendJobs } from "../backend";
import { formatJobName } from "../utils/formatJobName";

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

function JobCard({ 
  job, 
  onRefreshJobs
}: { 
  job: JobState, 
  onRefreshJobs: () => Promise<void>
}) {
  const [now, setNow] = useState(Date.now() / 1000);
  const [isTriggering, setIsTriggering] = useState(false);

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now() / 1000), 1000);
    return () => clearInterval(timer);
  }, []);

  const isActivelyRunning = job.isExecuting;

  const timeUntilNext = Math.max(0, job.nextRunUnix - now);
  const timeSinceLast = now - job.lastRunUnix;
  const totalInterval = job.interval;
  
  // Progress: 0 at last run, 100 at next run (or special handling when running)
  let progress = 0;
  if (isActivelyRunning) {
    // When running, show a pulsing full bar
    progress = 100;
  } else if (totalInterval > 0) {
    progress = Math.min(100, (timeSinceLast / totalInterval) * 100);
  }

  const formatTime = (seconds: number) => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    if (h > 0) return `${h}h ${m}m ${s}s`;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
  };

  const handleTrigger = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isTriggering) return;
    
    setIsTriggering(true);
    
    try {
      await backendJobs.Trigger(job.name);
      await onRefreshJobs();
    } catch (err) {} finally {
      setTimeout(() => setIsTriggering(false), 2000);
    }
  };

  return (
    <div className="job-card">
      <div className="job-card-header">
        <div className="job-name-and-controls">
          <h3 className="job-name">{formatJobName(job.name)}</h3>
          <span className={`job-status ${isActivelyRunning ? 'executing' : job.running ? 'running' : 'paused'}`}>
            {isActivelyRunning ? '⚡ Executing' : job.running ? '● Running' : '○ Paused'}
          </span>
          <button 
            className="job-reload-btn"
            onClick={handleTrigger}
            title="Reload job"
            disabled={isTriggering || isActivelyRunning || !job.running}
          >
            {(isTriggering || isActivelyRunning) ? (
              <span className="spinner" />
            ) : (
              '↻'
            )}
          </button>
        </div>
      </div>
      
      <div className="job-progress-section">
        <div className="job-progress-container">
          <div 
            className={`job-progress-bar ${job.isExecuting ? 'running' : ''}`} 
            style={{ width: `${progress}%` }} 
          />
        </div>
        <div className="job-timer">
          {formatTime(timeUntilNext)}
        </div>
      </div>
    </div>
  );
}

export default function Dashboard({ 
  jobs, 
  onRefreshJobs
}: { 
  jobs: JobState[], 
  onRefreshJobs: () => Promise<void>
}) {

  const enabledJobs = jobs.filter(job => job.running);

  return (
    <div className="dashboard">
      {enabledJobs.length === 0 ? (
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          height: '100%',
          fontStyle: 'italic',
          opacity: 0.6
        }}>
          No jobs currently enabled.
        </div>
      ) : (
        <div className="jobs-grid">
          {jobs.map((job) => (
            <JobCard 
              key={job.name} 
              job={job} 
              onRefreshJobs={onRefreshJobs}
            />
          ))}
        </div>
      )}
    </div>
  );
}

