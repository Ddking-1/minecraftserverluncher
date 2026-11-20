import { api } from '../data/api.js';
import { createRealAdminApi } from '../data/admin-real.js';
import { storage } from '../core/storage.js';
import { config } from '../config.js';
import { esc, formatDate, relativeTime } from '../core/dom.js';
import { icon } from '../core/icons.js';
import {
  PERMISSION_GROUPS,
  PERMISSIONS,
  ROLES,
  ROLE_ORDER,
  assignableRoles,
  canAssignRole,
  hasPermission,
  permissionsFor,
  roleDefinition,
} from '../core/permissions.js';
import { avatar, emptyState, roleBadge } from '../components/ui.js';
import { SkinViewer } from '../components/skin-viewer.js';
import { confirmDialog } from '../core/modal.js';
import { toast } from '../core/toast.js';

export const meta = {
  title: 'Admin panel',
  layout: 'app',
  nav: 'admin',
  subtitle: 'Rangok, jogosultságok és moderáció.',
};

const ACTION_LABEL = {
  'players.ban': 'Játékos letiltása',
  'players.kick': 'Játékos kirúgása',
  'players.mute': 'Némítás',
  'players.warn': 'Figyelmeztetés',
  'roles.assign': 'Rang kiosztás',
  'accounts.delete': 'Fiók törlése',
  'skins.approve': 'Skin jóváhagyás',
};

function accessDeniedHtml(actualUser) {
  const role = roleDefinition(actualUser.role);
  return `<section class="panel access-denied">
    <span class="denied-icon">${icon('lock', 26)}</span>
    <span class="eyebrow">403 · Nincs jogosultság</span>
    <h2>Az admin panel zárt terület</h2>
    <p>A(z) <strong>${esc(role.label)}</strong> rangod nem rendelkezik admin hozzáféréssel. A jogosultságokat a szerver ellenőrzi, a felület csak a rendelkezésre álló műveleteket mutatja.</p>
    <div class="denied-actions">
      <a class="btn btn-primary" href="/dashboard" data-link>${icon('arrow-left', 16)} Vissza az irányítópultra</a>
      ${config.features.permissionSimulator ? `<button type="button" class="btn btn-ghost" id="demo-admin-login">${icon('shield', 16)} Admin demó fiók használata</button>` : ''}
    </div>
    ${config.features.permissionSimulator ? '<p class="security-note">Alpha demó: a gomb átvált az Aurora tulajdonosi fiókra, hogy kipróbálhasd a jogosultsági szinteket.</p>' : ''}
  </section>`;
}

function simulatorHtml(actualUser, effectiveRole) {
  const activePermissions = permissionsFor(effectiveRole);
  return `<section class="admin-sim" aria-label="Jogosultság szimulátor">
    <div class="sim-copy">
      <span class="eyebrow">${icon('sparkles', 13)} Alpha eszköz</span>
      <h2>Jogosultság szimulátor</h2>
      <p>Válts rangot, és nézd meg, mely gombok válnak elérhetővé vagy tiltottá. A szimuláció csak a felületet befolyásolja.</p>
    </div>
    <div class="sim-controls">
      <label class="select-field" for="admin-role-sim">
        <span>Hatékony rang</span>
        <span class="select-wrap">
          <select id="admin-role-sim">
            ${ROLE_ORDER.map(
              (role) =>
                `<option value="${role}" ${role === effectiveRole ? 'selected' : ''}>${esc(ROLES[role].label)}${role === actualUser.role ? ' (saját)' : ''}</option>`,
            ).join('')}
          </select>
          ${icon('chevron-down', 14, 'select-caret')}
        </span>
      </label>
      <div class="sim-active" id="sim-role-badge">${roleBadge(effectiveRole)}<span id="sim-perm-count">${activePermissions.length}/${PERMISSIONS.length} jog</span></div>
    </div>
    <div class="sim-perms" id="sim-perms">
      ${PERMISSIONS.map((permission) => {
        const active = hasPermission(effectiveRole, permission.key);
        return `<span class="perm-chip ${active ? 'is-on' : 'is-off'}">${icon(active ? 'check' : 'lock', 12)}${esc(permission.label)}</span>`;
      }).join('')}
    </div>
    <p class="sim-note">${icon('info', 13)} Éles környezetben a backend minden műveletnél újra ellenőrzi a jogosultságot.</p>
  </section>`;
}

function summaryHtml(summary) {
  const cards = [
    { icon: 'users', value: summary.total, label: 'Összes fiók', sub: `${summary.staff} staff` },
    { icon: 'check', value: summary.active, label: 'Aktív fiók', sub: 'játékra jogosult' },
    { icon: 'activity', value: summary.online, label: 'Online', sub: 'jelenleg a hálózaton' },
    { icon: 'ban', value: summary.banned, label: 'Letiltva', sub: `${summary.pendingSkins} skin jóváhagyásra vár` },
  ];
  return `<div class="admin-kpis">
    ${cards
      .map(
        (card) => `<article class="kpi-card">
          <span class="kpi-icon">${icon(card.icon, 18)}</span>
          <div><strong>${esc(card.value)}</strong><span>${esc(card.label)}</span><small>${esc(card.sub)}</small></div>
        </article>`,
      )
      .join('')}
  </div>`;
}

