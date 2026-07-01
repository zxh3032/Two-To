import { ArrowRight, KeyRound, Mail } from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { AuthAnimalStage } from '../../features/auth-animal-stage/ui/AuthAnimalStage';
import { CaptchaField, VerificationCodeField } from '../../features/auth-form/AuthFields';
import { useCaptcha } from '../../features/auth-form/useCaptcha';
import { useCountdown } from '../../features/auth-form/useCountdown';
import { sendCode, verifyEmailRegister } from '../../shared/api/auth';
import { saveProfileSetupToken } from '../../shared/lib/profile-setup-token';

export function EmailRegisterPage() {
  const navigate = useNavigate();
  const captcha = useCaptcha();
  const countdown = useCountdown();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [emailCode, setEmailCode] = useState('');
  const [sending, setSending] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [focus, setFocus] = useState<'account' | 'password' | null>(null);

  async function handleSendCode() {
    setError('');
    if (!captcha.captcha || !captcha.captchaCode) {
      setError('请先输入图片验证码');
      return;
    }
    setSending(true);
    try {
      const data = await sendCode({
        codeType: 'email',
        scene: 'register',
        target: email,
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

  async function handleSubmit() {
    setSubmitting(true);
    setError('');
    try {
      const data = await verifyEmailRegister({ email, password, emailCode });
      saveProfileSetupToken(data.profileSetupToken);
      navigate('/auth/setup-profile');
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="auth-page auth-page--split">
      <AuthAnimalStage mode={focus === 'password' ? 'password' : focus === 'account' ? 'account' : 'idle'} />
      <section className="auth-panel" aria-labelledby="register-title">
        <div className="auth-panel__header">
          <p className="eyebrow">邮箱注册</p>
          <h1 id="register-title">创建 Two-To 账号</h1>
          <p>先验证邮箱，下一步补齐基础资料。</p>
        </div>

        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault();
            void handleSubmit();
          }}
        >
          <label className="field">
            <span>邮箱</span>
            <div className="field__control">
              <Mail size={17} />
              <input onFocus={() => setFocus('account')} onBlur={() => setFocus(null)} value={email} onChange={(event) => setEmail(event.target.value)} placeholder="you@example.com" />
            </div>
          </label>
          <label className="field">
            <span>密码</span>
            <div className="field__control">
              <KeyRound size={17} />
              <input onFocus={() => setFocus('password')} onBlur={() => setFocus(null)} type="password" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="8-64 位密码" />
            </div>
          </label>
          <CaptchaField captchaHook={captcha} />
          <VerificationCodeField countdownSeconds={countdown.seconds} kind="email" onChange={setEmailCode} onSend={handleSendCode} sending={sending} value={emailCode} />
          {error ? <p className="form-error">{error}</p> : null}
          <button className="primary-button" disabled={submitting} type="submit">
            <ArrowRight size={18} />
            {submitting ? '校验中' : '下一步'}
          </button>
        </form>

        <div className="auth-links">
          <Link to="/auth/login">已有账号，去登录</Link>
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
