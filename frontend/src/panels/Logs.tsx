import React, { useState, useEffect } from "react";
import { useJobExecution, JobExecution } from "../contexts/JobExecutionContext";
import { formatJobName } from "../utils/formatJobName";
import { backendJobs } from "../backend";

function JobExecutionItem({ execution, onSelect, isSelected, isExecuting }: {
  execution: JobExecution,
  onSelect: (execution: JobExecution) => void,
  isSelected: boolean,
  isExecuting: boolean
}) {
  const formatTime = (timeStr: string) => {
    return new Date(timeStr).toLocaleString();
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return '⚡';
      case 'completed': return '✅';
      case 'failed': return '❌';
      default: return '⏳';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'running': return 'EXECUTING';
      case 'completed': return 'COMPLETED';
      case 'failed': return 'FAILED';
      default: return 'UNKNOWN';
    }
  };

  return (
    <div
      className={`job-execution-item ${isSelected ? 'selected' : ''} ${isExecuting ? 'executing' : ''}`}
      onClick={() => onSelect(execution)}
      style={{
        cursor: isExecuting ? 'not-allowed' : 'pointer',
        opacity: isExecuting ? 0.6 : 1
      }}
    >
      <div className="execution-header">
        <span className={`execution-status ${execution.status}`}>
          {getStatusIcon(execution.status)} {getStatusText(execution.status)}
        </span>
        <span className="execution-time">
          {formatTime(execution.startTime)}
        </span>
      </div>
      <div className="execution-job-name">
        {formatJobName(execution.jobName)}
        {execution.endTime && (
          <span className="execution-duration">
            Duration: {Math.round((new Date(execution.endTime).getTime() - new Date(execution.startTime).getTime()) / 1000)}s
          </span>
        )}
      </div>
    </div>
  );
}

function LogViewerCard({ execution }: { execution: JobExecution }) {
  return (
    <div className="log-viewer-card">
      <div className="log-viewer-card-header">
        <div className="log-viewer-card-title">
          <span className="log-viewer-card-job-name">{formatJobName(execution.jobName)}</span>
          <span className="log-viewer-card-time">{new Date(execution.startTime).toLocaleString()}</span>
        </div>
        <div className="log-viewer-card-controls">
          <span className={`execution-status-badge ${execution.status}`}>
            {execution.status.toUpperCase()}
          </span>
        </div>
      </div>
      <div className="log-viewer-card-content">
        {execution.logs.length > 0 ? (
          execution.logs.map((log, idx) => (
            <div key={idx} className={`log-entry ${log.level}`}>
              <span className="log-timestamp">{log.timestamp}</span>
              <span className="log-message">{log.message}</span>
            </div>
          ))
        ) : (
          <div className="no-logs">No logs available for this execution.</div>
        )}
      </div>
    </div>
  );
}

export default function Logs({
  errorJobName,
  onErrorLogExpanded
}: {
  errorJobName: string | null;
  onErrorLogExpanded: () => void;
}) {
  const { executions } = useJobExecution();
  const [selectedExecution, setSelectedExecution] = useState<JobExecution | null>(null);
  const [isClosing, setIsClosing] = useState(false);
  const gridRef = React.useRef<HTMLDivElement>(null);
  const [columnsPerRow, setColumnsPerRow] = useState(1);
  const [executingJobs, setExecutingJobs] = useState<Set<string>>(new Set());
  const [currentExecutingExecutions, setCurrentExecutingExecutions] = useState<Map<string, JobExecution>>(new Map());

  useEffect(() => {
    const fetchJobStates = async () => {
      try {
        const jobs = await backendJobs.Jobs();
        const executing = new Set<string>();
        jobs.forEach((job: any) => {
          if (job.isExecuting) {
            executing.add(job.name);
          }
        });
        setExecutingJobs(executing);

        const runningExecutions = executions.filter(exec => exec.status === 'running');
        const executingExecutionsMap = new Map<string, JobExecution>();

        runningExecutions.forEach(exec => {
          const existing = executingExecutionsMap.get(exec.jobName);
          if (!existing || new Date(exec.startTime).getTime() > new Date(existing.startTime).getTime()) {
            executingExecutionsMap.set(exec.jobName, exec);
          }
        });

        setCurrentExecutingExecutions(executingExecutionsMap);
      } catch (err) {
      }
    };

    fetchJobStates();
    const interval = setInterval(fetchJobStates, 1000);
    return () => clearInterval(interval);
  }, [executions]);

  useEffect(() => {
    const updateColumns = () => {
      if (gridRef.current) {
        const gridWidth = gridRef.current.offsetWidth;
        const gap = 16;
        const minColumnWidth = 280;
        const columns = Math.floor((gridWidth + gap) / (minColumnWidth + gap));
        setColumnsPerRow(Math.max(1, columns));
      }
    };

    updateColumns();
    window.addEventListener('resize', updateColumns);
    return () => window.removeEventListener('resize', updateColumns);
  }, []);

  useEffect(() => {
    if (errorJobName && executions.length > 0) {
      const failedExecutions = executions.filter(exec =>
        exec.jobName === errorJobName && exec.status === 'failed'
      );

      if (failedExecutions.length > 0) {
        const mostRecentFailed = failedExecutions.sort((a, b) =>
          new Date(b.startTime).getTime() - new Date(a.startTime).getTime()
        )[0];

        setTimeout(() => {
          setSelectedExecution(mostRecentFailed);
          onErrorLogExpanded();
        }, 100);
      } else {
        onErrorLogExpanded();
      }
    }
  }, [errorJobName, executions, onErrorLogExpanded]);

  const handleToggleExecution = (execution: JobExecution) => {
    if (currentExecutingExecutions.get(execution.jobName)?.id === execution.id) {
      return;
    }

    if (selectedExecution?.id === execution.id) {
      setIsClosing(true);
      setTimeout(() => {
        setSelectedExecution(null);
        setIsClosing(false);
      }, 300);
    } else {
      setSelectedExecution(execution);
      setIsClosing(false);
    }
  };

  const getCardInsertionIndex = () => {
    if (!selectedExecution) return -1;

    const selectedIndex = executions.findIndex(exec => exec.id === selectedExecution.id);
    if (selectedIndex === -1) return -1;

    const currentRow = Math.floor(selectedIndex / columnsPerRow);
    const lastIndexInRow = Math.min((currentRow + 1) * columnsPerRow - 1, executions.length - 1);

    return lastIndexInRow;
  };

  const cardInsertionIndex = getCardInsertionIndex();

  return (
    <div className="panel-logs-container">
      <div className="executions-list">
        {executions.length === 0 ? (
          <div className="no-executions">
            No job executions found. Jobs will appear here once they run.
          </div>
        ) : (
          <div className="executions-grid" ref={gridRef}>
            {executions.map((execution, index) => (
              <React.Fragment key={execution.id}>
                <JobExecutionItem
                  execution={execution}
                  onSelect={handleToggleExecution}
                  isSelected={selectedExecution?.id === execution.id}
                  isExecuting={currentExecutingExecutions.get(execution.jobName)?.id === execution.id}
                />
                {selectedExecution && cardInsertionIndex === index && (
                  <div className={`log-viewer-card-wrapper ${isClosing ? 'closing' : ''}`}>
                    <LogViewerCard execution={selectedExecution} />
                  </div>
                )}
              </React.Fragment>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
