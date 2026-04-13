// Task status values
export type TaskStatusValue = 'pending' | 'processing' | 'completed' | 'failed' | 'expired';

// Task status interface
export interface TaskStatus {
  id: string;
  status: TaskStatusValue;
  progress: number;
  current_batch?: number;
  total_batches?: number;
  source_lang: string;
  target_lang: string;
  file_name: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  result_file_path?: string;
  error?: string;
}

// Task creation response
export interface TaskCreationResponse {
  task_id: string;
  status: string;
  message: string;
}

// Task result wrapper
export interface TaskResult {
  success: boolean;
  data?: TaskStatus;
  error?: {
    code: string;
    message: string;
  };
}

// Translation document request
export interface TranslateDocumentRequest {
  source_lang: string;
  target_lang: string;
}

// Translation request
export interface TranslateRequest {
  text: string;
  source_lang: string;
  target_lang: string;
}

// Translation response data
export interface TranslateData {
  result: string;
  id: number;
  alternatives: string[];
}

// Translation response
export interface TranslateResponse {
  success: boolean;
  data?: TranslateData;
  error?: {
    code: string;
    message: string;
  };
}

// Health check response
export interface HealthCheckResponse {
  status: string;
}

// Auth verify request
export interface AuthVerifyRequest {
  token: string;
}

// Auth verify response
export interface AuthVerifyResponse {
  success: boolean;
  data?: { valid: boolean };
  error?: { code: string; message: string };
}
