import { config } from '../config.js';
import { storage } from './storage.js';

const SERVER_KEY = 'server';
const CHANGE_EVENT = 'alpha:server-change';
let currentId = null;
let initialized = false;

function enabledServers() {
  return (config.servers || []).filter((server) => server.enabled !== false);
}

export function getServers() {
  return enabledServers();
}

export function getServer(id) {
  return enabledServers().find((server) => server.id === id) || null;
}

export function getCurrentServer() {
  if (!initialized) initServerContext();
  const servers = enabledServers();
  return servers.find((server) => server.id === currentId) || servers[0] || null;
}

/**
 * Resolve the initial server. Launcher handoff wins over a stored choice:
 * `?server=alpha-main` or `?instance=alpha-main`. Unknown ids fall back to
 * the first enabled server instead of producing an empty dashboard.
 */
export function initServerContext() {
  if (initialized) return getCurrentServer();
  initialized = true;

  const params = new URLSearchParams(window.location.search);
  const handoff = params.get('server') || params.get('instance') || params.get('instanceId');
  const stored = storage.get(SERVER_KEY);
  const candidate = getServer(handoff) ? handoff : getServer(stored) ? stored : enabledServers()[0]?.id || null;

  currentId = candidate;
  if (candidate) storage.set(SERVER_KEY, candidate);
  return getCurrentServer();
}

export function setCurrentServer(id) {
  if (!getServer(id) || id === currentId) return false;
  currentId = id;
  storage.set(SERVER_KEY, id);
  window.dispatchEvent(new CustomEvent(CHANGE_EVENT, { detail: { server: getCurrentServer() } }));
  return true;
}

export function onServerChange(handler) {
  const listener = (event) => handler(event.detail.server);
  window.addEventListener(CHANGE_EVENT, listener);
  return () => window.removeEventListener(CHANGE_EVENT, listener);
}

export function serverSwitcherHtml() {
  const servers = enabledServers();
  if (servers.length < 2) return '';
  const current = getCurrentServer();
  return `
    <label class="server-switcher" title="Szerver kiválasztása">
      <span class="server-switcher-label">Szerver</span>
      <span class="select-wrap">
        <select data-server-select aria-label="Szerver kiválasztása">
          ${servers
            .map(
              (server) =>
                `<option value="${server.id}" ${server.id === current?.id ? 'selected' : ''}>${server.name}</option>`,
            )
            .join('')}
        </select>
        <svg class="select-caret" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="m6 9 6 6 6-6"/></svg>
      </span>
    </label>
  `;
}
