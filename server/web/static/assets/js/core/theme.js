import { storage } from './storage.js';

const THEME_KEY = 'theme';
const DARK = 'dark';
const LIGHT = 'light';

export function getTheme() {
  const value = document.documentElement.getAttribute('data-theme');
  return value === LIGHT ? LIGHT : DARK;
}

export function resolveInitialTheme() {
  const stored = storage.get(THEME_KEY);
  if (stored === DARK || stored === LIGHT) return stored;
  return window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? LIGHT : DARK;
}

export function applyTheme(theme, { persist = false } = {}) {
  const next = theme === LIGHT ? LIGHT : DARK;
  document.documentElement.setAttribute('data-theme', next);
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) meta.setAttribute('content', next === DARK ? '#090d13' : '#f4f7fb');
  if (persist) storage.set(THEME_KEY, next);
  syncThemeButtons();
  window.dispatchEvent(new CustomEvent('alpha:theme', { detail: { theme: next } }));
  return next;
}

export function toggleTheme() {
  return applyTheme(getTheme() === DARK ? LIGHT : DARK, { persist: true });
}

export function initTheme() {
  const theme = resolveInitialTheme();
  return applyTheme(theme);
}

/**
 * Buttons render both icons; CSS reveals the right one per theme. This keeps
 * the markup declarative and avoids rebuilding DOM on every toggle.
 */
export function syncThemeButtons() {
  const theme = getTheme();
  const targetLabel = theme === DARK ? 'Világos mód bekapcsolása' : 'Sötét mód bekapcsolása';
  document.querySelectorAll('[data-theme-toggle]').forEach((button) => {
    button.setAttribute('aria-label', targetLabel);
    button.setAttribute('title', targetLabel);
    button.removeAttribute('aria-pressed');
  });
}
