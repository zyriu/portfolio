import React, { useState, useEffect } from "react";
import { useJobExecution, JobExecution } from "../contexts/JobExecutionContext";
import { formatJobName } from "../utils/formatJobName";

function JobExecutionItem({ execution, onSelect, isSelected }: { 
  execution: JobExecution, 
  onSelect: (execution: JobExecution) => void,
  isSelected: boolean 
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
      className={`job-execution-item ${isSelected ? 'selected' : ''}`}
      onClick={() => onSelect(execution)}
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

export default function LogPanel() {
  const { executions } = useJobExecution();
  const [selectedExecution, setSelectedExecution] = useState<JobExecution | null>(null);
  const [isClosing, setIsClosing] = useState(false);
  const gridRef = React.useRef<HTMLDivElement>(null);
  const [columnsPerRow, setColumnsPerRow] = useState(1);

  // Calculate how many columns fit in the grid
  useEffect(() => {
    const updateColumns = () => {
      if (gridRef.current) {
        const gridWidth = gridRef.current.offsetWidth;
        const gap = 16; // gap from CSS
        const minColumnWidth = 280; // minmax(280px, 1fr) from CSS
        const columns = Math.floor((gridWidth + gap) / (minColumnWidth + gap));
        setColumnsPerRow(Math.max(1, columns));
      }
    };

    updateColumns();
    window.addEventListener('resize', updateColumns);
    return () => window.removeEventListener('resize', updateColumns);
  }, []);

  const handleToggleExecution = (execution: JobExecution) => {
    if (selectedExecution?.id === execution.id) {
      // Start closing animation
      setIsClosing(true);
      // Wait for animation to complete before removing
      setTimeout(() => {
        setSelectedExecution(null);
        setIsClosing(false);
      }, 300); // Match the animation duration
    } else {
      setSelectedExecution(execution);
      setIsClosing(false);
    }
  };

  // Calculate where to insert the card - after the last item in the current row
  const getCardInsertionIndex = () => {
    if (!selectedExecution) return -1;
    
    const selectedIndex = executions.findIndex(exec => exec.id === selectedExecution.id);
    if (selectedIndex === -1) return -1;

    // Calculate which row this item is in (0-based)
    const currentRow = Math.floor(selectedIndex / columnsPerRow);
    // Calculate the last index in this row
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