function roleSelectHtml(user, effectiveRole, actualUser) {
  const isSelf = user.id === actualUser.id;
  const allowed = assignableRoles(effectiveRole);
  const canEdit = !isSelf && !user._real && allowed.includes(user.role) && hasPermission(effectiveRole, 'roles.assign');
  const options = canEdit
    ? Array.from(new Set([user.role, ...allowed]))
    : [user.role];
  const title = isSelf
    ? 'Saját rangodon nem módosíthatsz'
    : hasPermission(effectiveRole, 'roles.assign')
      ? canEdit
        ? 'Rang módosítása'
        : 'Ezt a rangot nem oszthatod ki'
      : 'Nincs jogosultságod rangot kiosztani';

  return `<span class="select-wrap select-wrap-sm" title="${esc(title)}">
    <select class="select-sm" data-role-select data-user-id="${esc(user.id)}" data-username="${esc(user.username)}" ${canEdit ? '' : 'disabled'} aria-label="${esc(user.username)} rangja">
      ${options
        .map((role) => `<option value="${role}" ${role === user.role ? 'selected' : ''}>${esc(ROLES[role].label)}</option>`)
        .join('')}
    </select>
    ${icon('chevron-down', 13, 'select-caret')}
  </span>`;
}

function actionButton({ action, user, effectiveRole, actualUser, permission, iconName, label, tone = '', forceDisabled = false, forceTitle = '' }) {
  const isSelf = user.id === actualUser.id;
  const allowed = hasPermission(effectiveRole, permission) && !isSelf && !forceDisabled;
  const title = allowed
    ? label
    : isSelf
      ? 'Saját fiókon nem hajtható végre'
      : forceTitle || 'Nincs jogosultságod ehhez a művelethez';
  return `<button type="button" class="icon-btn action-btn ${tone} ${allowed ? '' : 'is-denied'}" data-admin-action="${action}" data-user-id="${esc(user.id)}" data-username="${esc(user.username)}" ${allowed ? '' : 'disabled'} title="${esc(title)}" aria-label="${esc(title)}">
    ${icon(iconName, 16)}
  </button>`;
}

function userRowHtml(user, effectiveRole, actualUser) {
  const isSelf = user.id === actualUser.id;
  const banned = user.status === 'banned';
  return `<tr data-user-row data-username="${esc(user.username.toLowerCase())}" data-email="${esc(user.email.toLowerCase())}">
    <td>
      <div class="table-user">
        ${avatar(user, 'sm')}
        <div class="table-user-copy">
          <strong>${esc(user.username)}${isSelf ? '<span class="you-tag">te</span>' : ''}</strong>
          <span>${esc(user.email)}</span>
        </div>
      </div>
    </td>
    <td>${roleSelectHtml(user, effectiveRole, actualUser)}</td>
    <td>
      <div class="status-stack">
        <span class="badge ${banned ? 'badge-danger' : 'badge-success'}">${banned ? 'Letiltva' : user.online ? 'Online' : 'Aktív'}</span>
        ${user.muted ? `<span class="badge badge-warning">${icon('volume-x', 11)} Némítva</span>` : ''}
      </div>
    </td>
    <td class="hide-sm"><span class="table-muted">${esc(relativeTime(user.lastSeen))}</span></td>
    <td>
      <div class="row-actions">
        ${actionButton({ action: 'ban', user, effectiveRole, actualUser, permission: 'players.ban', iconName: 'ban', label: banned ? 'Letiltás feloldása' : 'Játékos letiltása', tone: banned ? 'is-warning' : 'is-danger' })}
        ${actionButton({ action: 'kick', user, effectiveRole, actualUser, permission: 'players.kick', iconName: 'logout', label: 'Kirúgás a szerverről', forceDisabled: !user.online, forceTitle: 'A játékos nincs online' })}
        ${actionButton({ action: 'mute', user, effectiveRole, actualUser, permission: 'players.mute', iconName: 'volume-x', label: user.muted ? 'Némítás feloldása' : 'Némítás', tone: user.muted ? 'is-active' : '' })}
        ${actionButton({ action: 'warn', user, effectiveRole, actualUser, permission: 'players.mute', iconName: 'alert', label: 'Figyelmeztetés', tone: '' })}
        ${actionButton({ action: 'delete', user, effectiveRole, actualUser, permission: 'accounts.delete', iconName: 'trash', label: 'Fiók végleges törlése', tone: 'is-danger' })}
        ${actionButton({ action: 'password-reset', user, effectiveRole, actualUser, permission: 'accounts.view', iconName: 'key', label: 'Jelszó visszaállítása', tone: '' })}
      </div>
    </td>
  </tr>`;
}

