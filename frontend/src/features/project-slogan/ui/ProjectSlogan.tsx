import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { fetchSlogan } from '../../../shared/api/system';

const STREAM_INTERVAL_MS = 110;

// ProjectSlogan 从后端读取项目口号，并用逐字流式效果作为首页首屏内容展示。
// 请求参数：无；返回值：包含加载、失败、成功三种状态的 React 节点。
export function ProjectSlogan() {
  const [visibleSlogan, setVisibleSlogan] = useState('');
  const { data, error, isLoading } = useQuery({
    queryKey: ['system', 'slogan'],
    queryFn: fetchSlogan,
  });
  const slogan = data?.slogan ?? '';
  const isStreaming = Boolean(slogan && visibleSlogan.length < slogan.length);

  useEffect(() => {
    if (!slogan) {
      return;
    }

    // 系统开启减少动态效果时，直接展示完整口号，避免动画造成阅读负担。
    const shouldReduceMotion =
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    if (shouldReduceMotion) {
      const immediateTimer = window.setTimeout(() => setVisibleSlogan(slogan), 0);
      return () => window.clearTimeout(immediateTimer);
    }

    let nextLength = 0;
    let streamTimer: number | undefined;

    const resetTimer = window.setTimeout(() => {
      setVisibleSlogan('');

      streamTimer = window.setInterval(() => {
        nextLength += 1;
        setVisibleSlogan(slogan.slice(0, nextLength));

        if (nextLength >= slogan.length && streamTimer) {
          window.clearInterval(streamTimer);
        }
      }, STREAM_INTERVAL_MS);
    }, 0);

    return () => {
      window.clearTimeout(resetTimer);
      if (streamTimer) {
        window.clearInterval(streamTimer);
      }
    };
  }, [slogan]);

  return (
    <section className="slogan-hero" aria-label="项目口号">
      <div className="slogan-hero__content">
        {isLoading && <h1 className="slogan-hero__text">正在准备一句温暖的开场...</h1>}
        {error && (
          <h1 className="slogan-hero__text slogan-hero__text--error">两两相逢，奔赴朝夕。</h1>
        )}
        {data && (
          <h1 className="slogan-hero__text" aria-live="polite">
            {visibleSlogan}
            {isStreaming && <span className="slogan-hero__cursor" aria-hidden="true" />}
          </h1>
        )}
        <p className="slogan-hero__subcopy">
          不只是宠物工具箱，而是一条从判断适不适合养、选择适合养什么，到长期照顾好的完整路径。
        </p>
      </div>
    </section>
  );
}
