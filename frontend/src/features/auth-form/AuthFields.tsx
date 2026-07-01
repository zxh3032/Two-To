import { Mail, MessageCircle } from 'lucide-react';

import { useCaptcha } from './useCaptcha';

interface CaptchaFieldProps {
  captchaHook: ReturnType<typeof useCaptcha>;
}

interface VerificationCodeFieldProps {
  countdownSeconds: number;
  kind: 'email' | 'sms';
  onChange: (value: string) => void;
  onSend: () => void | Promise<void>;
  sending: boolean;
  value: string;
}

export function CaptchaField({ captchaHook }: CaptchaFieldProps) {
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

export function VerificationCodeField({ countdownSeconds, kind, onChange, onSend, sending, value }: VerificationCodeFieldProps) {
  const Icon = kind === 'email' ? Mail : MessageCircle;
  const label = kind === 'email' ? '邮箱验证码' : '短信验证码';

  return (
    <label className="field">
      <span>{label}</span>
      <div className="field__control field__control--action">
        <Icon size={17} />
        <input value={value} onChange={(event) => onChange(event.target.value)} placeholder="6 位验证码" inputMode="numeric" />
        <button className="inline-action" disabled={sending || countdownSeconds > 0} onClick={onSend} type="button">
          {countdownSeconds > 0 ? `${countdownSeconds}s` : '发送'}
        </button>
      </div>
    </label>
  );
}
