import type {
  TaskStatus,
  TaskCreationResponse,
  TaskResult,
  TranslateDocumentRequest,
  TranslateRequest,
  TranslateResponse,
  HealthCheckResponse,
  AuthVerifyResponse,
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';
const TOKEN_KEY = 'deeplx_token';

class UnauthorizedError extends Error {
  constructor() {
    super('Unauthorized');
    this.name = 'UnauthorizedError';
  }
}

export { UnauthorizedError };

function getAuthHeaders(): HeadersInit {
  const headers: HeadersInit = {};
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (response.status === 401) {
    throw new UnauthorizedError();
  }
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: { message: response.statusText } }));
    throw new Error(error.error?.message || `Request failed: ${response.statusText}`);
  }
  return response.json();
}

export const api = {
  async verifyToken(token: string): Promise<AuthVerifyResponse> {
    const response = await fetch(`${API_BASE_URL}/auth/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token }),
    });
    return response.json();
  },

  async translate(request: TranslateRequest): Promise<TranslateResponse> {
    const response = await fetch(`${API_BASE_URL}/translate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
      },
      body: JSON.stringify(request),
    });
    return handleResponse<TranslateResponse>(response);
  },

  async translateDocument(
    file: File,
    request: TranslateDocumentRequest
  ): Promise<Blob> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('source_lang', request.source_lang);
    formData.append('target_lang', request.target_lang);

    const response = await fetch(`${API_BASE_URL}/translate/document`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: formData,
    });

    if (response.status === 401) throw new UnauthorizedError();
    if (!response.ok) {
      throw new Error(`Document translation failed: ${response.statusText}`);
    }
    return response.blob();
  },

  async healthCheck(): Promise<HealthCheckResponse> {
    const response = await fetch('/health');
    if (!response.ok) {
      throw new Error('Health check failed');
    }
    return response.json();
  },

  async createTranslationTask(
    file: File,
    request: TranslateDocumentRequest
  ): Promise<TaskCreationResponse> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('source_lang', request.source_lang);
    formData.append('target_lang', request.target_lang);

    const response = await fetch(`${API_BASE_URL}/translate/document`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: formData,
    });

    if (response.status === 401) throw new UnauthorizedError();
    const result = await response.json();
    if (!response.ok) {
      throw new Error(result.error?.message || 'Failed to create translation task');
    }
    return result.data;
  },

  async getTaskStatus(taskId: string): Promise<TaskStatus> {
    const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/status`, {
      headers: getAuthHeaders(),
    });

    if (response.status === 401) throw new UnauthorizedError();
    const result: TaskResult = await response.json();
    if (!response.ok || !result.success || !result.data) {
      throw new Error(result.error?.message || 'Failed to get task status');
    }
    return result.data;
  },

  async downloadTaskResult(taskId: string): Promise<Blob> {
    const response = await fetch(`${API_BASE_URL}/tasks/${taskId}/download`, {
      headers: getAuthHeaders(),
    });

    if (response.status === 401) throw new UnauthorizedError();
    if (response.status === 202) {
      throw new Error('Task is still processing');
    } else if (response.status === 400) {
      const error = await response.json();
      throw new Error(error.error?.message || 'Task failed');
    } else if (!response.ok) {
      throw new Error('Task not found or expired');
    }
    return response.blob();
  },
};
