import type { UserInfo } from '../../entities/user/types';
import { get, post } from './http';

export interface CaptchaResponse {
  captchaId: string;
  imageBase64: string;
  expiresIn: number;
}

export interface SendCodePayload {
  codeType: 'sms' | 'email';
  scene: 'login' | 'register' | 'forgot_password' | 'bind_phone' | 'bind_email';
  target: string;
  captchaId: string;
  captchaCode: string;
}

export interface SendCodeResponse {
  cooldownSeconds: number;
  expiresIn: number;
}

export interface LoginResponse {
  accessToken: string;
  accessTokenExpiresIn: number;
  refreshToken: string;
  refreshTokenExpiresIn: number;
  user: UserInfo;
}

export type EmailLoginResponse = LoginResponse;

export interface PhoneLoginResponse {
  requiresProfileSetup: boolean;
  profileSetupToken: string;
  auth?: LoginResponse;
}

export interface EmailRegisterVerifyResponse {
  profileSetupToken: string;
}

export interface CompleteProfilePayload {
  profileSetupToken: string;
  nickname: string;
  petStage: string;
  interestedPetTypes: string[];
  petExperience: string;
  dailyCompanyTime: string;
  livingSituation: string;
  petConstraints: string[];
}

export interface TokenPair {
  accessToken: string;
  accessTokenExpiresIn: number;
  refreshToken: string;
  refreshTokenExpiresIn: number;
}

export type CompleteProfileResponse = LoginResponse;

export type RefreshResponse = TokenPair;

export type LogoutResponse = Record<string, never>;

export interface ForgotPasswordVerifyCodeResponse {
  resetToken: string;
}

export type ForgotPasswordResetResponse = Record<string, never>;

export function getCaptcha() {
  return get<CaptchaResponse>('/api/v1/auth/captcha');
}

export function sendCode(payload: SendCodePayload) {
  return post<SendCodeResponse>('/api/v1/auth/send-code', payload);
}

export function emailLogin(payload: { email: string; password: string; captchaId?: string; captchaCode?: string }) {
  return post<EmailLoginResponse>('/api/v1/auth/email-login', payload);
}

export function phoneLogin(payload: { phone: string; smsCode: string }) {
  return post<PhoneLoginResponse>('/api/v1/auth/phone-login', payload);
}

export function verifyEmailRegister(payload: { email: string; password: string; emailCode: string }) {
  return post<EmailRegisterVerifyResponse>('/api/v1/auth/email-register/verify', payload);
}

export function completeProfile(payload: CompleteProfilePayload) {
  return post<CompleteProfileResponse>('/api/v1/auth/complete-profile', payload);
}

export function refreshToken(refreshTokenValue: string) {
  return post<RefreshResponse>('/api/v1/auth/refresh', { refreshToken: refreshTokenValue });
}

export function logout(refreshTokenValue: string) {
  return post<LogoutResponse>('/api/v1/auth/logout', { refreshToken: refreshTokenValue });
}

export function verifyForgotPasswordCode(payload: { identityType: 'email' | 'phone'; identityValue: string; code: string }) {
  return post<ForgotPasswordVerifyCodeResponse>('/api/v1/auth/forgot-password/verify-code', payload);
}

export function resetForgotPassword(payload: { resetToken: string; newPassword: string }) {
  return post<ForgotPasswordResetResponse>('/api/v1/auth/forgot-password/reset', payload);
}
