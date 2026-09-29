import React from 'react';
//COMPONETS,HOOKS,JSX
//Hooks = Functions that let functional components manage state, side effects, refs, context, and more.
import ReactDOM from 'react-dom/client';
import { App } from './app/App';
import './index.css';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