function usersHtml(state, actualUser) {
  const canView = hasPermission(state.effectiveRole, 'accounts.view');
  if (!canView) {
    return emptyState({
      iconName: 'lock',
      title: 'Nincs fióklistázási jogod',
      text: 'A jelenlegi rang nem tartalmazza az accounts.view jogosultságot.',
    });
  }

  return `<section class="panel admin-users">
    <div class="panel-head admin-users-head">
      <div>
        <span class="panel-eyebrow">Kezelés</span>
        <h3>Felhasználók</h3>
        <p>${state.users.length} regisztrált fiók · a műveletek azonnal frissülnek.</p>
      </div>
      <div class="admin-search">
        ${icon('search', 16)}
        <input type="search" id="admin-user-search" placeholder="Név vagy e-mail…" value="${esc(state.query)}" autocomplete="off" aria-label="Felhasználó keresése">
      </div>
    </div>
    <div class="table-wrap">
      <table class="data-table admin-table">
        <thead>
          <tr><th scope="col">Játékos</th><th scope="col">Rang</th><th scope="col">Állapot</th><th scope="col" class="hide-sm">Utolsó belépés</th><th scope="col" class="actions-col">Műveletek</th></tr>
        </thead>
        <tbody>
          ${state.users.map((user) => userRowHtml(user, state.effectiveRole, actualUser)).join('')}
          <tr id="admin-no-results" hidden><td colspan="5">${emptyState({ iconName: 'search', title: 'Nincs találat', text: 'Módosítsd a keresést.' })}</td></tr>
        </tbody>
      </table>
    </div>
    <p class="table-note">${icon('lock', 13)} A letiltott gombokhoz a jelenlegi rang nem rendelkezik jogosultsággal. A szürke rangválasztó nem kiosztható.</p>
  </section>`;
}

function matrixHtml(effectiveRole) {
  return `<section class="panel perm-matrix">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Jogosultságok</span>
        <h3>Rang mátrix</h3>
        <p>A szimulált rang oszlopa kiemelve.</p>
      </div>
      <span class="chip chip-neutral">${icon('shield', 13)} ${esc(roleDefinition(effectiveRole).label)}</span>
    </div>
    <div class="table-wrap">
      <table class="perm-table">
        <thead>
          <tr>
            <th>Jogosultság</th>
            ${ROLE_ORDER.map((role) => `<th class="${role === effectiveRole ? 'is-effective' : ''}">${esc(ROLES[role].label)}</th>`).join('')}
          </tr>
        </thead>
        <tbody>
          ${PERMISSION_GROUPS.map((group) => {
            const rows = PERMISSIONS.filter((permission) => permission.group === group);
            if (!rows.length) return '';
            return `<tr class="perm-group-row"><td colspan="${ROLE_ORDER.length + 1}">${esc(group)}</td></tr>${rows
              .map(
                (permission) => `<tr>
                  <td class="perm-name"><span>${esc(permission.label)}</span><small>${esc(permission.description)}</small></td>
                  ${ROLE_ORDER.map((role) => {
                    const allowed = hasPermission(role, permission.key);
                    return `<td class="${role === effectiveRole ? 'is-effective' : ''}">${allowed ? `<span class="perm-yes" title="Engedélyezett">${icon('check', 14)}</span>` : '<span class="perm-no" title="Nem engedélyezett">—</span>'}</td>`;
                  }).join('')}
                </tr>`,
              )
              .join('')}`;
          }).join('')}
        </tbody>
      </table>
    </div>
  </section>`;
}

function skinsHtml(queue) {
  const pending = queue.filter((item) => item.status === 'pending');
  return `<section class="panel admin-skins">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Tartalom</span>
        <h3>Skin jóváhagyás</h3>
        <p>${pending.length} feltöltés vár döntésre. A 3D előnézetben forgathatod a skint.</p>
      </div>
      <span class="icon-box">${icon('image', 18)}</span>
    </div>
    ${
      pending.length
        ? `<div class="skin-queue">
            ${pending
              .map(
                (item) => `<div class="skin-queue-row skin-queue-row-3d">
                  <canvas class="skin-3d-preview" data-skin-canvas="${esc(item.id)}" width="128" height="160" aria-label="${esc(item.username)} skin 3D előnézete"></canvas>
                  <div class="skin-queue-copy">
                    <strong>${esc(item.username)}</strong>
                    <span>${esc(item.fileName)} · ${esc(relativeTime(item.submittedAt))}</span>
                  </div>
                  <div class="row-actions">
                    <button type="button" class="icon-btn action-btn is-success" data-admin-action="skin-approve" data-skin-id="${esc(item.id)}" title="Jóváhagyás" aria-label="Skin jóváhagyása">${icon('check', 16)}</button>
                    <button type="button" class="icon-btn action-btn is-danger" data-admin-action="skin-reject" data-skin-id="${esc(item.id)}" title="Elutasítás" aria-label="Skin elutasítása">${icon('x', 16)}</button>
                  </div>
                </div>`,
              )
              .join('')}
          </div>
          <p class="table-note">${icon('info', 13)} Jóváhagyás után a skin auto-elfogadottá válik. Elutasítás esetén minden játékosról lekerül.</p>`
        : emptyState({ iconName: 'check', title: 'Nincs függőben lévő skin', text: 'Minden feltöltés elbírálásra került.' })
    }
  </section>`;
}

