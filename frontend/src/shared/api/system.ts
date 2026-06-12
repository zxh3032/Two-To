import { get } from './http';

export interface PingData {
  message: string;
  service: string;
  environment: string;
  requestId: string;
  timestamp: number;
}

export interface SloganData {
  slogan: string;
  requestId: string;
}

// fetchPing 调用后端 ping 接口，用于验证前后端基础联通链路。
export function fetchPing() {
  return get<PingData>('/api/v1/ping');
}

// fetchSlogan 调用后端 slogan 接口，用于展示 Two-To 项目核心口号。
export function fetchSlogan() {
  return get<SloganData>('/api/v1/slogan');
}
