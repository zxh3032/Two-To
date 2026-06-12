import { NavLink, Outlet } from 'react-router-dom';

const navItems = [
  { to: '/', label: '首页', end: true },
  { to: '/assessment', label: '适配测评' },
  { to: '/breeds', label: '品种解析' },
  { to: '/pets', label: '宠物档案' },
];

// AppLayout 提供全局导航和基础页面框架，页面内容通过 Outlet 注入。
// 请求参数：无；返回值：包含主导航和当前路由内容的 React 节点。
export function AppLayout() {
  return (
    <div className="app-shell">
      <aside className="sidebar" aria-label="主导航">
        <div className="brand">
          <img className="brand__mark" src="/two-to-mark.svg" alt="" />
          <div>
            <strong>Two-To</strong>
            <span>宠物陪伴服务</span>
          </div>
        </div>
        <nav className="nav-list">
          {navItems.map((item) => (
            <NavLink
              className={({ isActive }) => (isActive ? 'nav-link nav-link--active' : 'nav-link')}
              end={item.end}
              key={item.to}
              to={item.to}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <main className="main-panel">
        <Outlet />
      </main>
    </div>
  );
}
