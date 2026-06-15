import { KeyRound, Mail, Send, Smartphone } from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { AuthAnimalStage } from '../../features/auth-animal-stage/ui/AuthAnimalStage';
import { useCaptcha } from '../../features/auth-form/useCaptcha';
import { useCountdown } from '../../features/auth-form/useCountdown';
import { resetForgotPassword, sendCode, verifyForgotPasswordCode } from '../../shared/api/auth';

export function ForgotPasswordPage() {
  const navigate = useNavigate();
  const captcha = useCaptcha();
  const countdown = useCountdown();
  const [identityType, setIdentityType] = useState<'email' | 'phone'>('email');
  const [identityValue, setIdentityValue] = useState('');
  const [code, setCode] = useState('');
  const [resetToken, setResetToken] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [sending, setSending] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  async function handleSendCode() {
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
      countdown.start(data.cooldownSeconds);
      await captcha.reloadCaptcha();
    } catch (err) {
      setError(errorMessage(err));
      await captcha.reloadCaptcha();
    } finally {
      setSending(false);
    }
  }

  async function handleVerifyCode() {
    setSubmitting(true);
    setError('');
    try {
      const data = await verifyForgotPasswordCode({ identityType, identityValue, code });
      if (!data.resetToken) {
        setError('验证码已处理。若账号存在，请重新发起找回密码。');
        return;
      }
      setResetToken(data.resetToken);
    } catch (err) {
      setError(errorMessage(err));
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

  return (
    <main className="auth-page">
      <AuthAnimalStage mode={resetToken ? 'password' : 'account'} />
      <section className="auth-panel" aria-labelledby="forgot-title">
        <div className="auth-panel__header">
          <p className="eyebrow">找回密码</p>
          <h1 id="forgot-title">重置登录密码</h1>
          <p>通过已绑定邮箱或手机号校验后设置新密码。</p>
        </div>

        {!resetToken ? (
          <form className="form-stack" onSubmit={(event) => event.preventDefault()}>
            <div className="segmented" role="tablist" aria-label="账号类型">
              <button className={identityType === 'email' ? 'segmented__item segmented__item--active' : 'segmented__item'} onClick={() => setIdentityType('email')} type="button">
                <Mail size={16} />
                邮箱
              </button>
              <button className={identityType === 'phone' ? 'segmented__item segmented__item--active' : 'segmented__item'} onClick={() => setIdentityType('phone')} type="button">
                <Smartphone size={16} />
                手机号
              </button>
            </div>
            <label className="field">
              <span>{identityType === 'email' ? '邮箱' : '手机号'}</span>
              <div className="field__control">
                {identityType === 'email' ? <Mail size={17} /> : <Smartphone size={17} />}
                <input value={identityValue} onChange={(event) => setIdentityValue(event.target.value)} placeholder={identityType === 'email' ? 'you@example.com' : '+86 手机号'} />
              </div>
            </label>
            <label className="field">
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
              <span>验证码</span>
              <div className="field__control field__control--action">
                <Send size={17} />
                <input value={code} onChange={(event) => setCode(event.target.value)} placeholder="6 位验证码" inputMode="numeric" />
                <button className="inline-action" disabled={sending || countdown.seconds > 0} onClick={handleSendCode} type="button">
                  {countdown.seconds > 0 ? `${countdown.seconds}s` : '发送'}
                </button>
              </div>
            </label>
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" disabled={submitting} onClick={handleVerifyCode} type="button">
              下一步
            </button>
          </form>
        ) : (
          <form className="form-stack" onSubmit={(event) => event.preventDefault()}>
            <label className="field">
              <span>新密码</span>
              <div className="field__control">
                <KeyRound size={17} />
                <input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} placeholder="8-64 位密码" />
              </div>
            </label>
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" disabled={submitting} onClick={handleReset} type="button">
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
