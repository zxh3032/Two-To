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

  it('渲染首页并展示流式项目口号', async () => {
    const { container } = render(<App />);

    await waitFor(() => expect(screen.getByRole('heading', { name: '两两相逢，奔赴朝夕。' })).toBeInTheDocument(), {
      timeout: 2_000,
    });
    expect(container.querySelector('.slogan-hero .eyebrow')).toBeNull();
    expect(screen.getByText('从养前评估、品种选择到长期照护，让每一次陪伴都有清晰依据和持续记录。')).toBeInTheDocument();
    expect(screen.queryByText('把相遇，照顾成日常')).not.toBeInTheDocument();
    expect(screen.queryByText('先判断适不适合养，再理解适合养什么，最后把长期照护和成长记录串起来。')).not.toBeInTheDocument();
    expect(screen.queryByText(/Request ID/i)).not.toBeInTheDocument();
    expect(screen.queryByText('pong')).not.toBeInTheDocument();
  });
});
