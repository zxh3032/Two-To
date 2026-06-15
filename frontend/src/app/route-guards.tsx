import { Navigate, Outlet, useLocation } from 'react-router-dom';

import { useAuth } from './useAuth';

export function RequireAuth() {
  const auth = useAuth();
  const location = useLocation();

  if (auth.loading) {
    return <div className="boot-screen">正在确认登录状态</div>;
  }

  if (!auth.authenticated) {
    return <Navigate replace state={{ from: location }} to="/auth/login" />;
  }

  return <Outlet />;
}

export function PublicOnlyRoute() {
  const auth = useAuth();

  if (auth.loading) {
    return <div className="boot-screen">正在确认登录状态</div>;
  }

  if (auth.authenticated) {
    return <Navigate replace to="/" />;
  }

  return <Outlet />;
}
