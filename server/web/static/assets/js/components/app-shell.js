import { esc } from '../core/dom.js';
import { icon, brandMark } from '../core/icons.js';
import { config } from '../config.js';
import { hasPermission, roleDefinition } from '../core/permissions.js';
import { getCurrentServer, serverSwitcherHtml, setCurrentServer } from '../core/server-context.js';
import { avatar, roleBadge } from './ui.js';
import { toast } from '../core/toast.js';

function navItem({ href, label, iconName, active, tag = '' }) {
  return `<a class="nav-item ${active ? 'is-active' : ''}" href="${href}" data-link ${active ? 'aria-current="page"' : ''}>
    <span class="nav-icon">${icon(iconName, 18)}</span>
    <span>${esc(label)}</span>
    ${tag ? `<span class="nav-tag">${esc(tag)}</span>` : ''}
  </a>`;
}

function brandLockup() {
  return `<a class="brand" href="/dashboard" data-link aria-label="${esc(config.brand.name)} irányítópult">
    <span class="brand-logo">${brandMark(30)}</span>
    <span class="brand-text">
      <strong>${esc(config.brand.name)}</strong>
      <span>${esc(config.brand.tagline)}</span>
    </span>
  </a>`;
}

export function renderShell({ user, page, activeNav }) {
  const server = getCurrentServer();
  const showAdmin = hasPermission(user, 'admin.access');
  const role = roleDefinition(user.role);

  return `<div class="app-shell">
    <aside class="sidebar" id="sidebar" aria-label="Fő navigáció">
      <div class="sidebar-head">
        ${brandLockup()}
        <button type="button" class="icon-btn sidebar-close" data-action="close-nav" aria-label="Menü bezárása">${icon('x', 18)}</button>
      </div>

      <nav class="side-nav">
        <span class="nav-section">Áttekintés</span>
        ${navItem({ href: '/dashboard', label: 'Irányítópult', iconName: 'grid', active: activeNav === 'dashboard' })}
        ${navItem({ href: '/players', label: 'Játékosok', iconName: 'users', active: activeNav === 'players' })}
        ${
          showAdmin
            ? `<span class="nav-section">Staff</span>
               ${navItem({ href: '/admin', label: 'Admin panel', iconName: 'shield', active: activeNav === 'admin', tag: config.brand.alphaLabel })}`
            : ''
        }
      </nav>

      <div class="sidebar-foot">
        <div class="server-card">
          <div class="server-card-head">
            <span class="status is-online"><span class="status-dot"></span>${server?.online ? 'Online' : 'Offline'}</span>
            <span class="server-players">${server?.players ?? 0}/${server?.maxPlayers ?? 0}</span>
          </div>
          <strong>${esc(server?.name || 'Nincs szerver')}</strong>
          <span>${esc(server?.address || '')}</span>
        </div>

        <div class="side-user">
          ${avatar(user, 'sm')}
          <div class="side-user-copy">
            <strong>${esc(user.username)}</strong>
            <span>${esc(role.label)}</span>
          </div>
          <button type="button" class="icon-btn" data-action="logout" title="Kilépés" aria-label="Kilépés">${icon('logout', 17)}</button>
        </div>
      </div>
    </aside>

    <div class="nav-scrim" data-action="close-nav" hidden></div>

    <div class="app-body">
      <header class="topbar">
        <div class="topbar-left">
          <button type="button" class="icon-btn nav-toggle" data-action="open-nav" aria-label="Menü megnyitása" aria-expanded="false" aria-controls="sidebar">${icon('menu', 20)}</button>
          <div class="page-title">
            <h1>${esc(page.title)}</h1>
            <p>${esc(page.subtitle || '')}</p>
          </div>
        </div>
        <div class="topbar-actions">
          ${serverSwitcherHtml()}
          <button type="button" class="icon-btn theme-toggle" data-theme-toggle aria-label="Téma váltása">
            <span class="theme-icon theme-icon-sun">${icon('sun', 18)}</span>
            <span class="theme-icon theme-icon-moon">${icon('moon', 18)}</span>
          </button>
          <a class="topbar-user" href="/dashboard" data-link aria-label="Profil">
            ${avatar(user, 'sm')}
            <span class="topbar-user-copy">
              <strong>${esc(user.username)}</strong>
              ${roleBadge(user.role)}
            </span>
          </a>
        </div>
      </header>

      <main class="main-scroll">
        <div class="view" id="view-root"></div>
      </main>
    </div>
  </div>`;
}

export function mountShell(root, { user, navigate }) {
  const sidebar = root.querySelector('#sidebar');
  const scrim = root.querySelector('.nav-scrim');
  const toggle = root.querySelector('[data-action="open-nav"]');

  const setDrawer = (open) => {
    if (!sidebar) return;
    sidebar.classList.toggle('is-open', open);
    scrim.hidden = !open;
    scrim.classList.toggle('is-visible', open);
    document.body.classList.toggle('nav-open', open);
    if (toggle) toggle.setAttribute('aria-expanded', String(open));
  };

  const onClick = (event) => {
    const action = event.target instanceof Element ? event.target.closest('[data-action]') : null;
    if (!action) return;
    const name = action.getAttribute('data-action');
    if (name === 'open-nav') setDrawer(true);
    if (name === 'close-nav') setDrawer(false);
    if (name === 'logout') {
      event.preventDefault();
      import('../data/api.js').then(({ api }) =>
        api.logout().finally(() => {
          toast('Sikeresen kiléptél.', { tone: 'info', title: 'Viszlát!' });
          navigate('/login');
        }),
      );
    }
  };

  const onKeyDown = (event) => {
    if (event.key === 'Escape') setDrawer(false);
  };

  const onServerChange = (event) => {
    const select = event.target.closest('[data-server-select]');
    if (!select) return;
    if (setCurrentServer(select.value)) {
      const server = getCurrentServer();
      toast(`${server.name} kiválasztva.`, { tone: 'info', title: 'Szerver váltás' });
    }
  };

  const onNavClick = (event) => {
    const link = event.target instanceof Element ? event.target.closest('a[data-link]') : null;
    if (link) setDrawer(false);
  };

  root.addEventListener('click', onClick);
  root.addEventListener('change', onServerChange);
  root.addEventListener('click', onNavClick);
  document.addEventListener('keydown', onKeyDown);

  if (user.status === 'banned') {
    toast('A fiókod le van tiltva.', { tone: 'error', title: 'Hozzáférés korlátozva' });
  }

  return () => {
    root.removeEventListener('click', onClick);
    root.removeEventListener('change', onServerChange);
    root.removeEventListener('click', onNavClick);
    document.removeEventListener('keydown', onKeyDown);
    document.body.classList.remove('nav-open');
  };
}
