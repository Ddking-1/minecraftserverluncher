import { api } from '../data/api.js';
import { esc, formatDate, relativeTime } from '../core/dom.js';
import { icon } from '../core/icons.js';
import { getCurrentServer } from '../core/server-context.js';
import { hasPermission, roleDefinition } from '../core/permissions.js';
import { leaderboardModules, buildLeaderboard } from '../stats/registry.js';
import '../stats/modules.js';
import { SkinViewer } from '../components/skin-viewer.js';
import { avatar, chip, leaderboardRows, roleBadge, statModulesHtml, statusDot } from '../components/ui.js';
import { toast } from '../core/toast.js';
import { DEMO_SKIN } from '../data/mock-db.js';

export const meta = {
  title: 'Irányítópult',
  layout: 'app',
  nav: 'dashboard',
  subtitle: 'A fiókod és a szerver áttekintése.',
};

const ACTIVITY_ICON = {
  achievement: 'sparkles',
  join: 'play',
  skin: 'image',
  farm: 'wheat',
};

function heroHtml(user, profile, server) {
  const role = roleDefinition(user.role);
  return `<section class="hero-card">
    <div class="hero-glow" aria-hidden="true"></div>
    <div class="hero-main">
      <div class="hero-avatar-wrap">
        ${avatar(profile, 'xl')}
        <span class="hero-online" title="${profile.online ? 'Online' : 'Offline'}"></span>
      </div>
      <div class="hero-copy">
        <span class="eyebrow">${icon('sparkles', 14)} Alpha teszt · ${esc(server?.name || 'Szerver')}</span>
        <h2>Szia, ${esc(user.username)}!</h2>
        <div class="hero-sub">${roleBadge(user.role)} ${chip(profile.online ? 'Online' : 'Offline', profile.online ? 'green' : 'neutral', 'activity')} ${chip('Fiók aktív', 'green', 'check')}</div>
        <p>Itt a legfrissebb áttekintés a fejlődésedről, a skinedről és a szerver toplistáiról.</p>
      </div>
    </div>
    <div class="hero-meta">
      <div class="hero-meta-item">
        <span class="meta-label">${icon('server', 14)} Szerver</span>
        <strong>${esc(server?.name || '—')}</strong>
        <span class="meta-sub">${esc(server?.region || '')}</span>
      </div>
      <div class="hero-meta-item">
        <span class="meta-label">${icon('shield', 14)} Rang</span>
        <strong>${esc(role.label)}</strong>
        <span class="meta-sub">${hasPermission(user, 'admin.access') ? 'Staff hozzáférés' : 'Játékos fiók'}</span>
      </div>
      <div class="hero-meta-item">
        <span class="meta-label">${icon('clock', 14)} Csatlakozott</span>
        <strong>${esc(formatDate(profile.createdAt))}</strong>
        <span class="meta-sub">${esc(relativeTime(profile.lastSeen))} aktív</span>
      </div>
    </div>
  </section>`;
}

function leaderboardHtml(user, players) {
  const modules = leaderboardModules();
  if (!modules.length) return '';
  const active = modules[0];
  return `<section class="panel leaderboard-panel">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Toplisták</span>
        <h3>Legjobb játékosok</h3>
      </div>
      <span class="icon-box">${icon('crown', 18)}</span>
    </div>
    <div class="tab-row" role="tablist" aria-label="Toplista típusa">
      ${modules
        .map(
          (module, index) => `<button type="button" class="tab ${index === 0 ? 'is-active' : ''}" role="tab" aria-selected="${index === 0}" data-lb-tab="${esc(module.id)}">${esc(module.leaderboard.label)}</button>`,
        )
        .join('')}
    </div>
    <div class="lb-panels">
      ${modules
        .map(
          (module, index) => `<div class="lb-panel" data-lb-panel="${esc(module.id)}" ${index === 0 ? '' : 'hidden'}>
            ${leaderboardRows(buildLeaderboard(module, players, 5), {
              currentUsername: user.username,
              format: module.leaderboard.format,
            })}
          </div>`,
        )
        .join('')}
    </div>
    <a class="panel-link" href="/players" data-link>Összes játékos megtekintése ${icon('arrow-right', 15)}</a>
  </section>`;
}

