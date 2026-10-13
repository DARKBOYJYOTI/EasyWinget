import { mount } from 'svelte'
import './style.css'
import App from './App.svelte'

window.addEventListener('error', (e) => {
  console.error('EasyWinGet Window Error:', e.error || e.message);
});

window.addEventListener('unhandledrejection', (e) => {
  console.error('EasyWinGet Unhandled Rejection:', e.reason);
});

const target = document.getElementById('app')
let app;
try {
  app = mount(App, {
    target: target
  });
} catch (err) {
  console.error('Fatal mount error:', err);
  if (target) {
    target.innerHTML = `<div style="padding: 40px; color: #ff6b6b; font-family: Segoe UI, sans-serif; background: #202020; height: 100vh; box-sizing: border-box;">
      <h2 style="margin-top: 0; color: #fff;">EasyWinGet Initialization Error</h2>
      <p style="color: #aaa;">An unexpected error occurred while starting the UI:</p>
      <pre style="background: #161616; padding: 16px; border-radius: 6px; border: 1px solid #333; overflow: auto; font-family: Consolas, monospace; font-size: 13px;">${err?.stack || err?.message || err}</pre>
    </div>`;
  }
}

export default app
