import { cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { App } from './App';
import { router } from './router';

describe('App', () => {
  beforeEach(async () => {
    installLocalStorageMock();
    window.history.pushState({}, '', '/');
    await router.navigate('/');
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.endsWith('/api/v1/auth/captcha')) {
          return jsonResponse({
            captchaId: 'captcha-test',
            imageBase64: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="142" height="48"></svg>',
            expiresIn: 120,
          });
        }

        if (url.endsWith('/api/v1/slogan')) {
          return jsonResponse({
            slogan: '两两相逢，奔赴朝夕。',
            requestId: 'test-slogan-request-id',
          });
        }

        if (url.endsWith('/api/v1/auth/email-login')) {
          return jsonResponse({
            accessToken: 'access-test',
            accessTokenExpiresIn: 3600,
            refreshToken: 'refresh-test',
            refreshTokenExpiresIn: 86400,
            user: { id: 1, nickname: '测试用户', status: 1 },
          });
        }

        if (url.endsWith('/api/v1/account/me')) {
          return jsonResponse(accountMe());
        }

        if (url.endsWith('/api/v1/account/sessions')) {
          return jsonResponse({
            currentSessionId: 1,
            sessions: [
              { id: 1, deviceName: '当前 Chrome', status: 1, lastActiveTime: 1782864000, expireTime: 1782950400, current: true },
              { id: 2, deviceName: 'iPhone Safari', status: 1, lastActiveTime: 1782777600, expireTime: 1782950400, current: false },
            ],
          });
        }

        if (url.endsWith('/api/v1/account/security/email')) {
          return jsonResponse({});
        }

        return jsonResponse({
          message: 'pong',
          service: 'two-to-api',
          environment: 'test',
          requestId: 'test-request-id',
          timestamp: 1,
        });
      }),
    );
  });

  afterEach(() => {
    cleanup();
    window.localStorage.clear();
    window.history.pushState({}, '', '/');
    vi.unstubAllGlobals();
  });

  it('未登录时进入登录页', async () => {
    render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '登录 Two-To' })).toBeInTheDocument(), {
      timeout: 2_000,
    });
    expect(screen.getByText('先判断，')).toBeInTheDocument();
    expect(screen.getByText('再开始认真陪伴。')).toBeInTheDocument();
    expect(screen.getByRole('tablist', { name: '账号方式' })).toBeInTheDocument();
    expect(screen.getByText('登录 / 注册')).toBeInTheDocument();
    expect(screen.queryByText('从养前评估、品种选择到长期照护，让每一次陪伴都有清晰依据和持续记录。')).not.toBeInTheDocument();
  });

  it('邮箱登录支持 tab 键盘切换和 Enter 提交', async () => {
    const user = userEvent.setup();
    render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '登录 Two-To' })).toBeInTheDocument(), {
      timeout: 2_000,
    });

    screen.getByRole('tab', { name: '手机号' }).focus();
    await user.keyboard('{ArrowRight}');
    expect(screen.getByRole('tab', { name: '邮箱' })).toHaveAttribute('aria-selected', 'true');

    await user.type(screen.getByPlaceholderText('you@example.com'), 'user@example.com');
    await user.type(screen.getByPlaceholderText('8-64 位密码'), 'password123{Enter}');

    await waitFor(() => expect(screen.getByRole('heading', { name: '从想养，到养好' })).toBeInTheDocument());
  });

  it('找回密码默认使用手机号并保持和登录一致的顺序', async () => {
    const user = userEvent.setup();
    render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '登录 Two-To' })).toBeInTheDocument(), {
      timeout: 2_000,
    });
    await user.click(screen.getByRole('link', { name: '忘记密码' }));

    await waitFor(() => expect(screen.getByRole('heading', { name: '重置登录密码' })).toBeInTheDocument());
    const tabs = screen.getAllByRole('tab');
    expect(tabs.map((tab) => tab.textContent)).toEqual(['手机号', '邮箱']);
    expect(screen.getByRole('tab', { name: '手机号' })).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByPlaceholderText('+86 手机号')).toBeInTheDocument();
    expect(screen.getByText('短信验证码')).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: '邮箱' }));
    expect(screen.getByPlaceholderText('you@example.com')).toBeInTheDocument();
    expect(screen.getByText('邮箱验证码')).toBeInTheDocument();
  });

  it('切换登录方式时清理不属于当前方式的错误提示', async () => {
    const user = userEvent.setup();
    render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '登录 Two-To' })).toBeInTheDocument(), {
      timeout: 2_000,
    });
    await user.click(screen.getByRole('button', { name: '发送' }));

    expect(screen.getByText('请先输入图片验证码')).toBeInTheDocument();
    await user.click(screen.getByRole('tab', { name: '邮箱' }));

    expect(screen.queryByText('请先输入图片验证码')).not.toBeInTheDocument();
    expect(screen.queryByText('图片验证码')).not.toBeInTheDocument();
    expect(screen.getByPlaceholderText('8-64 位密码')).toBeInTheDocument();
  });

  it('切换找回密码账号方式时清理验证码错误和已输入身份', async () => {
    const user = userEvent.setup();
    render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '登录 Two-To' })).toBeInTheDocument(), {
      timeout: 2_000,
    });
    await user.click(screen.getByRole('link', { name: '忘记密码' }));

    await waitFor(() => expect(screen.getByRole('heading', { name: '重置登录密码' })).toBeInTheDocument());
    await user.type(screen.getByPlaceholderText('+86 手机号'), '13800000000');
    await user.click(screen.getByRole('button', { name: '发送' }));

    expect(screen.getByText('请先输入图片验证码')).toBeInTheDocument();
    await user.click(screen.getByRole('tab', { name: '邮箱' }));

    expect(screen.queryByText('请先输入图片验证码')).not.toBeInTheDocument();
    expect(screen.getByPlaceholderText('you@example.com')).toHaveValue('');
    expect(screen.getByText('邮箱验证码')).toBeInTheDocument();
  });

  it('账号安全敏感操作使用站内密码确认弹窗', async () => {
    const user = userEvent.setup();
    window.localStorage.setItem(
      'two-to.auth.tokens',
      JSON.stringify({
        accessToken: 'access-test',
        refreshToken: 'refresh-test',
        accessTokenExpiresAt: Date.now() + 3600_000,
        refreshTokenExpiresAt: Date.now() + 86400_000,
      }),
    );
    render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '从想养，到养好' })).toBeInTheDocument());
    await user.click(screen.getByRole('link', { name: /账号安全/ }));
    await waitFor(() => expect(screen.getByRole('heading', { name: '资料、联系方式和设备' })).toBeInTheDocument());
    await user.click(screen.getAllByRole('button', { name: '解绑' })[0]);

    const dialog = screen.getByRole('dialog', { name: '解绑邮箱' });
    expect(dialog).toBeInTheDocument();
    await user.type(within(dialog).getByLabelText('当前密码'), 'password123');
    await user.click(within(dialog).getByRole('button', { name: '确认解绑' }));

    await waitFor(() => expect(screen.getByText('邮箱已解绑')).toBeInTheDocument());
  });
});

