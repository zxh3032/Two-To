import { KeyRound, Mail, Smartphone } from 'lucide-react';
import { type KeyboardEvent, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { AuthAnimalStage } from '../../features/auth-animal-stage/ui/AuthAnimalStage';
import { CaptchaField, VerificationCodeField } from '../../features/auth-form/AuthFields';
import { useCaptcha } from '../../features/auth-form/useCaptcha';
import { useCountdown } from '../../features/auth-form/useCountdown';
import { resetForgotPassword, sendCode, verifyForgotPasswordCode } from '../../shared/api/auth';

export function ForgotPasswordPage() {
  const navigate = useNavigate();
  const captcha = useCaptcha();
  const countdown = useCountdown();
  const activeIdentityTypeRef = useRef<'email' | 'phone'>('phone');
  const [identityType, setIdentityType] = useState<'email' | 'phone'>('phone');
  const [identityValue, setIdentityValue] = useState('');
  const [code, setCode] = useState('');
  const [resetToken, setResetToken] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [sending, setSending] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  async function handleSendCode() {
    const requestIdentityType = identityType;
    setError('');
    if (!captcha.captcha || !captcha.captchaCode) {
      setError('请先输入图片验证码');
      return;
    }
    setSending(true);
    try {
      const data = await sendCode({
        codeType: identityType === 'email' ? 'email' : 'sms',
        scene: 'forgot_password',
        target: identityValue,
        captchaId: captcha.captcha.captchaId,
        captchaCode: captcha.captchaCode,
      });
      if (activeIdentityTypeRef.current === requestIdentityType) {
        countdown.start(data.cooldownSeconds);
        await captcha.reloadCaptcha();
      }
    } catch (err) {
      if (activeIdentityTypeRef.current === requestIdentityType) {
        setError(errorMessage(err));
        await captcha.reloadCaptcha();
      }
    } finally {
      setSending(false);
    }
  }

  async function handleVerifyCode() {
    const requestIdentityType = identityType;
    setSubmitting(true);
    setError('');
    try {
      const data = await verifyForgotPasswordCode({ identityType, identityValue, code });
      if (activeIdentityTypeRef.current !== requestIdentityType) {
        return;
      }
      if (!data.resetToken) {
        setError('验证码已处理。若账号存在，请重新发起找回密码。');
        return;
      }
      setResetToken(data.resetToken);
    } catch (err) {
      if (activeIdentityTypeRef.current === requestIdentityType) {
        setError(errorMessage(err));
      }
    } finally {
      setSubmitting(false);
    }
  }

  async function handleReset() {
    setSubmitting(true);
    setError('');
    try {
      await resetForgotPassword({ resetToken, newPassword });
      navigate('/auth/login', { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  function switchIdentityType(nextType: 'email' | 'phone') {
    if (nextType === identityType) {
      return;
    }
    activeIdentityTypeRef.current = nextType;
    setIdentityType(nextType);
    setIdentityValue('');
    setCode('');
    setSending(false);
    setError('');
    countdown.reset();
    captcha.setCaptchaCode('');
  }

  function handleIdentityTabKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const order: Array<'email' | 'phone'> = ['phone', 'email'];
    const currentIndex = order.indexOf(identityType);
    let nextType: 'email' | 'phone' | null = null;

    if (event.key === 'ArrowRight' || event.key === 'ArrowDown') {
      nextType = order[(currentIndex + 1) % order.length];
    }
    if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
      nextType = order[(currentIndex - 1 + order.length) % order.length];
    }
    if (!nextType) {
      return;
    }

    event.preventDefault();
    switchIdentityType(nextType);
    window.setTimeout(() => document.getElementById(`forgot-tab-${nextType}`)?.focus(), 0);
  }

  return (
    <main className="auth-page auth-page--split">
      <AuthAnimalStage mode={resetToken ? 'password' : 'account'} />
      <section className="auth-panel" aria-labelledby="forgot-title">
        <div className="auth-panel__header">
          <p className="eyebrow">找回密码</p>
          <h1 id="forgot-title">重置登录密码</h1>
          <p>通过已绑定手机号或邮箱校验后设置新密码。</p>
        </div>

        {!resetToken ? (
          <form
            className="form-stack"
            onSubmit={(event) => {
              event.preventDefault();
              void handleVerifyCode();
            }}
          >
            <div className="segmented" role="tablist" aria-label="账号方式" onKeyDown={handleIdentityTabKeyDown}>
              <button
                id="forgot-tab-phone"
                role="tab"
                aria-controls="forgot-panel"
                aria-selected={identityType === 'phone'}
                className={identityType === 'phone' ? 'segmented__item segmented__item--active' : 'segmented__item'}
                onClick={() => switchIdentityType('phone')}
                tabIndex={identityType === 'phone' ? 0 : -1}
                type="button"
              >
                <Smartphone size={16} />
                手机号
              </button>
              <button
                id="forgot-tab-email"
                role="tab"
                aria-controls="forgot-panel"
                aria-selected={identityType === 'email'}
                className={identityType === 'email' ? 'segmented__item segmented__item--active' : 'segmented__item'}
                onClick={() => switchIdentityType('email')}
                tabIndex={identityType === 'email' ? 0 : -1}
                type="button"
              >
                <Mail size={16} />
                邮箱
              </button>
            </div>
            <div id="forgot-panel" className="form-stack" role="tabpanel" aria-labelledby={`forgot-tab-${identityType}`}>
              <label className="field">
                <span>{identityType === 'email' ? '邮箱' : '手机号'}</span>
                <div className="field__control">
                  {identityType === 'email' ? <Mail size={17} /> : <Smartphone size={17} />}
                  <input value={identityValue} onChange={(event) => setIdentityValue(event.target.value)} placeholder={identityType === 'email' ? 'you@example.com' : '+86 手机号'} />
                </div>
              </label>
              <CaptchaField captchaHook={captcha} />
              <VerificationCodeField countdownSeconds={countdown.seconds} kind={identityType === 'email' ? 'email' : 'sms'} onChange={setCode} onSend={handleSendCode} sending={sending} value={code} />
            </div>
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" disabled={submitting} type="submit">
              下一步
            </button>
          </form>
        ) : (
          <form
            className="form-stack"
            onSubmit={(event) => {
              event.preventDefault();
              void handleReset();
            }}
          >
            <label className="field">
              <span>新密码</span>
              <div className="field__control">
                <KeyRound size={17} />
                <input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} placeholder="8-64 位密码" />
              </div>
            </label>
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" disabled={submitting} type="submit">
              保存新密码
            </button>
          </form>
        )}

        <div className="auth-links">
          <Link to="/auth/login">返回登录</Link>
        </div>
      </section>
    </main>
  );
}

function errorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return '操作失败，请稍后再试';
}
