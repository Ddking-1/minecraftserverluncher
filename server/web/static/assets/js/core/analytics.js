// Lightweight, consent-gated analytics for the Alpha web UI.
//
// No analytics event is sent before the user has accepted cookies.
// Events are anonymous: only the page/route and a coarse action are logged.
// During the alpha this is a no-op unless the backend exposes an endpoint.

import { cookieConsent } from './cookie-consent.js';

const ENDPOINT = '/api/v1/analytics/event';

export const analytics = {
  /** Track a page/route view. Safe to call on every navigation. */
  page(view) {
    if (!cookieConsent.accepted) return;
    this._send({ kind: 'page', view });
  },

  /** Track a UI action (button click, skin upload, login, etc.). */
  event(action, detail = {}) {
    if (!cookieConsent.accepted) return;
    this._send({ kind: 'event', action, detail });
  },

  /** Initialise: track the first page only after consent is known. */
  init(firstView) {
    if (cookieConsent.accepted) {
      this.page(firstView);
    } else {
      // Defer until the user accepts.
      this._waiting = firstView;
    }
  },

  /** Called when the user changes their consent choice. */
  onConsentChange(accepted) {
    if (!accepted) return;
    if (this._waiting) {
      this.page(this._waiting);
      this._waiting = null;
    }
  },

  async _send(body) {
    try {
      await fetch(ENDPOINT, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
        keepalive: true,
      });
    } catch (_) {
      /* Silently drop analytics failures — they must never break the app. */
    }
  },
};

export default analytics;
