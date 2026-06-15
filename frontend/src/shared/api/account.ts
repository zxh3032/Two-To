import type { AccountMeResponse, UserSession } from '../../entities/user/types';
import { del, get, patch, put, post } from './http';

export interface UpdateProfilePayload {
  nickname: string;
  petStage: string;
  interestedPetTypes: string[];
  petExperience: string;
  dailyCompanyTime: string;
  livingSituation: string;
  petConstraints: string[];
}

export function getMe() {
  return get<AccountMeResponse>('/api/v1/account/me', { auth: true });
}

export function updateProfile(payload: UpdateProfilePayload) {
  return patch<AccountMeResponse>('/api/v1/account/profile', payload, { auth: true });
}

export function updateEmail(payload: { email: string; emailCode: string }) {
  return put<void>('/api/v1/account/security/email', payload, { auth: true });
}

export function updatePhone(payload: { phone: string; smsCode: string }) {
  return put<void>('/api/v1/account/security/phone', payload, { auth: true });
}

export function unbindEmail(currentPassword: string) {
  return del<void>('/api/v1/account/security/email', { currentPassword }, { auth: true });
}

export function unbindPhone(currentPassword: string) {
  return del<void>('/api/v1/account/security/phone', { currentPassword }, { auth: true });
}

export function updatePassword(payload: { currentPassword: string; newPassword: string }) {
  return put<void>('/api/v1/account/security/password', payload, { auth: true });
}

export function getSessions() {
  return get<{ sessions: UserSession[]; currentSessionId: number }>('/api/v1/account/sessions', { auth: true });
}

export function revokeSession(sessionId: number) {
  return del<void>(`/api/v1/account/sessions/${sessionId}`, undefined, { auth: true });
}

export function revokeAllSessions(currentPassword: string) {
  return post<void>('/api/v1/account/sessions/revoke-all', { currentPassword }, { auth: true });
}
