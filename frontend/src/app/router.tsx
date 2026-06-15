import { Navigate, createBrowserRouter } from 'react-router-dom';

import { AppLayout } from './AppLayout';
import { PublicOnlyRoute, RequireAuth } from './route-guards';
import { SecurityPage } from '../pages/account/SecurityPage';
import { AssessmentPage } from '../pages/assessment/AssessmentPage';
import { EmailRegisterPage } from '../pages/auth/EmailRegisterPage';
import { ForgotPasswordPage } from '../pages/auth/ForgotPasswordPage';
import { LoginPage } from '../pages/auth/LoginPage';
import { SetupProfilePage } from '../pages/auth/SetupProfilePage';
import { BreedsPage } from '../pages/breeds/BreedsPage';
import { HomePage } from '../pages/home/HomePage';
import { NotFoundPage } from '../pages/not-found/NotFoundPage';
import { PetsPage } from '../pages/pets/PetsPage';

// router 只描述页面路由关系，具体业务动作由 pages/features 承接。
export const router = createBrowserRouter([
  {
    path: '/',
    element: <RequireAuth />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { index: true, element: <HomePage /> },
          { path: 'assessment', element: <AssessmentPage /> },
          { path: 'breeds', element: <BreedsPage /> },
          { path: 'pets', element: <PetsPage /> },
          { path: 'account/security', element: <SecurityPage /> },
          { path: '*', element: <NotFoundPage /> },
        ],
      },
    ],
  },
  {
    path: '/auth',
    element: <PublicOnlyRoute />,
    children: [
      { index: true, element: <Navigate replace to="/auth/login" /> },
      { path: 'login', element: <LoginPage /> },
      { path: 'register/email', element: <EmailRegisterPage /> },
      { path: 'setup-profile', element: <SetupProfilePage /> },
      { path: 'forgot-password', element: <ForgotPasswordPage /> },
    ],
  },
]);
