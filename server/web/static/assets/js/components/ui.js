import { esc, initials, nameHue } from '../core/dom.js';
import { icon } from '../core/icons.js';
import { roleDefinition } from '../core/permissions.js';
import { getStatModules } from '../stats/registry.js';

const ROLE_ICON = {
  owner: 'crown',
  admin: 'shield',
  moderator: 'shield',
  support: 'user',
  player: 'user',
};

export function roleBadge(role, { short = false } = {}) {
  const definition = roleDefinition(role);
  return `<span class="badge badge-role tone-${definition.tone}" title="${esc(definition.description)}">
    ${icon(ROLE_ICON[definition.id] || 'user', 12)}
    <span>${esc(short ? definition.short : definition.label)}</span>
  </span>`;
}

export function statusDot(online, label = '') {
  const text = label || (online ? 'Online' : 'Offline');
  return `<span class="status ${online ? 'is-online' : 'is-offline'}"><span class="status-dot"></span>${esc(text)}</span>`;
}

export function avatar(user, size = 'md') {
  const name = typeof user === 'string' ? user : user?.username || '?';
  const skin = typeof user === 'object' ? user?.skin : '';
  const hue = nameHue(name);
  if (!skin) {
    return `<span class="avatar avatar-${size} avatar-fallback" style="--avatar-hue:${hue}" aria-hidden="true"><span>${esc(initials(name))}</span></span>`;
  }
  const safeSrc = esc(skin);
  return `<span class="avatar avatar-${size}" style="--avatar-hue:${hue}" aria-hidden="true">
    <img class="avatar-layer avatar-base" src="${safeSrc}" alt="" loading="lazy" draggable="false">
    <img class="avatar-layer avatar-hat" src="${safeSrc}" alt="" loading="lazy" draggable="false">
  </span>`;
}

export function statModulesHtml(player) {
  return getStatModules()
    .map((module) => {
      const cards = module.metrics
        .map((metric) => {
          const value = metric.value(player.stats);
          return `<article class="stat-card" data-metric="${esc(metric.key)}">
            <span class="stat-icon">${icon(metric.icon || module.icon, 18)}</span>
            <div class="stat-copy">
              <span class="stat-label">${esc(metric.label)}</span>
              <strong class="stat-value">${esc(metric.format(value))}</strong>
              ${metric.hint ? `<span class="stat-hint">${esc(metric.hint)}</span>` : ''}
            </div>
          </article>`;
        })
        .join('');

      return `<section class="stat-group" data-stat-module="${esc(module.id)}">
        <div class="stat-group-head">
          <span class="icon-box">${icon(module.icon, 18)}</span>
          <div>
            <h3>${esc(module.title)}</h3>
            <p>${esc(module.description || '')}</p>
          </div>
        </div>
        <div class="stat-grid">${cards}</div>
      </section>`;
    })
    .join('');
}

export function leaderboardRows(entries, { currentUsername = '', format = String } = {}) {
  if (!entries.length) {
    return `<div class="lb-empty">${icon('activity', 18)}<span>Még nincs elérhető adat.</span></div>`;
  }
  return entries
    .map((entry, index) => {
      const player = entry.player;
      const isCurrent = player.username === currentUsername;
      return `<a class="lb-row ${isCurrent ? 'is-current' : ''}" href="/players?q=${encodeURIComponent(player.username)}" data-link>
        <span class="lb-rank rank-${index + 1}">${index + 1}</span>
        ${avatar(player, 'xs')}
        <span class="lb-name">${esc(player.username)}${isCurrent ? '<span class="lb-you">te</span>' : ''}</span>
        <span class="lb-value">${esc(format(entry.value))}</span>
      </a>`;
    })
    .join('');
}

export function emptyState({ iconName = 'search', title = 'Nincs találat', text = '', action = '' } = {}) {
  return `<div class="empty-state">
    <span class="empty-icon">${icon(iconName, 24)}</span>
    <h3>${esc(title)}</h3>
    <p>${esc(text)}</p>
    ${action}
  </div>`;
}

export function spinner(size = 16) {
  return `<span class="spinner" style="--spinner-size:${size}px" aria-hidden="true"></span>`;
}

export function chip(label, tone = 'neutral', iconName = '') {
  return `<span class="chip chip-${tone}">${iconName ? icon(iconName, 13) : ''}${esc(label)}</span>`;
}
