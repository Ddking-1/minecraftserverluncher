import { api } from '../data/api.js';
import { debounce, esc, formatDate, formatHours, relativeTime } from '../core/dom.js';
import { icon } from '../core/icons.js';
import { getCurrentServer } from '../core/server-context.js';
import { roleBadge, statModulesHtml, avatar, statusDot, emptyState } from '../components/ui.js';
import { SkinViewer } from '../components/skin-viewer.js';
import { roleDefinition } from '../core/permissions.js';

export const meta = {
  title: 'Játékosok',
  layout: 'app',
  nav: 'players',
  subtitle: 'Keresd meg a játékosokat és nézd meg a statisztikáikat.',
};

function playerRow(player, selected) {
  const role = roleDefinition(player.role);
  return `<button type="button" class="player-row ${selected ? 'is-selected' : ''}" data-player="${esc(player.username)}" aria-pressed="${selected}">
    <span class="player-row-avatar">${avatar(player, 'md')}${player.online ? '<span class="presence-dot"></span>' : ''}</span>
    <span class="player-row-main">
      <span class="player-row-name">${esc(player.username)}<span class="role-dot tone-${role.tone}" title="${esc(role.label)}"></span></span>
      <span class="player-row-sub">${esc(role.label)} · ${player.online ? 'online' : `utoljára ${relativeTime(player.lastSeen)}`}</span>
    </span>
    <span class="player-row-time">
      <strong>${esc(formatHours(player.stats.core.playtimeHours))}</strong>
      <span>játékidő</span>
    </span>
    <span class="player-row-arrow">${icon('arrow-right', 16)}</span>
  </button>`;
}

function listHtml(players, selectedName) {
  if (!players.length) {
    return emptyState({
      iconName: 'search',
      title: 'Nincs ilyen játékos',
      text: 'Próbálj másik nevet, vagy töröld a keresést.',
    });
  }
  return players.map((player) => playerRow(player, player.username === selectedName)).join('');
}

function detailHtml(player) {
  if (!player) {
    return emptyState({
      iconName: 'users',
      title: 'Válassz egy játékost',
      text: 'A bal oldali listából kiválasztva látod a profilját, a skinjét és a statisztikáit.',
    });
  }

  return `<div class="player-detail-inner">
    <div class="detail-head">
      <div class="detail-avatar">
        ${avatar(player, 'lg')}
        ${player.online ? '<span class="presence-dot"></span>' : ''}
      </div>
      <div class="detail-identity">
        <h2>${esc(player.username)}</h2>
        <div class="detail-badges">${roleBadge(player.role)} ${statusDot(player.online)} ${player.muted ? `<span class="chip chip-gold">${icon('volume-x', 13)} Némítva</span>` : ''}</div>
        <p>Csatlakozott: ${esc(formatDate(player.createdAt))} · Utolsó aktivitás: ${esc(relativeTime(player.lastSeen))}</p>
      </div>
    </div>

    <div class="detail-showcase">
      <div class="mini-stage">
        <div class="stage-grid" aria-hidden="true"></div>
        <canvas data-detail-skin tabindex="0" aria-label="${esc(player.username)} 3D skin előnézete. Forgasd húzással."></canvas>
        <span class="stage-hint">${icon('rotate', 13)} 3D skin</span>
      </div>
      <div class="detail-facts">
        <div class="detail-fact"><span>Játszott idő</span><strong>${esc(formatHours(player.stats.core.playtimeHours))}</strong></div>
        <div class="detail-fact"><span>Egyenleg</span><strong>${esc(new Intl.NumberFormat('hu-HU').format(player.stats.economy.balance))} coin</strong></div>
        <div class="detail-fact"><span>Farm szint</span><strong>${esc(player.stats.farm.level)}. szint</strong></div>
      </div>
    </div>

    <div class="detail-stats">
      <div class="detail-section-head">
        <h3>Statisztikák</h3>
        <span class="chip chip-neutral">${icon('activity', 13)} ${esc(getCurrentServer()?.name || 'Szerver')}</span>
      </div>
      ${statModulesHtml(player)}
    </div>
  </div>`;
}

