import { NavLink, Outlet } from 'react-router-dom';
import { HeartHandshake, Home, LogOut, PawPrint, ShieldCheck, Sparkles } from 'lucide-react';

import { useAuth } from './useAuth';

const navItems = [
  { to: '/', label: '首页', end: true, icon: Home },
  { to: '/assessment', label: '适配测评', icon: HeartHandshake },
  { to: '/breeds', label: '品种解析', icon: Sparkles },
  { to: '/pets', label: '宠物档案', icon: PawPrint },
  { to: '/account/security', label: '账号安全', icon: ShieldCheck },
];

// AppLayout 提供全局导航和基础页面框架，页面内容通过 Outlet 注入。
// 请求参数：无；返回值：包含主导航和当前路由内容的 React 节点。
export function AppLayout() {
  const auth = useAuth();

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
              <item.icon size={17} />
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="sidebar__account">
          <div>
            <strong>{auth.user?.nickname ?? 'Two-To 用户'}</strong>
            <span>{auth.identities[0]?.maskedValue ?? '已登录'}</span>
          </div>
          <button className="icon-button" onClick={auth.logout} type="button" aria-label="退出登录">
            <LogOut size={18} />
          </button>
        </div>
      </aside>

      <main className="main-panel">
        <Outlet />
      </main>
    </div>
  );
}
