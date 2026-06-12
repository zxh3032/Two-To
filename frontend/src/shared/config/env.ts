interface EnvConfig {
  apiBaseUrl: string;
}

// env 集中读取 Vite 环境变量，避免页面和 feature 直接访问 import.meta.env。
export const env: EnvConfig = {
  apiBaseUrl: (import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:0806').replace(/\/$/, ''),
};
