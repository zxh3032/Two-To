import { Link } from 'react-router-dom';

export function NotFoundPage() {
  return (
    <section className="page-stack">
      <div className="empty-panel">
        <h1>页面不存在</h1>
        <p>当前地址没有匹配到 Two-To 页面。</p>
        <Link className="button-link" to="/">
          返回控制台
        </Link>
      </div>
    </section>
  );
}
