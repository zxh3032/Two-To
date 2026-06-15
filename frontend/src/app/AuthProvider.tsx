import { useCallback, useEffect, useState } from 'react';
import type { PropsWithChildren } from 'react';

import type { AccountMeResponse, UserProfile } from '../entities/user/types';
import { getMe } from '../shared/api/account';
import { type LoginResponse, logout as requestLogout, refreshToken } from '../shared/api/auth';
import { ApiError } from '../shared/api/http';
import {
  clearStoredTokens,
  getStoredTokens,
  saveStoredTokens,
  type StoredAuthTokens,
  toStoredTokens,
} from '../shared/lib/auth-token';
import { AuthContext, type AuthContextValue } from './auth-context';

export function AuthProvider({ children }: PropsWithChildren) {
  const [loading, setLoading] = useState(true);
  const [tokens, setTokens] = useState<StoredAuthTokens | null>(() => getStoredTokens());
  const [me, setMe] = useState<AccountMeResponse | null>(null);

  const applyTokens = useCallback((nextTokens: StoredAuthTokens | null) => {
    setTokens(nextTokens);
    if (nextTokens) {
      saveStoredTokens(nextTokens);
    } else {
      clearStoredTokens();
    }
  }, []);

  const refreshMe = useCallback(async () => {
    try {
      const data = await getMe();
      setMe(data);
      return data;
    } catch (error) {
      console.error('获取当前用户失败', { error });
      return null;
    }
  }, []);

  const tryRefresh = useCallback(async () => {
    const stored = getStoredTokens();
    if (!stored || stored.refreshTokenExpiresAt <= Date.now()) {
      applyTokens(null);
      setMe(null);
      return false;
    }
    try {
      const pair = await refreshToken(stored.refreshToken);
      applyTokens(toStoredTokens(pair));
      return true;
    } catch (error) {
      console.error('刷新登录态失败', { error });
      applyTokens(null);
      setMe(null);
      return false;
    }
  }, [applyTokens]);

  useEffect(() => {
    let active = true;
    async function bootstrap() {
      const stored = getStoredTokens();
      if (!stored) {
        if (active) {
          setLoading(false);
        }
        return;
      }
      if (stored.accessTokenExpiresAt <= Date.now()) {
        const refreshed = await tryRefresh();
        if (!refreshed) {
          if (active) {
            setLoading(false);
          }
          return;
        }
      }
      const account = await refreshMe();
      if (!account) {
        const refreshed = await tryRefresh();
        if (refreshed) {
          await refreshMe();
        }
      }
      if (active) {
        setLoading(false);
      }
    }
    void bootstrap();
    return () => {
      active = false;
    };
  }, [refreshMe, tryRefresh]);

  const loginWithAuth = useCallback(
    async (payload: LoginResponse) => {
      applyTokens(toStoredTokens(payload));
      setMe({ user: payload.user, profile: emptyProfile(), identities: [] });
      await refreshMe();
    },
    [applyTokens, refreshMe],
  );

  const logout = useCallback(async () => {
    const refreshTokenValue = tokens?.refreshToken;
    applyTokens(null);
    setMe(null);
    if (!refreshTokenValue) {
      return;
    }
    try {
      await requestLogout(refreshTokenValue);
    } catch (error) {
      if (!(error instanceof ApiError && error.status === 401)) {
        console.error('退出登录失败', { error });
      }
    }
  }, [applyTokens, tokens?.refreshToken]);

  const value: AuthContextValue = {
    loading,
    authenticated: Boolean(tokens?.accessToken && me?.user),
    user: me?.user ?? null,
    profile: me?.profile ?? null,
    identities: me?.identities ?? [],
    loginWithAuth,
    refreshMe,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

function emptyProfile(): UserProfile {
  return {
    petStage: '',
    interestedPetTypes: [],
    petExperience: '',
    dailyCompanyTime: '',
    livingSituation: '',
    petConstraints: [],
  };
}
