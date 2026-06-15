import { createContext } from 'react';

import type { AccountMeResponse, UserIdentity, UserInfo, UserProfile } from '../entities/user/types';
import type { LoginResponse } from '../shared/api/auth';

export interface AuthContextValue {
  loading: boolean;
  authenticated: boolean;
  user: UserInfo | null;
  profile: UserProfile | null;
  identities: UserIdentity[];
  loginWithAuth: (payload: LoginResponse) => Promise<void>;
  refreshMe: () => Promise<AccountMeResponse | null>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | null>(null);
