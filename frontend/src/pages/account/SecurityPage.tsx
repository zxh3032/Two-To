import { Check, KeyRound, Mail, RefreshCcw, Send, ShieldCheck, Smartphone, Trash2, X } from 'lucide-react';
import { useEffect, useState } from 'react';

import { useAuth } from '../../app/useAuth';
import { useCaptcha } from '../../features/auth-form/useCaptcha';
import { useCountdown } from '../../features/auth-form/useCountdown';
import type { UserSession } from '../../entities/user/types';
import {
  getSessions,
  revokeAllSessions,
  revokeSession,
  unbindEmail,
  unbindPhone,
  updateEmail,
  updatePassword,
  updatePhone,
  updateProfile,
} from '../../shared/api/account';
import { sendCode } from '../../shared/api/auth';

type SensitiveAction = 'unbind-email' | 'unbind-phone' | 'revoke-all';

interface PasswordDialogState {
  action: SensitiveAction;
  title: string;
  description: string;
  confirmLabel: string;
  danger?: boolean;
}

export function SecurityPage() {
  const auth = useAuth();
  const captcha = useCaptcha();
  const countdown = useCountdown();
  const [nickname, setNickname] = useState(auth.user?.nickname ?? '');
  const [petStage, setPetStage] = useState(auth.profile?.petStage ?? 'planning');
  const [interestedPetTypes, setInterestedPetTypes] = useState<string[]>(auth.profile?.interestedPetTypes?.length ? auth.profile.interestedPetTypes : ['dog']);
  const [email, setEmail] = useState('');
  const [emailCode, setEmailCode] = useState('');
  const [phone, setPhone] = useState('');
  const [smsCode, setSmsCode] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [sessions, setSessions] = useState<UserSession[]>([]);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [passwordDialog, setPasswordDialog] = useState<PasswordDialogState | null>(null);
  const [confirmationPassword, setConfirmationPassword] = useState('');
  const [dialogError, setDialogError] = useState('');
  const [dialogSubmitting, setDialogSubmitting] = useState(false);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setNickname(auth.user?.nickname ?? '');
      setPetStage(auth.profile?.petStage || 'planning');
      setInterestedPetTypes(auth.profile?.interestedPetTypes?.length ? auth.profile.interestedPetTypes : ['dog']);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [auth.profile, auth.user]);

  useEffect(() => {
    void loadSessions();
  }, []);

  async function loadSessions() {
    const data = await getSessions();
    setSessions(data.sessions);
  }

  function togglePet(value: string) {
    setInterestedPetTypes((current) => (current.includes(value) ? current.filter((item) => item !== value) : [...current, value]));
  }

  async function handleUpdateProfile() {
    await run(async () => {
      await updateProfile({
        nickname,
        petStage,
        interestedPetTypes,
        petExperience: auth.profile?.petExperience ?? '',
        dailyCompanyTime: auth.profile?.dailyCompanyTime ?? '',
        livingSituation: auth.profile?.livingSituation ?? '',
        petConstraints: auth.profile?.petConstraints ?? [],
      });
      await auth.refreshMe();
      setMessage('资料已保存');
    });
  }

  async function handleSendIdentityCode(kind: 'email' | 'phone') {
    await run(async () => {
      if (!captcha.captcha || !captcha.captchaCode) {
        throw new Error('请先输入图片验证码');
      }
      const data = await sendCode({
        codeType: kind === 'email' ? 'email' : 'sms',
        scene: kind === 'email' ? 'bind_email' : 'bind_phone',
        target: kind === 'email' ? email : phone,
        captchaId: captcha.captcha.captchaId,
        captchaCode: captcha.captchaCode,
      });
      countdown.start(data.cooldownSeconds);
      await captcha.reloadCaptcha();
      setMessage('验证码已发送');
    });
  }

  async function handleBindEmail() {
    await run(async () => {
      await updateEmail({ email, emailCode });
      await auth.refreshMe();
      setMessage('邮箱已更新');
    });
  }

  async function handleBindPhone() {
    await run(async () => {
      await updatePhone({ phone, smsCode });
      await auth.refreshMe();
      setMessage('手机号已更新');
    });
  }

  async function handleUnbind(kind: 'email' | 'phone') {
    openPasswordDialog({
      action: kind === 'email' ? 'unbind-email' : 'unbind-phone',
      title: kind === 'email' ? '解绑邮箱' : '解绑手机号',
      description: '请输入当前密码确认本次账号安全变更。',
      confirmLabel: '确认解绑',
      danger: true,
    });
  }

  async function handlePassword() {
    await run(async () => {
      await updatePassword({ currentPassword, newPassword });
      setCurrentPassword('');
      setNewPassword('');
      await loadSessions();
      setMessage('密码已更新');
    });
  }

  async function handleRevoke(sessionId: number) {
    await run(async () => {
      await revokeSession(sessionId);
      await loadSessions();
      setMessage('设备已移除');
    });
  }

  async function handleRevokeAll() {
    openPasswordDialog({
      action: 'revoke-all',
      title: '退出全部设备',
      description: '确认后当前账号会从所有设备退出，需要重新登录。',
      confirmLabel: '退出全部设备',
      danger: true,
    });
  }

  function openPasswordDialog(nextDialog: PasswordDialogState) {
    setPasswordDialog(nextDialog);
    setConfirmationPassword('');
    setDialogError('');
    setError('');
    setMessage('');
  }

  function closePasswordDialog() {
    if (dialogSubmitting) {
      return;
    }
    setPasswordDialog(null);
    setConfirmationPassword('');
    setDialogError('');
  }

  async function handleConfirmPasswordDialog() {
    if (!passwordDialog) {
      return;
    }
    if (!confirmationPassword) {
      setDialogError('请输入当前密码');
      return;
    }

    setDialogSubmitting(true);
    setDialogError('');
    setError('');
    setMessage('');

    try {
      if (passwordDialog.action === 'unbind-email') {
        await unbindEmail(confirmationPassword);
        await auth.refreshMe();
        setMessage('邮箱已解绑');
      }
      if (passwordDialog.action === 'unbind-phone') {
        await unbindPhone(confirmationPassword);
        await auth.refreshMe();
        setMessage('手机号已解绑');
      }
      if (passwordDialog.action === 'revoke-all') {
        await revokeAllSessions(confirmationPassword);
        setPasswordDialog(null);
        setConfirmationPassword('');
        await auth.logout();
        return;
      }
      setPasswordDialog(null);
      setConfirmationPassword('');
    } catch (err) {
      setDialogError(errorMessage(err));
    } finally {
      setDialogSubmitting(false);
    }
  }

  async function run(action: () => Promise<void>) {
    setError('');
    setMessage('');
    try {
      await action();
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  const emailIdentity = auth.identities.find((item) => item.identityType === 'email');
  const phoneIdentity = auth.identities.find((item) => item.identityType === 'phone');

  return (
    <section className="page-stack security-page">
      <div className="page-header">
        <div>
          <p className="eyebrow">账号与安全</p>
          <h1>资料、联系方式和设备</h1>
        </div>
        <button className="button-link" onClick={auth.logout} type="button">
          退出登录
        </button>
      </div>

      {message ? <p className="notice notice--ok">{message}</p> : null}
      {error ? <p className="notice notice--error">{error}</p> : null}

      <section className="settings-section">
        <div>
          <h2>基础资料</h2>
          <p>这些资料会用于后续适配测评。</p>
        </div>
        <div className="settings-form">
          <label className="field">
            <span>昵称</span>
            <input value={nickname} onChange={(event) => setNickname(event.target.value)} />
          </label>
          <label className="field">
            <span>养宠阶段</span>
            <select value={petStage} onChange={(event) => setPetStage(event.target.value)}>
              <option value="planning">准备养宠</option>
              <option value="comparing">正在比较品种</option>
              <option value="already_have">已经有宠物</option>
              <option value="helping_family">帮家人做选择</option>
            </select>
          </label>
          <div className="choice-grid">
            {[
              ['dog', '小狗'],
              ['cat', '小猫'],
              ['small_pet', '小型宠物'],
              ['not_sure', '还不确定'],
            ].map(([value, label]) => (
              <button className={interestedPetTypes.includes(value) ? 'choice choice--active' : 'choice'} key={value} onClick={() => togglePet(value)} type="button">
                {label}
              </button>
            ))}
          </div>
          <button className="primary-button primary-button--compact" onClick={handleUpdateProfile} type="button">
            <Check size={17} />
            保存资料
          </button>
        </div>
      </section>

      <section className="settings-section">
        <div>
          <h2>联系方式</h2>
          <p>账号至少保留一种可用联系方式。</p>
        </div>
        <div className="settings-form settings-form--split">
          <IdentityCard title="邮箱" value={emailIdentity?.maskedValue} onUnbind={emailIdentity ? () => handleUnbind('email') : undefined} />
          <IdentityCard title="手机号" value={phoneIdentity?.maskedValue} onUnbind={phoneIdentity ? () => handleUnbind('phone') : undefined} />
          <label className="field">
            <span>新邮箱</span>
            <div className="field__control">
              <Mail size={17} />
              <input value={email} onChange={(event) => setEmail(event.target.value)} placeholder="you@example.com" />
            </div>
          </label>
          <label className="field">
            <span>新手机号</span>
            <div className="field__control">
              <Smartphone size={17} />
              <input value={phone} onChange={(event) => setPhone(event.target.value)} placeholder="+86 手机号" />
            </div>
          </label>
          <label className="field settings-form__full">
            <span>图片验证码</span>
            <div className="captcha-row">
              <div className="field__control">
                <input value={captcha.captchaCode} onChange={(event) => captcha.setCaptchaCode(event.target.value)} placeholder="输入图中字符" />
              </div>
              <button className="captcha-image" disabled={captcha.captchaLoading} onClick={captcha.reloadCaptcha} type="button" aria-label="刷新图片验证码">
                {captcha.captcha ? <img src={captcha.captcha.imageBase64} alt="图片验证码" /> : '刷新'}
              </button>
            </div>
          </label>
          <label className="field">
            <span>邮箱验证码</span>
            <div className="field__control field__control--action">
              <Send size={17} />
              <input value={emailCode} onChange={(event) => setEmailCode(event.target.value)} placeholder="6 位验证码" />
              <button className="inline-action" disabled={countdown.seconds > 0} onClick={() => handleSendIdentityCode('email')} type="button">
                {countdown.seconds > 0 ? `${countdown.seconds}s` : '发送'}
              </button>
            </div>
          </label>
          <label className="field">
            <span>短信验证码</span>
            <div className="field__control field__control--action">
              <Send size={17} />
              <input value={smsCode} onChange={(event) => setSmsCode(event.target.value)} placeholder="6 位验证码" />
              <button className="inline-action" disabled={countdown.seconds > 0} onClick={() => handleSendIdentityCode('phone')} type="button">
                {countdown.seconds > 0 ? `${countdown.seconds}s` : '发送'}
              </button>
            </div>
          </label>
          <button className="button-link" onClick={handleBindEmail} type="button">
            更新邮箱
          </button>
          <button className="button-link" onClick={handleBindPhone} type="button">
            更新手机号
          </button>
        </div>
      </section>

      <section className="settings-section">
        <div>
          <h2>密码</h2>
          <p>修改密码后，其他设备会重新登录。</p>
        </div>
        <div className="settings-form settings-form--split">
          <label className="field">
            <span>当前密码</span>
            <div className="field__control">
              <KeyRound size={17} />
              <input type="password" value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} />
            </div>
          </label>
          <label className="field">
            <span>新密码</span>
            <div className="field__control">
              <KeyRound size={17} />
              <input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} />
            </div>
          </label>
          <button className="primary-button primary-button--compact" onClick={handlePassword} type="button">
            <ShieldCheck size={17} />
            保存密码
          </button>
        </div>
      </section>

      <section className="settings-section">
        <div>
          <h2>登录设备</h2>
          <p>只保留你认识的设备。</p>
        </div>
        <div className="session-list">
          <button className="button-link" onClick={loadSessions} type="button">
            <RefreshCcw size={16} />
            刷新
          </button>
          {sessions.map((session) => (
            <div className="session-item" key={session.id}>
              <div>
                <strong>{session.deviceName}</strong>
                <span>{session.current ? '当前设备' : formatTime(session.lastActiveTime)}</span>
              </div>
              {!session.current ? (
                <button className="icon-button" onClick={() => handleRevoke(session.id)} type="button" aria-label="移除设备">
                  <Trash2 size={17} />
                </button>
              ) : null}
            </div>
          ))}
          <button className="button-link button-link--danger" onClick={handleRevokeAll} type="button">
            退出全部设备
          </button>
        </div>
      </section>

      {passwordDialog ? (
        <div className="modal-backdrop" role="presentation" onMouseDown={closePasswordDialog}>
          <section
            className="confirm-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="password-dialog-title"
            aria-describedby="password-dialog-description"
            onMouseDown={(event) => event.stopPropagation()}
          >
            <form
              className="form-stack"
              onSubmit={(event) => {
                event.preventDefault();
                void handleConfirmPasswordDialog();
              }}
            >
              <div className="confirm-dialog__header">
                <div>
                  <p className="eyebrow">安全确认</p>
                  <h2 id="password-dialog-title">{passwordDialog.title}</h2>
                </div>
                <button className="icon-button" onClick={closePasswordDialog} type="button" aria-label="关闭确认弹窗" disabled={dialogSubmitting}>
                  <X size={17} />
                </button>
              </div>
              <p id="password-dialog-description" className="confirm-dialog__description">
                {passwordDialog.description}
              </p>
              <label className="field">
                <span>当前密码</span>
                <div className="field__control">
                  <KeyRound size={17} />
                  <input
                    type="password"
                    autoComplete="current-password"
                    value={confirmationPassword}
                    onChange={(event) => setConfirmationPassword(event.target.value)}
                    autoFocus
                  />
                </div>
              </label>
              {dialogError ? <p className="form-error">{dialogError}</p> : null}
              <div className="dialog-actions">
                <button className="button-link" onClick={closePasswordDialog} type="button" disabled={dialogSubmitting}>
                  取消
                </button>
                <button
                  className={passwordDialog.danger ? 'primary-button primary-button--danger' : 'primary-button'}
                  disabled={dialogSubmitting || !confirmationPassword}
                  type="submit"
                >
                  {dialogSubmitting ? '处理中' : passwordDialog.confirmLabel}
                </button>
              </div>
            </form>
          </section>
        </div>
      ) : null}
    </section>
  );
}

function IdentityCard({ title, value, onUnbind }: { title: string; value?: string; onUnbind?: () => void }) {
  return (
    <div className="identity-card">
      <span>{title}</span>
      <strong>{value || '未绑定'}</strong>
      {onUnbind ? (
        <button className="inline-action inline-action--danger" onClick={onUnbind} type="button">
          解绑
        </button>
      ) : null}
    </div>
  );
}

function errorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return '操作失败，请稍后再试';
}

function formatTime(value: number) {
  if (!value) {
    return '暂无活跃记录';
  }
  return new Date(value * 1000).toLocaleString();
}
