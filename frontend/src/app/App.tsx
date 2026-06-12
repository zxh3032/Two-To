import { RouterProvider } from 'react-router-dom';

import { AppProviders } from './providers';
import { router } from './router';

// App 只负责装配全局 Provider 和路由，不承载具体业务逻辑。
export function App() {
  return (
    <AppProviders>
      <RouterProvider router={router} />
    </AppProviders>
  );
}
