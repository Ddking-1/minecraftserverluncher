import { icon, brandMark } from '../core/icons.js';
import { config } from '../config.js';
import { DEMO_SKIN } from '../data/mock-db.js';
import { SkinViewer } from './skin-viewer.js';

const COPY = {
  login: {
    eyebrow: 'Alpha teszt · 2026',
    title: 'Lépj be a hálózatba.',
    text: 'A skined, a statisztikáid és a farm toplisták egyetlen letisztult felületen. Az alpha időszakban minden adat teszt jellegű.',
    points: [
      { icon: 'layers', text: 'Élő, forgatható 3D skin előnézet' },
      { icon: 'server', text: 'Szerverenkénti statisztikák és toplisták' },
      { icon: 'shield', text: 'Granuláris staff jogosultságok' },
    ],
    foot: 'A hozzáférésed a szerveroldali jogosultságok alapján érvényes.',
  },
  register: {
    eyebrow: 'Csatlakozás',
    title: 'Csatlakozz az alfa körhöz.',
    text: 'Hozd létre a fiókod, válaszd ki a skined és kövesd a fejlődésed. A regisztráció külön oldalon történik, nem keveredik a belépéssel.',
    points: [
      { icon: 'sparkles', text: 'Azonnali hozzáférés az irányítópulthoz' },
      { icon: 'users', text: 'Saját profil és játékoskereső' },
      { icon: 'wheat', text: 'Bővíthető plugin statisztikák' },
    ],
    foot: 'Alpha build: a funkciók és az adatok változhatnak a tesztidőszakban.',
  },
};

export function authLayout(mode, content) {
  const copy = COPY[mode] || COPY.login;
  return `<div class="auth-layout">
    <section class="auth-brand" aria-hidden="true">
      <div class="auth-brand-inner">
        <div class="auth-brand-top">
          <span class="brand brand-static">
            <span class="brand-logo">${brandMark(34)}</span>
            <span class="brand-text"><strong>${config.brand.name}</strong><span>${config.brand.tagline}</span></span>
          </span>
          <span class="chip chip-green">${icon('sparkles', 13)}${config.brand.alphaLabel}</span>
        </div>

        <div class="auth-copy">
          <span class="eyebrow">${copy.eyebrow}</span>
          <h1>${copy.title}</h1>
          <p>${copy.text}</p>
          <ul class="auth-points">
            ${copy.points
              .map(
                (point) => `<li><span class="point-icon">${icon(point.icon, 16)}</span>${point.text}</li>`,
              )
              .join('')}
          </ul>
        </div>

        <div class="auth-art">
          <div class="art-glow"></div>
          <div class="art-grid"></div>
          <div class="art-stage">
            <canvas class="auth-skin-canvas" data-auth-skin aria-label="Skin előnézet"></canvas>
          </div>
          <div class="art-chip art-chip-top"><span class="status-dot"></span>Alpha Main · online</div>
          <div class="art-chip art-chip-bottom">${icon('layers', 14)} 3D skin előnézet</div>
        </div>

        <p class="auth-brand-foot">${copy.foot}</p>
      </div>
    </section>

    <section class="auth-side">
      <div class="auth-side-top">
        <span class="brand brand-mobile">
          <span class="brand-logo">${brandMark(28)}</span>
          <span class="brand-text"><strong>${config.brand.name}</strong><span>${config.brand.tagline}</span></span>
        </span>
        <button type="button" class="icon-btn theme-toggle" data-theme-toggle aria-label="Téma váltása">
          <span class="theme-icon theme-icon-sun">${icon('sun', 18)}</span>
          <span class="theme-icon theme-icon-moon">${icon('moon', 18)}</span>
        </button>
      </div>

      <div class="auth-card-wrap">
        <div class="auth-card">${content}</div>
        <p class="auth-micro">© 2026 ${config.brand.name} · Alpha build</p>
      </div>
    </section>
  </div>`;
}

export function mountAuthLayout(root) {
  const canvas = root.querySelector('[data-auth-skin]');
  if (!canvas) return () => {};
  const viewer = new SkinViewer(canvas, { skinUrl: DEMO_SKIN, autoRotate: true, showOverlay: true });
  return () => viewer.destroy();
}

export function setButtonLoading(button, loading, label = '') {
  if (!button) return;
  if (loading) {
    button.dataset.originalHtml = button.innerHTML;
    button.disabled = true;
    button.innerHTML = `<span class="spinner" aria-hidden="true"></span><span>${label || 'Feldolgozás…'}</span>`;
  } else {
    button.disabled = false;
    if (button.dataset.originalHtml) {
      button.innerHTML = button.dataset.originalHtml;
      delete button.dataset.originalHtml;
    }
  }
}

export function fieldError(root, field, message = '') {
  const holder = root.querySelector(`[data-error-for="${field}"]`);
  if (holder) {
    holder.textContent = message;
    holder.hidden = !message;
  }
  const input = root.querySelector(`#${field}`) || root.querySelector(`[name="${field}"]`);
  if (input) {
    input.classList.toggle('is-invalid', Boolean(message));
    input.setAttribute('aria-invalid', message ? 'true' : 'false');
  }
}
