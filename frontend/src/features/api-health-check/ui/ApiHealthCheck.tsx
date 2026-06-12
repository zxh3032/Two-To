import { useQuery } from '@tanstack/react-query';

import { fetchPing } from '../../../shared/api/system';

// ApiHealthCheck 用于必要时确认前后端是否可用，但不向用户暴露 requestId 等排查字段。
// 请求参数：无；返回值：面向用户的简化服务状态 React 节点。
export function ApiHealthCheck() {
  const { data, error, isLoading, refetch } = useQuery({
    queryKey: ['system', 'ping'],
    queryFn: fetchPing,
    refetchInterval: 30_000,
  });

  return (
    <section className="health-panel" aria-label="后端联通状态">
      <div>
        <p className="eyebrow">连接状态</p>
        <h2>基础服务</h2>
      </div>

      {isLoading && <span className="health-state health-state--pending">准备中</span>}
      {error && <span className="health-state health-state--error">暂不可用</span>}
      {data && <span className="health-state health-state--ok">已连接</span>}

      {error && <p className="error-text">服务暂时没有响应，可以稍后再试。</p>}
      <button className="button-link" type="button" onClick={() => void refetch()}>
        重新尝试
      </button>
    </section>
  );
}
