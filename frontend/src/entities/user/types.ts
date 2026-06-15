export interface UserInfo {
  id: number;
  nickname: string;
  status: number;
}

export interface UserProfile {
  petStage: string;
  interestedPetTypes: string[];
  petExperience: string;
  dailyCompanyTime: string;
  livingSituation: string;
  petConstraints: string[];
}

export interface UserIdentity {
  id: number;
  identityType: 'email' | 'phone';
  identityValue: string;
  maskedValue: string;
  verifyTime: number;
}

export interface UserSession {
  id: number;
  deviceName: string;
  status: number;
  lastActiveTime: number;
  expireTime: number;
  current: boolean;
}

export interface AccountMeResponse {
  user: UserInfo;
  profile: UserProfile;
  identities: UserIdentity[];
}
