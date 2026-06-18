import { Eye, EyeOff, KeyRound, Mail, MessageCircle, PawPrint, Smartphone } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { AuthAnimalStage } from '../../features/auth-animal-stage/ui/AuthAnimalStage';
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
      setError(errorMessage(err));
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
      setError(errorMessage(err));
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

        <div className="segmented" role="tablist" aria-label="登录方式">
          <button
            className={tab === 'phone' ? 'segmented__item segmented__item--active' : 'segmented__item'}
            onClick={() => {
              setTab('phone');
              setFocusField(null);
            }}
            type="button"
          >
            <Smartphone size={16} />
            手机号
          </button>
          <button
            className={tab === 'email' ? 'segmented__item segmented__item--active' : 'segmented__item'}
            onClick={() => {
              setTab('email');
              setFocusField(null);
            }}
            type="button"
          >
            <Mail size={16} />
            邮箱
          </button>
        </div>

        {tab === 'phone' ? (
          <form className="form-stack" onSubmit={(event) => event.preventDefault()}>
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
            <label className="field">
              <span>短信验证码</span>
              <div className="field__control field__control--action">
                <MessageCircle size={17} />
                <input value={smsCode} onChange={(event) => setSmsCode(event.target.value)} placeholder="6 位验证码" inputMode="numeric" />
                <button className="inline-action" disabled={sending || countdown.seconds > 0} onClick={handleSendSMS} type="button">
                  {countdown.seconds > 0 ? `${countdown.seconds}s` : '发送'}
                </button>
              </div>
            </label>
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button" disabled={submitting} onClick={handlePhoneLogin} type="button">
              <PawPrint size={18} />
              {submitting ? '登录中' : '登录 / 注册'}
            </button>
          </form>
        ) : (
          <form className="form-stack" onSubmit={(event) => event.preventDefault()}>
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
                  placeholder="输入密码"
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
            <button className="primary-button" disabled={submitting} onClick={handleEmailLogin} type="button">
              <PawPrint size={18} />
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

function CaptchaField({ captchaHook }: { captchaHook: ReturnType<typeof useCaptcha> }) {
  return (
    <label className="field">
      <span>图片验证码</span>
      <div className="captcha-row">
        <div className="field__control">
          <input value={captchaHook.captchaCode} onChange={(event) => captchaHook.setCaptchaCode(event.target.value)} placeholder="输入图中字符" />
        </div>
        <button className="captcha-image" disabled={captchaHook.captchaLoading} onClick={captchaHook.reloadCaptcha} type="button" aria-label="刷新图片验证码">
          {captchaHook.captcha ? <img src={captchaHook.captcha.imageBase64} alt="图片验证码" /> : '刷新'}
        </button>
      </div>
    </label>
  );
}

function errorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return '操作失败，请稍后再试';
}
