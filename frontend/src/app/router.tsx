import { createBrowserRouter } from 'react-router-dom';

import { AppLayout } from './AppLayout';
import { AssessmentPage } from '../pages/assessment/AssessmentPage';
import { BreedsPage } from '../pages/breeds/BreedsPage';
import { HomePage } from '../pages/home/HomePage';
import { NotFoundPage } from '../pages/not-found/NotFoundPage';
import { PetsPage } from '../pages/pets/PetsPage';

// router 只描述页面路由关系，具体业务动作由 pages/features 承接。
export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: 'assessment', element: <AssessmentPage /> },
      { path: 'breeds', element: <BreedsPage /> },
      { path: 'pets', element: <PetsPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
]);
