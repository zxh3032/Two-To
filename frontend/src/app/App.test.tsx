import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { App } from './App';

describe('App', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.endsWith('/api/v1/slogan')) {
          return new Response(
            JSON.stringify({
              code: 0,
              message: 'success',
              data: {
                slogan: '两两相逢，奔赴朝夕。',
                requestId: 'test-slogan-request-id',
              },
              requestId: 'test-slogan-request-id',
            }),
            {
              status: 200,
              headers: { 'Content-Type': 'application/json' },
            },
          );
        }

        return new Response(
          JSON.stringify({
            code: 0,
            message: 'success',
            data: {
              message: 'pong',
              service: 'two-to-api',
              environment: 'test',
              requestId: 'test-request-id',
              timestamp: 1,
            },
            requestId: 'test-request-id',
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        );
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('未登录时进入登录页', async () => {
    render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '登录 Two-To' })).toBeInTheDocument(), {
      timeout: 2_000,
    });
    expect(screen.getByText('两两相逢，先认出彼此')).toBeInTheDocument();
    expect(screen.getByRole('tablist', { name: '登录方式' })).toBeInTheDocument();
    expect(screen.getByText('登录 / 注册')).toBeInTheDocument();
    expect(screen.queryByText('从养前评估、品种选择到长期照护，让每一次陪伴都有清晰依据和持续记录。')).not.toBeInTheDocument();
  });
});