function rejectedSkinsHtml(blacklist) {
  return `<section class="panel admin-rejected-skins">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Tartalom</span>
        <h3>Elutasított skinek (blacklist)</h3>
        <p>${blacklist.length} hash van tiltólistán. Itt visszavonhatod a döntést.</p>
      </div>
      <span class="icon-box">${icon('ban', 18)}</span>
    </div>
    ${
      blacklist.length
        ? `<ul class="rejected-list">
            ${blacklist.map((hash) => `<li class="rejected-item">
              <code class="rejected-hash">${esc(hash)}</code>
              <button type="button" class="btn btn-ghost btn-sm" data-admin-action="blacklist-remove" data-hash="${esc(hash)}">Visszavonás</button>
            </li>`).join('')}
          </ul>`
        : emptyState({ iconName: 'check', title: 'Üres a tiltólista', text: 'Nincs elutasított skin.' })
    }
  </section>`;
}

function auditHtml(audit) {
  return `<section class="panel admin-audit">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Napló</span>
        <h3>Audit események</h3>
        <p>Az utolsó adminisztratív műveletek.</p>
      </div>
      <span class="icon-box">${icon('file-text', 18)}</span>
    </div>
    <div class="audit-list">
      ${audit
        .slice(0, 8)
        .map(
          (entry) => `<div class="audit-row">
            <span class="audit-icon">${icon('shield', 14)}</span>
            <div class="audit-copy">
              <p><strong>${esc(entry.actor)}</strong> · ${esc(ACTION_LABEL[entry.action] || entry.action)} <span>${esc(entry.target)}</span></p>
              <span>${esc(entry.detail)} · ${esc(formatDate(entry.at))}</span>
            </div>
          </div>`,
        )
        .join('')}
    </div>
  </section>`;
}

function adminContentHtml(state, actualUser) {
  const role = state.effectiveRole;
  if (!hasPermission(role, 'admin.access')) {
    return `<section class="panel access-denied access-denied-inline">
      <span class="denied-icon">${icon('lock', 24)}</span>
      <span class="eyebrow">Hozzáférés megtagadva</span>
      <h2>A(z) ${esc(roleDefinition(role).label)} rang nem látja az admin panelt</h2>
      <p>Válts vissza egy staff rangra a fenti szimulátorban. A tiltott műveletek helyett itt csak az olvasható állapot és a biztonsági magyarázat jelenik meg.</p>
    </section>`;
  }

  return `
    ${summaryHtml(state.summary)}
    ${usersHtml(state, actualUser)}
    ${matrixHtml(role)}
    <div class="admin-grid">
      ${hasPermission(role, 'skins.approve') ? skinsHtml(state.queue) : ''}
      ${hasPermission(role, 'skins.approve') ? rejectedSkinsHtml(state.blacklist || []) : ''}
      ${hasPermission(role, 'audit.view') ? auditHtml(state.audit) : ''}
    </div>
  `;
}

export async function render(ctx) {
  const actualUser = ctx.user || { id: 'real', username: 'admin', role: 'owner' };
  const adminToken = storage.get('adminToken') || '';
  const realMode = !!adminToken;

  if (!realMode && !hasPermission(actualUser, 'admin.access')) {
    ctx.adminState = null;
    return `<div class="page page-admin">${accessDeniedHtml(actualUser)}</div>`;
  }

  const adminApi = realMode ? createRealAdminApi(adminToken) : api.admin;

  let users = [], summary = null, queue = [], audit = [], blacklist = [];
  try {
    [users, summary, queue, audit, blacklist] = await Promise.all([
      adminApi.listUsers(),
      adminApi.summary(),
      adminApi.skinsQueue(),
      adminApi.audit(),
      adminApi.blacklist(),
    ]);
  } catch (e) {
    toast(e.message || 'Backend admin hiba.', { tone: 'error' });
    users = []; summary = { total: 0, staff: 0, active: 0, online: 0, banned: 0, pendingSkins: 0 };
    queue = []; audit = []; blacklist = [];
  }
  if (!summary) summary = { total: users.length, staff: 0, active: users.length, online: 0, banned: 0, pendingSkins: 0 };

  const state = {
    actualUser,
    effectiveRole: actualUser.role || 'owner',
    users,
    summary,
    queue,
    audit,
    blacklist,
    query: '',
    realMode,
    adminToken,
    adminApi,
  };
  ctx.adminState = state;

  const tokenBar = `<div class="panel admin-token-bar" style="margin-bottom:1rem">
    <div style="display:flex;gap:.5rem;align-items:center;flex-wrap:wrap">
      <span class="panel-eyebrow">Backend admin token</span>
      <input type="password" id="admin-token-input" placeholder="X-Admin-Token" value="${esc(adminToken)}" style="flex:1;min-width:200px" aria-label="Admin token">
      <button type="button" class="btn btn-primary btn-sm" id="admin-token-set">Beállítás</button>
      <button type="button" class="btn btn-ghost btn-sm" id="admin-token-clear">Törlés</button>
      <span style="font-size:.8rem;opacity:.7">${realMode ? 'Valós backend aktív' : 'Mock mód'}</span>
    </div>
  </div>`;

  return `<div class="page page-admin">
    ${tokenBar}
    ${!realMode && config.features.permissionSimulator ? simulatorHtml(actualUser, state.effectiveRole) : ''}
    <div id="admin-content">${adminContentHtml(state, actualUser)}</div>
  </div>`;
}

