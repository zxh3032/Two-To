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

export type UpdateProfileResponse = AccountMeResponse;

export type UpdateEmailResponse = Record<string, never>;

export type UpdatePhoneResponse = Record<string, never>;

export type UnbindEmailResponse = Record<string, never>;

export type UnbindPhoneResponse = Record<string, never>;

export type UpdatePasswordResponse = Record<string, never>;

export interface SessionsResponse {
  sessions: UserSession[];
  currentSessionId: number;
}

export type RevokeSessionResponse = Record<string, never>;

export type RevokeAllSessionsResponse = Record<string, never>;

export function getMe() {
  return get<AccountMeResponse>('/api/v1/account/me', { auth: true });
}

export function updateProfile(payload: UpdateProfilePayload) {
  return patch<UpdateProfileResponse>('/api/v1/account/profile', payload, { auth: true });
}

export function updateEmail(payload: { email: string; emailCode: string }) {
  return put<UpdateEmailResponse>('/api/v1/account/security/email', payload, { auth: true });
}

export function updatePhone(payload: { phone: string; smsCode: string }) {
  return put<UpdatePhoneResponse>('/api/v1/account/security/phone', payload, { auth: true });
}

export function unbindEmail(currentPassword: string) {
  return del<UnbindEmailResponse>('/api/v1/account/security/email', { currentPassword }, { auth: true });
}

export function unbindPhone(currentPassword: string) {
  return del<UnbindPhoneResponse>('/api/v1/account/security/phone', { currentPassword }, { auth: true });
}

export function updatePassword(payload: { currentPassword: string; newPassword: string }) {
  return put<UpdatePasswordResponse>('/api/v1/account/security/password', payload, { auth: true });
}

export function getSessions() {
  return get<SessionsResponse>('/api/v1/account/sessions', { auth: true });
}

export function revokeSession(sessionId: number) {
  return del<RevokeSessionResponse>(`/api/v1/account/sessions/${sessionId}`, undefined, { auth: true });
}

export function revokeAllSessions(currentPassword: string) {
  return post<RevokeAllSessionsResponse>('/api/v1/account/sessions/revoke-all', { currentPassword }, { auth: true });
}
