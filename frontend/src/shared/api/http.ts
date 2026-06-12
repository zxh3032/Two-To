import { env } from '../config/env';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
  requestId?: string;
}

export class ApiError extends Error {
  code: number;
  requestId?: string;

  constructor(message: string, code: number, requestId?: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.requestId = requestId;
  }
}

// get 统一处理 GET 请求、响应解析、错误日志和业务错误码。
export async function get<T>(path: string): Promise<T> {
  const url = `${env.apiBaseUrl}${path}`;
  const response = await fetch(url, {
    headers: {
      Accept: 'application/json',
    },
  });

  let body: ApiResponse<T>;
  try {
    body = (await response.json()) as ApiResponse<T>;
  } catch (error) {
    console.error('API 响应解析失败', { path, status: response.status, error });
    throw new ApiError('API 响应解析失败', response.status);
  }

  if (!response.ok || body.code !== 0) {
    console.error('API 请求失败', {
      path,
      status: response.status,
      code: body.code,
      message: body.message,
      requestId: body.requestId,
    });
    throw new ApiError(body.message || 'API 请求失败', body.code, body.requestId);
  }

  return body.data as T;
}
