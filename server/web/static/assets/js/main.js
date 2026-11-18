import { api } from './data/api.js';
import { config } from './config.js';
import { Router } from './core/router.js';
import { initTheme, syncThemeButtons, toggleTheme } from './core/theme.js';
import { initServerContext, onServerChange } from './core/server-context.js';
import { renderShell, mountShell } from './components/app-shell.js';
import { brandMark, icon } from './core/icons.js';
import { cookieConsent } from './core/cookie-consent.js';
import { analytics } from './core/analytics.js';
import * as loginView from './views/login.js';
import * as registerView from './views/register.js';
import * as dashboardView from './views/dashboard.js';
import * as playersView from './views/players.js';
import * as adminView from './views/admin.js';
import * as notFoundView from './views/not-found.js';

const app = document.getElementById('app');

const routes = [
  { path: '/', redirect: () => (api.getCurrentUser() ? '/dashboard' : '/login') },
  { path: '/login', view: loginView },
  { path: '/register', view: registerView },
  { path: '/dashboard', view: dashboardView },
  { path: '/players', view: playersView },
  { path: '/admin', view: adminView },
];

function dismissBootScreen() {
  const boot = document.getElementById('boot-screen');
  if (!boot || boot.classList.contains('is-done')) return;
  boot.classList.add('is-done');
  window.setTimeout(() => boot.remove(), 280);
}

function plainShell(content) {
  return `<div class="plain-shell">
    <header class="plain-header">
      <a class="brand" href="/dashboard" data-link>
        <span class="brand-logo">${brandMark(30)}</span>
        <span class="brand-text"><strong>${config.brand.name}</strong><span>${config.brand.tagline}</span></span>
      </a>
      <button type="button" class="icon-btn theme-toggle" data-theme-toggle aria-label="Téma váltása">
        <span class="theme-icon theme-icon-sun">${icon('sun', 18)}</span>
        <span class="theme-icon theme-icon-moon">${icon('moon', 18)}</span>
      </button>
    </header>
    <main class="plain-main">${content}</main>
    <footer class="plain-foot">© 2026 ${config.brand.name} · ${config.brand.alphaLabel} build</footer>
  </div>`;
}

async function renderRoute(route, ctx) {
  const user = api.getCurrentUser();
  const view = route.view || notFoundView;
  const meta = view.meta || {};
  const layout = typeof meta.layout === 'function' ? meta.layout(user) : meta.layout || 'plain';

  if (route.redirect) {
    const target = typeof route.redirect === 'function' ? route.redirect(user) : route.redirect;
    router.navigate(target, { replace: true });
    return null;
  }

  if (meta.guestOnly && user) {
    router.navigate('/dashboard', { replace: true });
    return null;
  }

  if (layout === 'app' && !user) {
    const next = `${ctx.path}${ctx.query.toString() ? `?${ctx.query.toString()}` : ''}`;
    router.navigate(`/login?next=${encodeURIComponent(next)}`, { replace: true });
    return null;
  }

  document.title = `${meta.title || config.brand.name} · ${config.brand.name}`;
  document.body.dataset.layout = layout;

  const viewCtx = {
    ...ctx,
    user,
    navigate: (url, options) => router.navigate(url, options),
    refresh: () => router.go(),
  };

  const viewHtml = await view.render(viewCtx);
  if (!viewHtml) return null;

  const cleanups = [];

  if (layout === 'app' && user) {
    app.innerHTML = renderShell({ user, page: meta, activeNav: meta.nav });
    const viewRoot = app.querySelector('#view-root');
    viewRoot.innerHTML = viewHtml;
    cleanups.push(mountShell(app, { user, navigate: viewCtx.navigate }));
    try {
      const viewCleanup = typeof view.mount === 'function' ? view.mount(viewRoot, viewCtx) : null;
      if (typeof viewCleanup === 'function') cleanups.push(viewCleanup);
    } catch (error) {
      cleanups.forEach((cleanup) => cleanup());
      throw error;
    }
  } else if (layout === 'auth') {
    // Auth views already render their own split layout (brand panel + form).
    app.innerHTML = viewHtml;
    const viewCleanup = typeof view.mount === 'function' ? view.mount(app, viewCtx) : null;
    if (typeof viewCleanup === 'function') cleanups.push(viewCleanup);
  } else {
    app.innerHTML = plainShell(viewHtml);
    const compact = app.querySelector('.plain-main');
    const viewCleanup = typeof view.mount === 'function' ? view.mount(compact, viewCtx) : null;
    if (typeof viewCleanup === 'function') cleanups.push(viewCleanup);
  }

  syncThemeButtons();
  dismissBootScreen();

  return () => {
    cleanups.forEach((cleanup) => {
      try {
        cleanup();
      } catch (error) {
        console.error('view cleanup failed', error);
      }
    });
  };
}

initTheme();
initServerContext();

// Cookie consent: show the popup before the app renders if the user has not
// made a choice yet. The app shell renders behind it, but no analytics or
// non-essential cookies are set until the choice is made.
let consentTeardown = null;
if (!cookieConsent.hasChoice) {
  const host = document.createElement('div');
  host.setAttribute('data-cookie-consent-host', '');
  host.setAttribute('aria-hidden', 'false');
  // The consent card sits above the boot screen so it is visible immediately.
  document.body.appendChild(host);
  consentTeardown = cookieConsent.render(host);
  // Listen for future consent changes (e.g. if the user visits again later).
  // For the alpha we only have localStorage, so this is a one-time render.
}

const router = new Router({
  routes,
  fallback: { path: '/404', view: notFoundView },
  render: renderRoute,
});

// Wire analytics to consent changes so tracking starts only after acceptance.
if (cookieConsent.accepted) {
  analytics.init('/');
} else {
  // Re-run once the user accepts (the popup handles the choice).
  const originalSetChoice = cookieConsent.setChoice.bind(cookieConsent);
  cookieConsent.setChoice = function (accepted) {
    originalSetChoice(accepted);
    if (accepted) analytics.onConsentChange(true);
  };
}

document.addEventListener('click', (event) => {
  const toggle = event.target instanceof Element ? event.target.closest('[data-theme-toggle]') : null;
  if (!toggle) return;
  event.preventDefault();
  toggleTheme();
});

onServerChange(() => router.go());

router.start();
