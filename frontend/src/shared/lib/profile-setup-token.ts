const storageKey = 'two-to.profile-setup-token';

export function saveProfileSetupToken(token: string) {
  window.sessionStorage.setItem(storageKey, token);
}

export function getProfileSetupToken() {
  return window.sessionStorage.getItem(storageKey) ?? '';
}

export function clearProfileSetupToken() {
  window.sessionStorage.removeItem(storageKey);
}