function activityHtml(activity, players) {
  const byName = new Map(players.map((player) => [player.username.toLowerCase(), player]));
  return `<section class="panel activity-panel">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Közösség</span>
        <h3>Legutóbbi aktivitás</h3>
      </div>
      <span class="icon-box">${icon('activity', 18)}</span>
    </div>
    <div class="activity-list">
      ${activity
        .slice(0, 4)
        .map((item) => {
          const player = byName.get(item.user.toLowerCase());
          return `<div class="activity-row">
            <span class="activity-avatar">${avatar(player || item.user, 'xs')}<span class="activity-badge">${icon(ACTIVITY_ICON[item.kind] || 'activity', 11)}</span></span>
            <div class="activity-copy">
              <p><strong>${esc(item.user)}</strong> ${esc(item.text)}</p>
              <span>${esc(relativeTime(item.at))}</span>
            </div>
          </div>`;
        })
        .join('')}
    </div>
  </section>`;
}

function accountHtml(user, profile) {
  return `<section class="panel account-panel">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Fiók</span>
        <h3>Fiók állapota</h3>
      </div>
      <span class="icon-box">${icon('lock', 18)}</span>
    </div>
    <dl class="account-list">
      <div><dt>Státusz</dt><dd>${chip('Aktív', 'green', 'check')}</dd></div>
      <div><dt>Jogosultság</dt><dd>${roleBadge(user.role)}</dd></div>
      <div><dt>E-mail</dt><dd class="account-email">${esc(profile.email || '')}<span>${chip(profile.emailVerified ? 'Igazolt' : 'Nem igazolt', profile.emailVerified ? 'green' : 'gold', profile.emailVerified ? 'check' : 'mail')}</span></dd></div>
      <div><dt>2FA</dt><dd>${chip(profile.twoFactor ? 'Aktív' : 'Nincs beállítva', profile.twoFactor ? 'green' : 'neutral', 'shield')}</dd></div>
    </dl>
  </section>`;
}

function skinPanelHtml(profile) {
  return `<section class="panel skin-panel">
    <div class="panel-head">
      <div>
        <span class="panel-eyebrow">Megjelenés</span>
        <h3>Skin kezelés</h3>
        <p>Húzd a PNG-t a mezőre, és az előnézet azonnal frissül.</p>
      </div>
      <span class="chip chip-neutral" id="skin-model-chip">${icon('layers', 13)} 64×64 · 1.8</span>
    </div>

    <div class="skin-layout">
      <div class="skin-side">
        <div class="dropzone" id="skin-dropzone" role="button" tabindex="0" aria-label="Skin feltöltése: kattints vagy húzd ide a PNG fájlt">
          <span class="dropzone-icon">${icon('upload', 22)}</span>
          <strong>Skin feltöltése</strong>
          <span>Húzd ide, vagy kattints a tallózáshoz</span>
          <small>PNG · 64×64 vagy 64×32 · max. 512 KB</small>
          <input type="file" id="skin-file" accept="image/png,.png" hidden>
        </div>

        <div class="skin-actions">
          <button type="button" class="btn btn-primary" id="skin-save" disabled>${icon('check', 16)} Skin mentése</button>
          <button type="button" class="btn btn-ghost" id="skin-reset" disabled>${icon('refresh', 16)} Visszaállítás</button>
        </div>

        <p class="skin-message" id="skin-message" role="status">Az aktuális skin van betöltve.</p>

        <dl class="skin-facts">
          <div><dt>Modell</dt><dd id="skin-model">64×64 · 1.8</dd></div>
          <div><dt>Forrás</dt><dd>Helyi előnézet</dd></div>
          <div><dt>Mentés</dt><dd>Alpha mock</dd></div>
        </dl>
      </div>

      <div class="skin-stage" id="skin-stage">
        <div class="stage-grid" aria-hidden="true"></div>
        <canvas id="skin-canvas" tabindex="0" aria-label="Interaktív 3D skin előnézet. Húzd az egérrel a forgatáshoz, nyílbillentyűkkel is forgatható."></canvas>
        <span class="stage-hint">${icon('rotate', 14)} Húzd a forgatáshoz</span>
        <div class="stage-tools">
          <button type="button" class="icon-btn" data-skin-action="reset" title="Nézet visszaállítása" aria-label="Nézet visszaállítása">${icon('rotate', 16)}</button>
          <button type="button" class="icon-btn is-active" data-skin-action="overlay" title="Külső réteg megjelenítése" aria-label="Külső réteg megjelenítése" aria-pressed="true">${icon('layers', 16)}</button>
        </div>
      </div>
    </div>
  </section>`;
}

