import { useCallback, useEffect, useState } from 'react';

import { getCaptcha, type CaptchaResponse } from '../../shared/api/auth';

export function useCaptcha(autoLoad = true) {
  const [captcha, setCaptcha] = useState<CaptchaResponse | null>(null);
  const [captchaCode, setCaptchaCode] = useState('');
  const [loading, setLoading] = useState(false);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getCaptcha();
      setCaptcha(data);
      setCaptchaCode('');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (autoLoad) {
      const timer = window.setTimeout(() => void reload(), 0);
      return () => window.clearTimeout(timer);
    }
    return undefined;
  }, [autoLoad, reload]);

  return {
    captcha,
    captchaCode,
    setCaptchaCode,
    captchaLoading: loading,
    reloadCaptcha: reload,
  };
}