function jsonResponse(data: unknown) {
  return new Response(
    JSON.stringify({
      code: 0,
      message: 'success',
      data,
      requestId: 'test-request-id',
    }),
    {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    },
  );
}

function accountMe() {
  return {
    user: { id: 1, nickname: '测试用户', status: 1 },
    profile: {
      petStage: 'planning',
      interestedPetTypes: ['dog', 'cat'],
      petExperience: 'newbie',
      dailyCompanyTime: '1_3h',
      livingSituation: 'alone',
      petConstraints: [],
    },
    identities: [
      { id: 10, identityType: 'email', identityValue: 'user@example.com', maskedValue: 'u***@example.com', verifyTime: 1782864000 },
      { id: 11, identityType: 'phone', identityValue: '+8613800000000', maskedValue: '+86 138****0000', verifyTime: 1782864000 },
    ],
  };
}

function installLocalStorageMock() {
  const store = new Map<string, string>();
  Object.defineProperty(window, 'localStorage', {
    configurable: true,
    value: {
      getItem: (key: string) => store.get(key) ?? null,
      setItem: (key: string, value: string) => {
        store.set(key, value);
      },
      removeItem: (key: string) => {
        store.delete(key);
      },
      clear: () => store.clear(),
      key: (index: number) => Array.from(store.keys())[index] ?? null,
      get length() {
        return store.size;
      },
    },
  });
}
