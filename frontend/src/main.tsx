import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { App } from './app/App';
import './shared/styles/global.css';

const root = document.getElementById('root');

if (!root) {
  throw new Error('React 根节点不存在，请检查 index.html 中的 #root 容器。');
}

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
