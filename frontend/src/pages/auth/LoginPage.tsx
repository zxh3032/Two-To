import { ArrowRight, Eye, EyeOff, KeyRound, Mail, Smartphone } from 'lucide-react';
import { type KeyboardEvent, useMemo, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { AuthAnimalStage } from '../../features/auth-animal-stage/ui/AuthAnimalStage';
import { CaptchaField, VerificationCodeField } from '../../features/auth-form/AuthFields';
import { useCaptcha } from '../../features/auth-form/useCaptcha';
import { useCountdown } from '../../features/auth-form/useCountdown';
import { useAuth } from '../../app/useAuth';
import { emailLogin, phoneLogin, sendCode } from '../../shared/api/auth';
import { ApiError } from '../../shared/api/http';
import { saveProfileSetupToken } from '../../shared/lib/profile-setup-token';

type LoginTab = 'phone' | 'email';
type AnimalMode = 'idle' | 'account' | 'password' | 'secret';

export function LoginPage() {
  const navigate = useNavigate();
  const auth = useAuth();
  const captcha = useCaptcha();
  const countdown = useCountdown();
  const activeTabRef = useRef<LoginTab>('phone');
  const [tab, setTab] = useState<LoginTab>('phone');
  const [phone, setPhone] = useState('');
  const [smsCode, setSmsCode] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [needEmailCaptcha, setNeedEmailCaptcha] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState('');
  const [focusField, setFocusField] = useState<'account' | 'password' | null>(null);

  const animalMode = useMemo<AnimalMode>(() => {
    if (passwordVisible && focusField === 'password') {
      return 'secret';
    }
    if (focusField === 'password') {
      return 'password';
    }
    if (focusField === 'account') {
      return 'account';
    }
    return 'idle';
  }, [focusField, passwordVisible]);

  function switchTab(nextTab: LoginTab) {
    activeTabRef.current = nextTab;
    setTab(nextTab);
    setFocusField(null);
    setError('');
  }

  function handleTabKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const order: LoginTab[] = ['phone', 'email'];
    const currentIndex = order.indexOf(tab);
    let nextTab: LoginTab | null = null;

    if (event.key === 'ArrowRight' || event.key === 'ArrowDown') {
      nextTab = order[(currentIndex + 1) % order.length];
    }
    if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
      nextTab = order[(currentIndex - 1 + order.length) % order.length];
    }
    if (!nextTab) {
      return;
    }

    event.preventDefault();
    switchTab(nextTab);
    window.setTimeout(() => document.getElementById(`login-tab-${nextTab}`)?.focus(), 0);
  }

  async function handleSendSMS() {
    setError('');
    if (!captcha.captcha?.captchaId || !captcha.captchaCode) {
      setError('请先输入图片验证码');
      return;
    }
    setSending(true);
    try {
      const data = await sendCode({
        codeType: 'sms',
        scene: 'login',
        target: phone,
        captchaId: captcha.captcha.captchaId,
        captchaCode: captcha.captchaCode,
      });
      countdown.start(data.cooldownSeconds);
      await captcha.reloadCaptcha();
    } catch (err) {
      if (activeTabRef.current === 'phone') {
        setError(errorMessage(err));
      }
      await captcha.reloadCaptcha();
    } finally {
      setSending(false);
    }
  }

  async function handlePhoneLogin() {
    setSubmitting(true);
    setError('');
    try {
      const data = await phoneLogin({ phone, smsCode });
      if (data.requiresProfileSetup) {
        saveProfileSetupToken(data.profileSetupToken);
        navigate('/auth/setup-profile');
        return;
      }
      if (data.auth) {
        await auth.loginWithAuth(data.auth);
        navigate('/');
      }
    } catch (err) {
      if (activeTabRef.current === 'phone') {
        setError(errorMessage(err));
      }
    } finally {
      setSubmitting(false);
    }
  }

  async function handleEmailLogin() {
    setSubmitting(true);
    setError('');
    try {
      const data = await emailLogin({
        email,
        password,
        captchaId: needEmailCaptcha ? captcha.captcha?.captchaId : undefined,
        captchaCode: needEmailCaptcha ? captcha.captchaCode : undefined,
      });
      await auth.loginWithAuth(data);
      navigate('/');
    } catch (err) {
      if (activeTabRef.current !== 'email') {
        return;
      }
      if (err instanceof ApiError && err.data && typeof err.data === 'object' && 'captchaRequired' in err.data) {
        setNeedEmailCaptcha(true);
        await captcha.reloadCaptcha();
      }
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="auth-page auth-page--split">
      <AuthAnimalStage mode={animalMode} />
      <section className="auth-panel" aria-labelledby="login-title">
        <div className="auth-panel__header">
          <p className="eyebrow">账号入口</p>
          <h1 id="login-title">登录 Two-To</h1>
          <p>继续完善你的适配测评、宠物档案和照护记录。</p>
        </div>

        <div className="segmented" role="tablist" aria-label="账号方式" onKeyDown={handleTabKeyDown}>
          <button
            id="login-tab-phone"
            role="tab"
            aria-controls="login-panel-phone"
            aria-selected={tab === 'phone'}
            className={tab === 'phone' ? 'segmented__item segmented__item--active' : 'segmented__item'}
            onClick={() => switchTab('phone')}
            tabIndex={tab === 'phone' ? 0 : -1}
            type="button"
          >
            <Smartphone size={16} />
            手机号
          </button>
          <button
            id="login-tab-email"
            role="tab"
            aria-controls="login-panel-email"
            aria-selected={tab === 'email'}
            className={tab === 'email' ? 'segmented__item segmented__item--active' : 'segmented__item'}
            onClick={() => switchTab('email')}
            tabIndex={tab === 'email' ? 0 : -1}
            type="button"
          >
            <Mail size={16} />
            邮箱
          </button>
        </div>

        {tab === 'phone' ? (
          <form
            id="login-panel-phone"
            className="form-stack"
            role="tabpanel"
            aria-labelledby="login-tab-phone"
            onSubmit={(event) => {
              event.preventDefault();
              void handlePhoneLogin();
            }}
          >
            <label className="field">
              <span>手机号</span>
              <div className="field__control">
                <Smartphone size={17} />
                <input
                  onClick={() => setFocusField('account')}
                  onFocus={() => setFocusField('account')}
                  onBlur={() => setFocusField(null)}
                  value={phone}
                  onChange={(event) => {
                    setFocusField('account');
                    setPhone(event.target.value);
                  }}
                  placeholder="+86 手机号"
                />
              </div>
            </label>
            <CaptchaField captchaHook={captcha} />
            <VerificationCodeField countdownSeconds={countdown.seconds} kind="sms" onChange={setSmsCode} onSend={handleSendSMS} sending={sending} value={smsCode} />
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" disabled={submitting} type="submit">
              <ArrowRight size={18} />
              {submitting ? '登录中' : '登录 / 注册'}
            </button>
          </form>
        ) : (
          <form
            id="login-panel-email"
            className="form-stack"
            role="tabpanel"
            aria-labelledby="login-tab-email"
            onSubmit={(event) => {
              event.preventDefault();
              void handleEmailLogin();
            }}
          >
            <label className="field">
              <span>邮箱</span>
              <div className="field__control">
                <Mail size={17} />
                <input
                  onClick={() => setFocusField('account')}
                  onFocus={() => setFocusField('account')}
                  onBlur={() => setFocusField(null)}
                  value={email}
                  onChange={(event) => {
                    setFocusField('account');
                    setEmail(event.target.value);
                  }}
                  placeholder="you@example.com"
                />
              </div>
            </label>
            <label className="field">
              <span>密码</span>
              <div className="field__control">
                <KeyRound size={17} />
                <input
                  onClick={() => setFocusField('password')}
                  onFocus={() => setFocusField('password')}
                  onBlur={() => setFocusField(null)}
                  type={passwordVisible ? 'text' : 'password'}
                  value={password}
                  onChange={(event) => {
                    setFocusField('password');
                    setPassword(event.target.value);
                  }}
                  placeholder="8-64 位密码"
                />
                <button
                  className="icon-button"
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => {
                    setFocusField('password');
                    setPasswordVisible((value) => !value);
                  }}
                  type="button"
                  aria-label={passwordVisible ? '隐藏密码' : '显示密码'}
                >
                  {passwordVisible ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </label>
            {needEmailCaptcha ? <CaptchaField captchaHook={captcha} /> : null}
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" disabled={submitting} type="submit">
              <ArrowRight size={18} />
              {submitting ? '登录中' : '登录'}
            </button>
          </form>
        )}

        <div className="auth-links">
          <Link to="/auth/register/email">注册邮箱账号</Link>
          <Link to="/auth/forgot-password">忘记密码</Link>
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
