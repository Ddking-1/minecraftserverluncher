import { icon } from '../core/icons.js';

export const meta = {
  title: 'Az oldal nem található',
  subtitle: 'A keresett cím nem létezik.',
  layout: (user) => (user ? 'app' : 'plain'),
  nav: '',
};

export function render() {
  return `<div class="not-found">
    <div class="not-found-card">
      <span class="not-found-code">404</span>
      <span class="eyebrow">${icon('alert', 14)} Eltévedtél</span>
      <h1>Ez az oldal nem létezik</h1>
      <p>A keresett cím elavult lehet, vagy soha nem is tartozott a hálózathoz. Térj vissza egy ismert oldalra.</p>
      <div class="not-found-actions">
        <a class="btn btn-primary" href="/dashboard" data-link>${icon('grid', 16)} Irányítópult</a>
        <a class="btn btn-ghost" href="/login" data-link>${icon('key', 16)} Bejelentkezés</a>
      </div>
    </div>
  </div>`;
}
