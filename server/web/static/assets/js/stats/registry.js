/**
 * Statistic module registry.
 *
 * A new plugin dataset (economy, farm top lists, playtime, ...) is added by
 * registering one module object. Views render whatever is registered, so no
 * dashboard/player markup changes are required for a new metric group.
 *
 * Module shape:
 * {
 *   id: 'economy',
 *   title: 'Gazdaság',
 *   icon: 'coins',
 *   description: '...',
 *   metrics: [{ key, label, icon, value(stats), format(value), hint }],
 *   leaderboard?: { label, value(stats), format(value), icon }
 * }
 */
const modules = new Map();

export function registerStatModule(definition) {
  if (!definition || !definition.id) throw new Error('Stat module needs an id');
  modules.set(definition.id, definition);
  return definition;
}

export function getStatModules() {
  return Array.from(modules.values()).sort((a, b) => (a.order || 0) - (b.order || 0));
}

export function getStatModule(id) {
  return modules.get(id) || null;
}

export function leaderboardModules() {
  return getStatModules().filter((module) => module.leaderboard);
}

/**
 * @param {object} module registered stat module
 * @param {Array} players public player objects that carry `stats`
 */
export function buildLeaderboard(module, players, limit = 5) {
  if (!module || !module.leaderboard) return [];
  const valueOf = module.leaderboard.value;
  return players
    .map((player) => ({ player, value: Number(valueOf(player.stats)) }))
    .filter((entry) => Number.isFinite(entry.value))
    .sort((a, b) => b.value - a.value)
    .slice(0, limit);
}