export async function render(ctx) {
  const server = getCurrentServer();
  const [profile, players, activity] = await Promise.all([
    api.getPlayer(ctx.user.username, server.id),
    api.listPlayers({ serverId: server.id }),
    api.getActivity(),
  ]);

  const safeProfile = {
    ...ctx.user,
    ...(profile || {}),
    // getPlayer returns the public profile, but the account panel is allowed
    // to show the signed-in user's own verification details.
    email: ctx.user.email,
    emailVerified: ctx.user.emailVerified,
    twoFactor: ctx.user.twoFactor,
    stats: profile?.stats || {
      core: { playtimeHours: 0, blocksMined: 0, mobsKilled: 0, deaths: 0 },
      economy: { balance: 0, earned: 0 },
      farm: { level: 0, harvested: 0 },
    },
  };

  return `
    <div class="page page-dashboard">
      ${heroHtml(ctx.user, safeProfile, server)}

      <div class="dashboard-grid">
        <div class="dashboard-main">
          <section class="panel stats-panel">
            <div class="panel-head">
              <div>
                <span class="panel-eyebrow">Statisztikák</span>
                <h3>Teljesítményed</h3>
                <p>${esc(server?.name || '')} · összesített adatok</p>
              </div>
              <span class="chip chip-green">${icon('activity', 13)} Élő adatok</span>
            </div>
            <div class="stat-modules" id="stat-modules">
              ${statModulesHtml(safeProfile)}
              <p class="stat-note">${icon('info', 14)} A modulok a regisztrált plugin adatokból épülnek fel, így új statisztika hozzáadása nem igényel oldal-átír\u00e1st.</p>
            </div>
          </section>

          ${skinPanelHtml(safeProfile)}
        </div>

        <aside class="dashboard-side">
          ${leaderboardHtml(ctx.user, players)}
          ${activityHtml(activity, players)}
          ${accountHtml(ctx.user, safeProfile)}
        </aside>
      </div>
    </div>
  `;
}