export async function render(ctx) {
  const server = getCurrentServer();
  const query = (ctx.query.get('q') || '').trim();
  const players = await api.listPlayers({ query, serverId: server.id });
  const selected = players[0] || null;
  ctx.playersInitial = { players, selectedName: selected?.username || '' };

  return `<div class="page page-players">
    <section class="panel search-panel">
      <div class="search-field">
        <span class="search-icon">${icon('search', 20)}</span>
        <input type="search" id="player-search" placeholder="Játékos keresése név alapján…" autocomplete="off" value="${esc(query)}" aria-label="Játékos keresése">
        <button type="button" class="search-clear" id="search-clear" aria-label="Keresés törlése" ${query ? '' : 'hidden'}>${icon('x', 16)}</button>
        <kbd class="search-kbd">/</kbd>
      </div>
      <div class="search-meta">
        <span>${icon('users', 15)} <strong id="player-count">${players.length}</strong> játékos</span>
        <span class="server-inline">${icon('server', 15)} ${esc(server?.name || '')}</span>
      </div>
    </section>

    <div class="players-layout">
      <section class="panel players-list-panel">
        <div class="panel-head">
          <div>
            <span class="panel-eyebrow">Címtár</span>
            <h3>Játékosok</h3>
          </div>
          <span class="chip chip-green">${icon('activity', 13)} ${players.filter((player) => player.online).length} online</span>
        </div>
        <div class="player-list" id="player-list">${listHtml(players, selected?.username)}</div>
      </section>

      <section class="panel player-detail-panel" id="player-detail" aria-live="polite">
        ${detailHtml(selected)}
      </section>
    </div>
  </div>`;
}

export function mount(root, ctx) {
  const server = getCurrentServer();
  const initial = ctx.playersInitial || { players: [], selectedName: '' };
  const search = root.querySelector('#player-search');
  const clear = root.querySelector('#search-clear');
  const list = root.querySelector('#player-list');
  const detail = root.querySelector('#player-detail');
  const count = root.querySelector('#player-count');
  const onlineChip = root.querySelector('.players-list-panel .panel-head .chip');
  let players = initial.players;
  let selectedName = initial.selectedName;
  let detailViewer = null;
  let searchToken = 0;

  const destroyDetailViewer = () => {
    if (detailViewer) {
      detailViewer.destroy();
      detailViewer = null;
    }
  };

  const selectPlayer = (name) => {
    const player = players.find((entry) => entry.username === name) || null;
    selectedName = player ? player.username : '';
    list.querySelectorAll('.player-row').forEach((row) => {
      const isSelected = row.dataset.player === selectedName;
      row.classList.toggle('is-selected', isSelected);
      row.setAttribute('aria-pressed', String(isSelected));
    });
    destroyDetailViewer();
    detail.innerHTML = detailHtml(player);
    if (player) {
      const canvas = detail.querySelector('[data-detail-skin]');
      if (canvas) {
        detailViewer = new SkinViewer(canvas, { skinUrl: player.skin, autoRotate: true });
      }
    }
  };

  const applyPlayers = (nextPlayers) => {
    players = nextPlayers;
    if (count) count.textContent = String(players.length);
    if (onlineChip) onlineChip.innerHTML = `${icon('activity', 13)} ${players.filter((player) => player.online).length} online`;

    const stillVisible = players.some((player) => player.username === selectedName);
    const nextSelected = stillVisible ? selectedName : players[0]?.username || '';
    list.innerHTML = listHtml(players, nextSelected);
    selectPlayer(nextSelected);
  };

  const runSearch = async () => {
    const token = ++searchToken;
    const query = search.value.trim();
    if (clear) clear.hidden = !query;
    const nextPlayers = await api.listPlayers({ query, serverId: server.id });
    if (token !== searchToken) return;
    applyPlayers(nextPlayers);
    const url = query ? `/players?q=${encodeURIComponent(query)}` : '/players';
    window.history.replaceState({}, '', url);
  };

  const debouncedSearch = debounce(runSearch, 220);

  const onInput = () => {
    if (clear) clear.hidden = !search.value;
    debouncedSearch();
  };

  const onClear = () => {
    search.value = '';
    if (clear) clear.hidden = true;
    runSearch();
    search.focus();
  };

  const onListClick = (event) => {
    const row = event.target.closest('[data-player]');
    if (!row) return;
    selectPlayer(row.dataset.player);
  };

  const onKeyDown = (event) => {
    if (event.key === '/' && document.activeElement !== search && !/^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement.tagName)) {
      event.preventDefault();
      search.focus();
    }
    if (event.key === 'Escape' && document.activeElement === search && search.value) {
      onClear();
    }
  };

  search.addEventListener('input', onInput);
  clear?.addEventListener('click', onClear);
  list.addEventListener('click', onListClick);
  document.addEventListener('keydown', onKeyDown);

  const initialPlayer = players.find((player) => player.username === selectedName) || players[0];
  if (initialPlayer) {
    const canvas = detail.querySelector('[data-detail-skin]');
    if (canvas) detailViewer = new SkinViewer(canvas, { skinUrl: initialPlayer.skin, autoRotate: true });
  }

  return () => {
    debouncedSearch.cancel();
    destroyDetailViewer();
    search.removeEventListener('input', onInput);
    clear?.removeEventListener('click', onClear);
    list.removeEventListener('click', onListClick);
    document.removeEventListener('keydown', onKeyDown);
  };
}
