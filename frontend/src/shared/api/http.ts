import { env } from '../config/env';
import { getAccessToken } from '../lib/auth-token';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
  requestId?: string;
}

export class ApiError extends Error {
  code: number;
  requestId?: string;
  status: number;
  data?: unknown;

  constructor(message: string, code: number, requestId?: string, status = 0, data?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.requestId = requestId;
    this.status = status;
    this.data = data;
  }
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';
  body?: unknown;
  auth?: boolean;
}

// request 统一处理请求、响应解析、鉴权 header 和业务错误码。
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const url = `${env.apiBaseUrl}${path}`;
  const headers: Record<string, string> = {
    Accept: 'application/json',
  };
  if (options.body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }
  if (options.auth) {
    const token = getAccessToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
  }

  const response = await fetch(url, {
    method: options.method ?? 'GET',
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  let body: ApiResponse<T>;
  try {
    body = (await response.json()) as ApiResponse<T>;
  } catch (error) {
    console.error('API 响应解析失败', { path, status: response.status, error });
    throw new ApiError('API 响应解析失败', response.status, undefined, response.status);
  }

  if (!response.ok || body.code !== 0) {
    console.error('API 请求失败', {
      path,
      status: response.status,
      code: body.code,
      message: body.message,
      requestId: body.requestId,
    });
    throw new ApiError(body.message || 'API 请求失败', body.code, body.requestId, response.status, body.data);
  }

  return body.data as T;
}

export function get<T>(path: string, options: Omit<RequestOptions, 'method' | 'body'> = {}): Promise<T> {
  return request<T>(path, { ...options, method: 'GET' });
}

export function post<T>(path: string, body?: unknown, options: Omit<RequestOptions, 'method' | 'body'> = {}): Promise<T> {
  return request<T>(path, { ...options, method: 'POST', body });
}

export function patch<T>(path: string, body?: unknown, options: Omit<RequestOptions, 'method' | 'body'> = {}): Promise<T> {
  return request<T>(path, { ...options, method: 'PATCH', body });
}

export function put<T>(path: string, body?: unknown, options: Omit<RequestOptions, 'method' | 'body'> = {}): Promise<T> {
  return request<T>(path, { ...options, method: 'PUT', body });
}

export function del<T>(path: string, body?: unknown, options: Omit<RequestOptions, 'method' | 'body'> = {}): Promise<T> {
  return request<T>(path, { ...options, method: 'DELETE', body });
}