export function mount(root, ctx) {
  const canvas = root.querySelector('#skin-canvas');
  const dropzone = root.querySelector('#skin-dropzone');
  const fileInput = root.querySelector('#skin-file');
  const saveButton = root.querySelector('#skin-save');
  const resetButton = root.querySelector('#skin-reset');
  const message = root.querySelector('#skin-message');
  const modelChip = root.querySelector('#skin-model-chip');
  const modelValue = root.querySelector('#skin-model');
  const viewer = new SkinViewer(canvas, { skinUrl: ctx.user.skin || DEMO_SKIN, autoRotate: true });
  let pendingSkin = '';

  const setModel = (model) => {
    if (!model) return;
    if (modelChip) modelChip.innerHTML = `${icon('layers', 13)} ${esc(model)}`;
    if (modelValue) modelValue.textContent = model;
  };

  const setMessage = (text, tone = '') => {
    if (!message) return;
    message.textContent = text;
    message.dataset.tone = tone;
  };

  const stateAfterPreview = (dataUrl) => {
    pendingSkin = dataUrl;
    saveButton.disabled = false;
    resetButton.disabled = false;
    setMessage('Előnézet frissítve. A mentéshez kattints a „Skin mentése” gombra.', 'success');
  };

  const readFile = (file) => {
    if (!file) return;
    if (file.type !== 'image/png' && !file.name.toLowerCase().endsWith('.png')) {
      setMessage('Csak PNG fájl tölthető fel.', 'error');
      toast('A skin csak PNG formátumban tölthető fel.', { tone: 'error' });
      return;
    }
    if (file.size > 512 * 1024) {
      setMessage('A fájl túl nagy (max. 512 KB).', 'error');
      toast('A skin legfeljebb 512 KB lehet.', { tone: 'error' });
      return;
    }

    const reader = new FileReader();
    reader.onerror = () => setMessage('A fájl beolvasása nem sikerült.', 'error');
    reader.onload = async () => {
      try {
        const result = await viewer.setSkin(String(reader.result));
        setModel(result.model);
        stateAfterPreview(String(reader.result));
      } catch (error) {
        setMessage(error.message || 'A skin előnézete nem készíthető el.', 'error');
        toast(error.message || 'Érvénytelen skin fájl.', { tone: 'error' });
      }
    };
    reader.readAsDataURL(file);
  };

  const openPicker = () => fileInput.click();

  const onDropzoneKey = (event) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      openPicker();
    }
  };

  const onDragOver = (event) => {
    event.preventDefault();
    dropzone.classList.add('is-dragover');
  };
  const onDragLeave = () => dropzone.classList.remove('is-dragover');
  const onDrop = (event) => {
    event.preventDefault();
    dropzone.classList.remove('is-dragover');
    readFile(event.dataTransfer?.files?.[0]);
  };

  dropzone.addEventListener('click', openPicker);
  dropzone.addEventListener('keydown', onDropzoneKey);
  dropzone.addEventListener('dragover', onDragOver);
  dropzone.addEventListener('dragleave', onDragLeave);
  dropzone.addEventListener('drop', onDrop);
  const stopInputBubble = (event) => event.stopPropagation();
  fileInput.addEventListener('click', stopInputBubble);
  fileInput.addEventListener('change', () => readFile(fileInput.files?.[0]));

  saveButton.addEventListener('click', async () => {
    if (!pendingSkin) return;
    saveButton.disabled = true;
    saveButton.innerHTML = '<span class="spinner" aria-hidden="true"></span> Mentés…';
    try {
      await api.updateSkin(pendingSkin);
      toast('A skin mentve. Az admin jóváhagyás után jelenik meg mindenhol.', { tone: 'success', title: 'Skin mentve' });
      ctx.refresh();
    } catch (error) {
      toast(error.message || 'A skin mentése nem sikerült.', { tone: 'error' });
      saveButton.disabled = false;
      saveButton.innerHTML = `${icon('check', 16)} Skin mentése`;
    }
  });

  resetButton.addEventListener('click', async () => {
    pendingSkin = '';
    saveButton.disabled = true;
    resetButton.disabled = true;
    setMessage('Az aktuális skin van betöltve.');
    try {
      const result = await viewer.setSkin(ctx.user.skin || DEMO_SKIN);
      setModel(result.model);
    } catch (error) {
      setMessage('A skin visszaállítása nem sikerült.', 'error');
    }
  });

  const onStageAction = (event) => {
    const action = event.target.closest('[data-skin-action]');
    if (!action) return;
    if (action.dataset.skinAction === 'reset') {
      viewer.resetView();
      return;
    }
    if (action.dataset.skinAction === 'overlay') {
      const next = action.getAttribute('aria-pressed') !== 'true';
      viewer.setOverlay(next);
      action.setAttribute('aria-pressed', String(next));
      action.classList.toggle('is-active', next);
    }
  };
  root.querySelector('#skin-stage').addEventListener('click', onStageAction);

  const onTabClick = (event) => {
    const tab = event.target.closest('[data-lb-tab]');
    if (!tab) return;
    root.querySelectorAll('[data-lb-tab]').forEach((button) => {
      button.classList.toggle('is-active', button === tab);
      button.setAttribute('aria-selected', String(button === tab));
    });
    root.querySelectorAll('[data-lb-panel]').forEach((panel) => {
      panel.hidden = panel.dataset.lbPanel !== tab.dataset.lbTab;
    });
  };
  root.addEventListener('click', onTabClick);

  return () => {
    viewer.destroy();
    dropzone.removeEventListener('click', openPicker);
    dropzone.removeEventListener('keydown', onDropzoneKey);
    dropzone.removeEventListener('dragover', onDragOver);
    dropzone.removeEventListener('dragleave', onDragLeave);
    dropzone.removeEventListener('drop', onDrop);
    fileInput.removeEventListener('click', stopInputBubble);
    root.removeEventListener('click', onTabClick);
  };
}
