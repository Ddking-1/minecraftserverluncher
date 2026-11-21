// Cookie consent: popup, localStorage state, and consent-gated feature gating.
//
// The consent is stored in localStorage so the popup can be dismissed without
// backend involvement during the alpha. When the user accepts, the frontend
// also sends the choice to the backend so server-side features (session
// cookie, analytics events) can be enabled.
//
// IMPORTANT: no analytics or non-essential cookie is set BEFORE the user
// has accepted. Functional cookies (session) are only set after acceptance
// or after a successful login, whichever comes first.

import { config } from '../config.js';

const STORAGE_KEY = config.storagePrefix + 'cookieConsent';
const ACCEPTED_VALUE = 'accepted';
const REJECTED_VALUE = 'rejected';

export const cookieConsent = {
  /** Whether the user has already made a choice (accepted or rejected). */
  get hasChoice() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      return stored === ACCEPTED_VALUE || stored === REJECTED_VALUE;
    } catch (_) {
      return false;
    }
  },

  /** Whether the user accepted (so analytics and non-essential cookies may run). */
  get accepted() {
    try {
      return localStorage.getItem(STORAGE_KEY) === ACCEPTED_VALUE;
    } catch (_) {
      return false;
    }
  },

  /** Persist the choice locally and notify the backend. */
  setChoice(accepted) {
    const value = accepted ? ACCEPTED_VALUE : REJECTED_VALUE;
    try {
      localStorage.setItem(STORAGE_KEY, value);
    } catch (_) {
      /* A locked-down Qt profile can disable localStorage; the choice is lost. */
    }

    if (accepted) {
      this._notifyBackend(true);
    }
  },

  /** Ask the backend to enable server-side features (session cookie, analytics). */
  async _notifyBackend(accepted) {
    try {
      const response = await fetch('/api/v1/cookie-consent', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ accepted }),
      });
      // The backend may 404 during the alpha if the endpoint is not wired yet;
      // that is fine — the frontend choice is enough for the UI-side features.
      if (!response.ok) {
        console.warn('cookie consent backend notify failed:', response.status);
      }
    } catch (_) {
      /* offline or malformed response — the UI-side choice remains. */
    }
  },

  /** Render the consent popup into the given container. Returns a teardown fn. */
  render(container) {
    if (!container) return () => {};
    container.innerHTML = consentHtml();
    const card = container.querySelector('.cookie-consent-card');
    const accept = card.querySelector('[data-cookie-accept]');
    const reject = card.querySelector('[data-cookie-reject]');
    const decline = card.querySelector('[data-cookie-decline]');

    const activate = () => {
      accept.addEventListener('click', () => {
        cookieConsent.setChoice(true);
        this._dismiss(card);
      });
      reject.addEventListener('click', () => {
        cookieConsent.setChoice(false);
        this._dismiss(card);
      });
      decline.addEventListener('click', () => {
        cookieConsent.setChoice(false);
        this._dismiss(card);
      });
    };

    activate();
    return () => {
      card.remove();
    };
  },

  _dismiss(card) {
    card.classList.add('is-done');
    window.setTimeout(() => card.remove(), 260);
  },
};

function consentHtml() {
  return `
    <div class="cookie-consent-card" role="dialog" aria-label="Sütik és adatgyűjtés">
      <div class="cookie-consent-head">
        <span class="cookie-consent-icon">${icon('shield', 22)}</span>
        <strong>Sütik és adatgyűjtés</strong>
      </div>
      <p class="cookie-consent-copy">
        Az Alpha Network sütiket és egyszerű tevékenységnaplót használ a
        működés és a fejlesztés támogatásához. A működéshez szükséges
        sütiket (pl. bejelentkezés) csak akkor állítjuk be, ha elfogadod;
        az adatok gyűjtése is csak a beleegyezés után indul.
      </p>
      <dl class="cookie-consent-list">
        <div><dt>Működési sütik</dt><dd>Bejelentkezés, session — szükséges a fiók használatához.</dd></div>
        <div><dt>Adatgyűjtés</dt><dd>Oldalnézetek, kattintások — csak fejlesztéshez, anonim.</dd></div>
      </dl>
      <div class="cookie-consent-actions">
        <button type="button" class="btn btn-primary" data-cookie-accept>Mindent elfogadok</button>
        <button type="button" class="btn btn-ghost" data-cookie-reject>Csak a szükségeseket</dd>
        <button type="button" class="btn btn-ghost" data-cookie-decline>Nem, elutasítom</button>
      </div>
    </div>
  `;
}

function icon(name, size) {
  /* Keep this module self-contained; reuse the same minimal SVG set the app uses. */
  const icons = {
    shield: `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>`,
  };
  return icons[name] || '';
}

export default cookieConsent;
