import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
// Autoalojado (sin CDN de terceros): relevante para privacidad porque este
// sistema sirve a menores de edad, igual que ../usbi/frontend/src/main.tsx.
import '@fontsource/inter/400.css';
import '@fontsource/inter/500.css';
import '@fontsource/inter/700.css';
import './index.css';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
