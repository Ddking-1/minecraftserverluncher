import { formatCoins, formatHours, formatNumber } from '../core/dom.js';
import { registerStatModule } from './registry.js';

/**
 * Built-in alpha metric groups. Keep each group independent: the registry is
 * the only place views know about, so adding a fourth group is a one-file
 * change (register it below, or from a future plugin adapter).
 */

registerStatModule({
  id: 'core',
  order: 10,
  title: 'Alap statisztikák',
  icon: 'activity',
  description: 'Játékidő, bányászat és harc.',
  metrics: [
    {
      key: 'playtime',
      label: 'Játszott idő',
      icon: 'clock',
      hint: 'összesen',
      value: (stats) => stats.core.playtimeHours,
      format: formatHours,
    },
    {
      key: 'blocks',
      label: 'Kibányászott blokk',
      icon: 'pickaxe',
      hint: 'összesen',
      value: (stats) => stats.core.blocksMined,
      format: formatNumber,
    },
    {
      key: 'mobs',
      label: 'Megölt mob',
      icon: 'slash',
      hint: 'összesen',
      value: (stats) => stats.core.mobsKilled,
      format: formatNumber,
    },
    {
      key: 'deaths',
      label: 'Halálozás',
      icon: 'alert',
      hint: 'összesen',
      value: (stats) => stats.core.deaths,
      format: formatNumber,
    },
  ],
  leaderboard: {
    label: 'Játékidő',
    icon: 'clock',
    value: (stats) => stats.core.playtimeHours,
    format: formatHours,
  },
});

registerStatModule({
  id: 'economy',
  order: 20,
  title: 'Gazdaság',
  icon: 'coins',
  description: 'A szerver gazdasági pluginjának adatai.',
  metrics: [
    {
      key: 'balance',
      label: 'Egyenleg',
      icon: 'coins',
      hint: 'jelenlegi',
      value: (stats) => stats.economy.balance,
      format: formatCoins,
    },
    {
      key: 'earned',
      label: 'Összes bevétel',
      icon: 'activity',
      hint: 'összesen',
      value: (stats) => stats.economy.earned,
      format: formatCoins,
    },
  ],
  leaderboard: {
    label: 'Vagyon',
    icon: 'coins',
    value: (stats) => stats.economy.balance,
    format: formatCoins,
  },
});

registerStatModule({
  id: 'farm',
  order: 30,
  title: 'Farm',
  icon: 'wheat',
  description: 'Farm plugin szintek és betakarítási toplista.',
  metrics: [
    {
      key: 'level',
      label: 'Farm szint',
      icon: 'sparkles',
      hint: 'jelenlegi',
      value: (stats) => stats.farm.level,
      format: (value) => `${formatNumber(value)}. szint`,
    },
    {
      key: 'harvested',
      label: 'Betakarítás',
      icon: 'wheat',
      hint: 'összesen',
      value: (stats) => stats.farm.harvested,
      format: formatNumber,
    },
  ],
  leaderboard: {
    label: 'Betakarítás',
    icon: 'wheat',
    value: (stats) => stats.farm.harvested,
    format: formatNumber,
  },
});
