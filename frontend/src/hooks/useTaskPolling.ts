import { useState, useEffect, useRef, useCallback } from 'react';
import { api } from '../lib/api';
import type { TaskStatus } from '../lib/types';

interface UseTaskPollingOptions {
  taskId: string;
  onProgress?: (progress: number, currentBatch: number, totalBatches: number) => void;
  onCompleted?: (taskStatus: TaskStatus) => void;
  onFailed?: (error: string) => void;
  initialInterval?: number; // Default 2000ms (2 seconds)
  maxInterval?: number;     // Default 5000ms (5 seconds)
  enabled?: boolean;        // Enable/disable polling
}

export function useTaskPolling({
  taskId,
  onProgress,
  onCompleted,
  onFailed,
  initialInterval = 2000,
  maxInterval = 5000,
  enabled = true,
}: UseTaskPollingOptions) {
  const [taskStatus, setTaskStatus] = useState<TaskStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPolling, setIsPolling] = useState(false);
  const intervalRef = useRef<number>(initialInterval);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastProgressRef = useRef<number>(0);
  const stagnationCountRef = useRef<number>(0);
  const hasStartedRef = useRef<boolean>(false); // Track if polling has started

  // Adaptive polling interval logic
  const updatePollingInterval = useCallback((currentProgress: number) => {
    // If progress hasn't changed, increase stagnation counter
    if (currentProgress === lastProgressRef.current) {
      stagnationCountRef.current += 1;
    } else {
      stagnationCountRef.current = 0;
      lastProgressRef.current = currentProgress;
    }

    // Adjust interval based on progress and stagnation
    if (currentProgress > 0 && currentProgress < 100) {
      // Progress is being made, use 2 second interval
      intervalRef.current = 2000;
    } else if (currentProgress === 0) {
      // Task not started yet (pending), use slower polling
      intervalRef.current = 3000;
    } else if (stagnationCountRef.current >= 3) {
      // No progress for 3+ consecutive polls, slow down to 5 seconds
      intervalRef.current = maxInterval;
    } else {
      // Completed or other state, use initial interval
      intervalRef.current = initialInterval;
    }
  }, [initialInterval, maxInterval]);

  // Stop polling
  const stopPolling = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    setIsPolling(false);
  }, []);

  // Poll task status
  const poll = useCallback(async () => {
    try {
      const status = await api.getTaskStatus(taskId);
      setTaskStatus(status);
      setError(null);

      // Call progress callback if available
      if (onProgress && status.current_batch && status.total_batches) {
        onProgress(status.progress, status.current_batch, status.total_batches);
      }

      // Update polling interval based on progress
      updatePollingInterval(status.progress);

      // Check if task is completed
      if (status.status === 'completed') {
        stopPolling();
        if (onCompleted) {
          onCompleted(status);
        }
        return;
      }

      // Check if task failed
      if (status.status === 'failed') {
        stopPolling();
        const errorMessage = status.error || 'Translation failed';
        setError(errorMessage);
        if (onFailed) {
          onFailed(errorMessage);
        }
        return;
      }

      // Continue polling if task is still processing
      if (status.status === 'pending' || status.status === 'processing') {
        timeoutRef.current = setTimeout(poll, intervalRef.current);
      } else {
        // Task expired or other final state
        stopPolling();
      }
    } catch (err) {
      stopPolling();
      const errorMessage = err instanceof Error ? err.message : 'Failed to check task status';
      setError(errorMessage);
      if (onFailed) {
        onFailed(errorMessage);
      }
    }
  }, [taskId, onProgress, onCompleted, onFailed, updatePollingInterval, stopPolling]);

  // Start polling
  const startPolling = useCallback(() => {
    if (isPolling) return;

    setIsPolling(true);
    lastProgressRef.current = 0;
    stagnationCountRef.current = 0;
    intervalRef.current = initialInterval;

    // Initial poll
    poll();
  }, [isPolling, initialInterval, poll]);

  // Auto-start polling when enabled and taskId changes
  useEffect(() => {
    // Only start if enabled, has taskId, and hasn't started yet
    if (enabled && taskId && !hasStartedRef.current) {
      hasStartedRef.current = true;
      setIsPolling(true);
      lastProgressRef.current = 0;
      stagnationCountRef.current = 0;
      intervalRef.current = initialInterval;

      // Initial poll
      poll();
    }

    // Cleanup on unmount or when taskId changes
    return () => {
      stopPolling();
      hasStartedRef.current = false;
    };
  }, [enabled, taskId]); // Only depend on enabled and taskId

  return {
    taskStatus,
    error,
    isPolling,
    startPolling,
    stopPolling,
    currentInterval: intervalRef.current,
  };
}
