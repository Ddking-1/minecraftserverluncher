import { esc } from './dom.js';
import { icon } from './icons.js';

const MAX_TOASTS = 3;
let container = null;

function ensureContainer() {
  container = document.getElementById('toast-root');
  if (!container) {
    container = document.createElement('div');
    container.id = 'toast-root';
    container.className = 'toast-root';
    container.setAttribute('aria-live', 'polite');
    container.setAttribute('aria-atomic', 'false');
    document.body.appendChild(container);
  }
  return container;
}

const TONE_ICON = {
  success: 'check',
  error: 'alert',
  info: 'info',
  warning: 'alert',
};

/**
 * @param {string} message
 * @param {{title?: string, tone?: 'success'|'error'|'info'|'warning', duration?: number}} options
 */
export function toast(message, { title = '', tone = 'success', duration = 3800 } = {}) {
  const root = ensureContainer();
  while (root.children.length >= MAX_TOASTS) root.firstElementChild?.remove();

  const node = document.createElement('div');
  node.className = `toast toast-${tone}`;
  node.setAttribute('role', tone === 'error' ? 'alert' : 'status');
  node.innerHTML = `
    <span class="toast-icon">${icon(TONE_ICON[tone] || 'info', 18)}</span>
    <div class="toast-copy">
      ${title ? `<strong>${esc(title)}</strong>` : ''}
      <span>${esc(message)}</span>
    </div>
    <button class="toast-close" type="button" aria-label="Értesítés bezárása">${icon('x', 14)}</button>
  `;

  const remove = () => {
    if (!node.isConnected) return;
    node.classList.add('is-leaving');
    window.setTimeout(() => node.remove(), 180);
  };

  node.querySelector('.toast-close').addEventListener('click', remove);
  root.appendChild(node);
  window.requestAnimationFrame(() => node.classList.add('is-visible'));
  window.setTimeout(remove, duration);
  return remove;
}
