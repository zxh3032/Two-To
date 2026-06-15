export interface StoredAuthTokens {
  accessToken: string;
  refreshToken: string;
  accessTokenExpiresAt: number;
  refreshTokenExpiresAt: number;
}

const storageKey = 'two-to.auth.tokens';
let memoryTokens: string | null = null;

export function getStoredTokens(): StoredAuthTokens | null {
  const raw = getStorage()?.getItem(storageKey) ?? memoryTokens;
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as StoredAuthTokens;
  } catch (error) {
    console.error('登录态读取失败', { error });
    getStorage()?.removeItem(storageKey);
    memoryTokens = null;
    return null;
  }
}

export function saveStoredTokens(tokens: StoredAuthTokens) {
  const raw = JSON.stringify(tokens);
  const storage = getStorage();
  if (storage) {
    storage.setItem(storageKey, raw);
    return;
  }
  memoryTokens = raw;
}

export function clearStoredTokens() {
  getStorage()?.removeItem(storageKey);
  memoryTokens = null;
}

export function getAccessToken(): string | null {
  return getStoredTokens()?.accessToken ?? null;
}

export function toStoredTokens(payload: {
  accessToken: string;
  accessTokenExpiresIn: number;
  refreshToken: string;
  refreshTokenExpiresIn: number;
}): StoredAuthTokens {
  const now = Date.now();
  return {
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
    accessTokenExpiresAt: now + payload.accessTokenExpiresIn * 1000,
    refreshTokenExpiresAt: now + payload.refreshTokenExpiresIn * 1000,
  };
}

function getStorage(): Storage | null {
  if (typeof window === 'undefined' || !window.localStorage) {
    return null;
  }
  if (typeof window.localStorage.getItem !== 'function' || typeof window.localStorage.setItem !== 'function') {
    return null;
  }
  return window.localStorage;
}
