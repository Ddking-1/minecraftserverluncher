/** Shared DOM and formatting helpers. */

export const qs = (selector, root = document) => root.querySelector(selector);
export const qsa = (selector, root = document) => Array.from(root.querySelectorAll(selector));

/** Escape interpolated values before they reach innerHTML. */
export function esc(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

/**
 * Delegated event listener. Returns a cleanup function so views can register
 * it through the router and never leak listeners between routes.
 */
export function on(root, type, selector, handler, options) {
  const listener = (event) => {
    const target = event.target instanceof Element ? event.target.closest(selector) : null;
    if (!target || !root.contains(target)) return;
    handler(event, target);
  };
  root.addEventListener(type, listener, options);
  return () => root.removeEventListener(type, listener, options);
}

export function debounce(fn, wait = 200) {
  let timer = 0;
  const debounced = (...args) => {
    window.clearTimeout(timer);
    timer = window.setTimeout(() => fn(...args), wait);
  };
  debounced.cancel = () => window.clearTimeout(timer);
  return debounced;
}

const numberFormatter = new Intl.NumberFormat('hu-HU');

export function formatNumber(value) {
  return numberFormatter.format(Number(value) || 0);
}

export function formatHours(value) {
  const hours = Number(value) || 0;
  if (hours < 1) return `${formatNumber(Math.round(hours * 60))} perc`;
  return `${numberFormatter.format(hours % 1 === 0 ? hours : Number(hours.toFixed(1)))} óra`;
}

export function formatCoins(value) {
  return `${formatNumber(value)} coin`;
}

export function formatDate(value) {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return '—';
  return new Intl.DateTimeFormat('hu-HU', { year: 'numeric', month: 'short', day: 'numeric' }).format(date);
}

export function relativeTime(value) {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return '—';
  const diff = Date.now() - date.getTime();
  const minutes = Math.round(diff / 60000);
  if (minutes < 2) return 'most';
  if (minutes < 60) return `${minutes} perce`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours} órája`;
  const days = Math.round(hours / 24);
  if (days === 1) return 'tegnap';
  if (days < 30) return `${days} napja`;
  return formatDate(date);
}

export function initials(name) {
  const clean = String(name || '?').replace(/[^a-zA-Z0-9_]/g, '');
  return (clean.slice(0, 2) || '?').toUpperCase();
}

/** Stable hue so generated avatars keep the same colour between renders. */
export function nameHue(name) {
  let hash = 0;
  const text = String(name || '');
  for (let i = 0; i < text.length; i += 1) {
    hash = (hash * 31 + text.charCodeAt(i)) % 360;
  }
  return hash;
}

export function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

export function sleep(ms) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

export function isOnline() {
  return navigator.onLine !== false;
}

export function prefersReducedMotion() {
  return window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}