export function mount(root, ctx) {
  const state = ctx.adminState;
  const content = root.querySelector('#admin-content');

  if (!state || !content) {
    const demoButton = root.querySelector('#demo-admin-login');
    if (demoButton) {
      demoButton.addEventListener('click', async () => {
        demoButton.disabled = true;
        try {
          await api.demoLogin('admin');
          toast('Belépve az Aurora admin demó fiókba.', { tone: 'info', title: 'Demó mód' });
          ctx.navigate('/admin');
        } catch (error) {
          toast(error.message || 'A demó belépés nem sikerült.', { tone: 'error' });
          demoButton.disabled = false;
        }
      });
    }
    return () => {};
  }

  const simSelect = root.querySelector('#admin-role-sim');

  const syncSimulator = () => {
    if (!simSelect) return;
    simSelect.value = state.effectiveRole;
    const badge = root.querySelector('#sim-role-badge');
    if (badge) badge.innerHTML = `${roleBadge(state.effectiveRole)}<span id="sim-perm-count">${permissionsFor(state.effectiveRole).length}/${PERMISSIONS.length} jog</span>`;
    const chips = root.querySelector('#sim-perms');
    if (chips) {
      chips.innerHTML = PERMISSIONS.map((permission) => {
        const active = hasPermission(state.effectiveRole, permission.key);
        return `<span class="perm-chip ${active ? 'is-on' : 'is-off'}">${icon(active ? 'check' : 'lock', 12)}${esc(permission.label)}</span>`;
      }).join('');
    }
  };

  const applySearch = () => {
    const needle = state.query.trim().toLowerCase();
    let visible = 0;
    content.querySelectorAll('[data-user-row]').forEach((row) => {
      const match = !needle || row.dataset.username.includes(needle) || row.dataset.email.includes(needle);
      row.hidden = !match;
      if (match) visible += 1;
    });
    const empty = content.querySelector('#admin-no-results');
    if (empty) empty.hidden = visible > 0;
  };

  const skinViewers = [];
  const destroySkinViewers = () => {
    skinViewers.forEach((v) => { try { v.destroy(); } catch (e) {} });
    skinViewers.length = 0;
  };
  const initSkinViewers = () => {
    destroySkinViewers();
    content.querySelectorAll('[data-skin-canvas]').forEach((canvas) => {
      const id = canvas.dataset.skinCanvas;
      const item = (state.queue || []).find((q) => q.id === id);
      if (!item) return;
      try {
        const viewer = new SkinViewer(canvas, { skinUrl: item.skin, autoRotate: true });
        skinViewers.push(viewer);
      } catch (e) {
        console.error('skin viewer init failed', e);
      }
    });
  };

  const rerenderContent = () => {
    destroySkinViewers();
    content.innerHTML = adminContentHtml(state, state.actualUser);
    const search = content.querySelector('#admin-user-search');
    if (search) search.value = state.query;
    applySearch();
    syncSimulator();
    initSkinViewers();
  };

  const refreshData = async () => {
    state.users = await state.adminApi.listUsers();
    state.summary = await state.adminApi.summary();
    if (hasPermission(state.effectiveRole, 'skins.approve')) {
      state.queue = await state.adminApi.skinsQueue();
      state.blacklist = await state.adminApi.blacklist();
    }
    if (hasPermission(state.effectiveRole, 'audit.view')) state.audit = await state.adminApi.audit();
    rerenderContent();
  };

  const findUser = (id) => state.users.find((user) => user.id === id);

  const openAdvancedModal = (action, user) => {
    return new Promise((resolve) => {
      const titleMap = { ban: 'Letiltás', kick: 'Kirúgás', mute: 'Némítás', warn: 'Figyelmeztetés' };
      const tone = action === 'ban' || action === 'kick' ? 'danger' : 'primary';
      const overlay = document.createElement('div');
      overlay.className = 'modal-backdrop';
      overlay.setAttribute('role', 'dialog');
      overlay.setAttribute('aria-modal', 'true');
      overlay.setAttribute('aria-labelledby', 'adv-modal-title');
      overlay.innerHTML = `
        <div class="modal" onclick="event.stopPropagation()">
          <h2 id="adv-modal-title">${esc(titleMap[action] || action)} · ${esc(user.username)}</h2>
          <div class="form-row"><label for="adv-duration">Időtartam (perc, 0 = végleges)</label><input type="number" id="adv-duration" value="0" min="0"></div>
          <div class="form-row"><label for="adv-reason">Indok (a játékosnak megjelenik)</label><input type="text" id="adv-reason" maxlength="500" placeholder="indok"></div>
          <div class="form-row"><label for="adv-scope">Hatókör</label>
            <select id="adv-scope">
              <option value="mc">Csak Minecraft</option>
              <option value="platform">Teljes platform</option>
              <option value="both" selected>Mindkettő</option>
            </select></div>
          <div class="form-row"><span>Célpontok</span>
            <label><input type="checkbox" value="account" checked> Fiók</label>
            <label><input type="checkbox" value="hwid"> Minden HWID (30 nap)</label>
            <label><input type="checkbox" value="ip"> IP-k (30 nap)</label>
          </div>
          <div class="form-row" id="adv-preview" style="font-size:.85rem;opacity:.85"><span>Előnézet betöltése…</span></div>
          <div class="modal-actions">
            <button type="button" class="btn btn-ghost" data-adv-cancel>Mégse</button>
            <button type="button" class="btn ${tone === 'danger' ? 'btn-danger' : 'btn-primary'}" data-adv-confirm>${esc(titleMap[action] || action)}</button>
          </div>
        </div>`;
      const previousFocus = document.activeElement;
      const cancelBtn = overlay.querySelector('[data-adv-cancel]');
      const confirmBtn = overlay.querySelector('[data-adv-confirm]');
      const close = (result) => {
        if (!overlay.isConnected) return;
        document.removeEventListener('keydown', onKeyDown);
        overlay.classList.remove('is-visible');
        window.setTimeout(() => overlay.remove(), 160);
        if (previousFocus && typeof previousFocus.focus === 'function') previousFocus.focus();
        resolve(result);
      };
      const onKeyDown = (event) => {
        if (event.key === 'Escape') {
          event.preventDefault();
          close(null);
          return;
        }
        if (event.key !== 'Tab') return;
        const focusables = [...overlay.querySelectorAll('button, input, select, textarea, [tabindex]:not([tabindex="-1"])')].filter((el) => !el.disabled);
        if (focusables.length === 0) return;
        const index = focusables.indexOf(document.activeElement);
        event.preventDefault();
        const next = event.shiftKey
          ? (index <= 0 ? focusables.length - 1 : index - 1)
          : (index === focusables.length - 1 ? 0 : index + 1);
        focusables[next].focus();
      };
      cancelBtn.addEventListener('click', () => close(null));
      confirmBtn.addEventListener('click', () => {
        const targets = [...overlay.querySelectorAll('input[type=checkbox]:checked')].map((c) => c.value);
        if (targets.length === 0) {
          toast('Válassz legalább egy célpontot.', { tone: 'warning' });
          return;
        }
        const duration = parseInt(overlay.querySelector('#adv-duration').value || '0', 10);
        if (!Number.isFinite(duration) || duration < 0) {
          toast('Az időtartam 0 vagy nagyobb legyen.', { tone: 'warning' });
          return;
        }
        const reason = overlay.querySelector('#adv-reason').value;
        if (reason.length > 500) {
          toast('Az indok max 500 karakter.', { tone: 'warning' });
          return;
        }
        close({
          durationMinutes: duration,
          reason,
          scope: overlay.querySelector('#adv-scope').value,
          targets,
          lookbackDays: 30,
        });
      });
      overlay.addEventListener('click', (event) => {
        if (event.target === overlay) close(null);
      });
      document.addEventListener('keydown', onKeyDown);
      document.body.appendChild(overlay);
      window.requestAnimationFrame(() => overlay.classList.add('is-visible'));
      // Load device/IP preview for real backend mode.
      const previewEl = overlay.querySelector('#adv-preview');
      const loadPreview = async () => {
        if (!previewEl) return;
        if (!state.realMode) {
          previewEl.innerHTML = '<span>Mock mód: előnézet nem elérhető.</span>';
          return;
        }
        try {
          const d = await state.adminApi.userDevices(user.id);
          const hwids = d.devices_last_30d || [];
          const ips = d.ips_last_30d || [];
          previewEl.innerHTML = `
            <div><strong>HWID-k (30 nap):</strong> ${hwids.length} eszköz</div>
            <div><strong>IP-k (30 nap):</strong></div>
            <ul style="margin:.25rem 0 0 1rem;padding:0;list-style:disc">
              ${ips.map((ip) => `<li>${esc(ip.ip)} — ${esc(ip.accounts_last_30d)} fiók használta</li>`).join('') || '<li>nincs</li>'}
            </ul>`;
        } catch (e) {
          previewEl.innerHTML = `<span>Előnézet hiba: ${esc(e.message || '')}</span>`;
        }
      };
      loadPreview();
      confirmBtn.focus();
    });
  };

  const handleUserAction = async (action, button) => {
    const user = findUser(button.dataset.userId);
    if (!user) return;
    if (user.id === state.actualUser.id) {
      toast('Saját fiókon nem hajtható végre ez a művelet.', { tone: 'warning' });
      return;
    }

    if (action === 'ban') {
      if (!hasPermission(state.effectiveRole, 'players.ban')) return;
      if (user.status === 'banned') {
        const confirmed = await confirmDialog({
          title: 'Tiltás feloldása',
          message: `${user.username} tiltása feloldásra kerül, a fiók újra használható.`,
          confirmLabel: 'Feloldás',
          tone: 'primary',
        });
        if (!confirmed) return;
        try {
          await state.adminApi.unbanUser(user.id, ['account', 'hwid', 'ip']);
          await refreshData();
          toast(`${user.username} tiltása feloldva.`, { tone: 'success', title: 'Tiltás feloldva' });
        } catch (error) {
          toast(error.message || 'A feloldás nem sikerült.', { tone: 'error' });
        }
        return;
      }
      const params = await openAdvancedModal('ban', user);
      if (!params) return;
      try {
        await state.adminApi.banUser(user.id, params);
        await refreshData();
        toast(`${user.username} letiltva.`, { tone: 'warning', title: 'Letiltás' });
      } catch (error) {
        toast(error.message || 'A letiltás nem sikerült.', { tone: 'error' });
      }
      return;
    }

    if (action === 'kick') {
      if (!hasPermission(state.effectiveRole, 'players.kick')) return;
      if (!user.online) {
        toast('A játékos nincs online.', { tone: 'warning' });
        return;
      }
      const params = await openAdvancedModal('kick', user);
      if (!params) return;
      try {
        await state.adminApi.kickUser(user.id, params);
        await refreshData();
        toast(`${user.username} eltávolítva a szerverről.`, { tone: 'success', title: 'Kirúgás' });
      } catch (error) {
        toast(error.message || 'A kirúgás nem sikerült.', { tone: 'error' });
      }
      return;
    }

    if (action === 'mute') {
      if (!hasPermission(state.effectiveRole, 'players.mute')) return;
      if (user.muted) {
        const confirmed = await confirmDialog({
          title: 'Némítás feloldása',
          message: `${user.username} némítása feloldásra kerül.`,
          confirmLabel: 'Feloldás',
          tone: 'primary',
        });
        if (!confirmed) return;
        try {
          await state.adminApi.unmuteUser(user.id, ['account', 'hwid', 'ip']);
          await refreshData();
          toast(`${user.username} némítása feloldva.`, { tone: 'success', title: 'Némítás' });
        } catch (error) {
          toast(error.message || 'A feloldás nem sikerült.', { tone: 'error' });
        }
        return;
      }
      const params = await openAdvancedModal('mute', user);
      if (!params) return;
      try {
        await state.adminApi.muteUser(user.id, params);
        await refreshData();
        toast(`${user.username} némítva.`, { tone: 'success', title: 'Némítás' });
      } catch (error) {
        toast(error.message || 'A némítás nem sikerült.', { tone: 'error' });
      }
      return;
    }

    if (action === 'warn') {
      if (!hasPermission(state.effectiveRole, 'players.mute')) return;
      const params = await openAdvancedModal('warn', user);
      if (!params) return;
      try {
        await state.adminApi.warnUser(user.id, params);
        await refreshData();
        toast(`${user.username} figyelmeztetve.`, { tone: 'info', title: 'Figyelmeztetés' });
      } catch (error) {
        toast(error.message || 'A figyelmeztetés nem sikerült.', { tone: 'error' });
      }
      return;
    }

    if (action === 'delete') {
      if (!hasPermission(state.effectiveRole, 'accounts.delete')) return;
      const confirmed = await confirmDialog({
        title: 'Fiók végleges törlése',
        message: `${user.username} fiókja és minden statisztikája véglegesen törlődik. Ez a művelet nem vonható vissza.`,
        confirmLabel: 'Végleges törlés',
        tone: 'danger',
      });
      if (!confirmed) return;
      try {
        await state.adminApi.deleteUser(user.id);
        await refreshData();
        toast(`${user.username} fiókja törölve.`, { tone: 'success', title: 'Fiók törölve' });
      } catch (error) {
        toast(error.message || 'A törlés nem sikerült.', { tone: 'error' });
      }
    }
  };

  const onClick = async (event) => {
    const button = event.target.closest('[data-admin-action]');
    if (!button) return;
    const action = button.dataset.adminAction;

    if (action === 'skin-approve' || action === 'skin-reject') {
      if (!hasPermission(state.effectiveRole, 'skins.approve')) return;
      const decision = action === 'skin-approve' ? 'approved' : 'rejected';
      try {
        state.queue = await state.adminApi.reviewSkin(button.dataset.skinId, decision);
        state.summary = await state.adminApi.summary();
        state.audit = await state.adminApi.audit();
        state.blacklist = await state.adminApi.blacklist();
        rerenderContent();
        toast(decision === 'approved' ? 'Skin jóváhagyva (auto-elfogadva a jövőben).' : 'Skin elutasítva (minden játékosról lekerült).', {
          tone: decision === 'approved' ? 'success' : 'info',
          title: 'Skin döntés',
        });
      } catch (error) {
        toast(error.message || 'A döntés nem sikerült.', { tone: 'error' });
      }
      return;
    }

    if (action === 'blacklist-remove') {
      if (!hasPermission(state.effectiveRole, 'skins.approve')) return;
      try {
        await state.adminApi.removeFromBlacklist(button.dataset.hash);
        state.blacklist = await state.adminApi.blacklist();
        rerenderContent();
        toast('A skin visszavonva a tiltólistáról.', { tone: 'success', title: 'Blacklist' });
      } catch (error) {
        toast(error.message || 'A visszavonás nem sikerült.', { tone: 'error' });
      }
      return;
    }

    if (action === 'password-reset') {
      if (!hasPermission(state.effectiveRole, 'accounts.view')) return;
      const user = findUser(button.dataset.userId);
      if (!user) return;
      const confirmed = await confirmDialog({
        title: 'Jelszó visszaállítása',
        message: `${user.username} fiókjához ideiglenes jelszó generálódik. A felhasználónak belépés után új jelszót kell beállítania.`,
        confirmLabel: 'Visszaállítás',
        tone: 'primary',
      });
      if (!confirmed) return;
      try {
        const result = await state.adminApi.resetPassword(user.id);
        await refreshData();
        // Show the temp password once.
        await confirmDialog({
          title: 'Ideiglenes jelszó',
          message: `Felhasználó: ${result.username}\nIdeiglenes jelszó: ${result.tempPassword}\n\nKüldd el a felhasználónak biztonságos csatornán. Belépés után kötelező új jelszót beállítani.`,
          confirmLabel: 'Rendben',
          tone: 'primary',
        });
        toast(`${user.username} jelszava visszaállítva.`, { tone: 'success', title: 'Jelszó reset' });
      } catch (error) {
        toast(error.message || 'A visszaállítás nem sikerült.', { tone: 'error' });
      }
      return;
    }

    await handleUserAction(action, button);
  };

  const onChange = async (event) => {
    if (event.target === simSelect) {
      state.effectiveRole = simSelect.value;
      rerenderContent();
      toast(`Szimulált rang: ${ROLES[state.effectiveRole].label}.`, { tone: 'info', title: 'Jogosultság szimulátor' });
      return;
    }

    const roleSelect = event.target.closest('[data-role-select]');
    if (!roleSelect) return;
    const user = findUser(roleSelect.dataset.userId);
    const nextRole = roleSelect.value;
    if (!user || user.id === state.actualUser.id) return;

    if (!hasPermission(state.effectiveRole, 'roles.assign') || !canAssignRole(state.effectiveRole, nextRole)) {
      toast('Ezt a rangot a jelenlegi rangod nem oszthatja ki.', { tone: 'warning' });
      roleSelect.value = user.role;
      return;
    }

    try {
      await state.adminApi.setRole(user.id, nextRole);
      await refreshData();
      toast(`${user.username} új rangja: ${ROLES[nextRole].label}.`, { tone: 'success', title: 'Rang módosítva' });
    } catch (error) {
      toast(error.message || 'A rang módosítása nem sikerült.', { tone: 'error' });
      roleSelect.value = user.role;
    }
  };

  const onInput = (event) => {
    if (event.target.id !== 'admin-user-search') return;
    state.query = event.target.value;
    applySearch();
  };

  initSkinViewers();

  const tokenInput = root.querySelector('#admin-token-input');
  const tokenSetBtn = root.querySelector('#admin-token-set');
  const tokenClearBtn = root.querySelector('#admin-token-clear');
  if (tokenSetBtn) {
    tokenSetBtn.addEventListener('click', () => {
      const val = (tokenInput && tokenInput.value || '').trim();
      if (val) storage.set('adminToken', val);
      else storage.remove('adminToken');
      ctx.navigate('/admin');
    });
  }
  if (tokenClearBtn) {
    tokenClearBtn.addEventListener('click', () => {
      storage.remove('adminToken');
      ctx.navigate('/admin');
    });
  }

  root.addEventListener('click', onClick);
  root.addEventListener('change', onChange);
  root.addEventListener('input', onInput);

  return () => {
    destroySkinViewers();
    root.removeEventListener('click', onClick);
    root.removeEventListener('change', onChange);
    root.removeEventListener('input', onInput);
  };
}
