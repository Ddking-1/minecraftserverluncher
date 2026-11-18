import { storage } from '../core/storage.js';

/**
 * Local alpha dataset.
 *
 * This is a UI prototype provider, not an authentication or authorization
 * authority. It lives entirely in the browser so the frontend can be reviewed
 * and extended before the matching backend endpoints are wired in. Passwords
 * are stored with a deliberately simple non-cryptographic mock hash; never
 * reuse this approach in production.
 */
export const DEMO_SKIN = '/assets/img/demo-skin.png';
export const SEED_PASSWORD = 'alpha1234';
const DB_KEY = 'mock-db';
const DB_VERSION = 1;

export function mockHash(password) {
  const text = `alpha-mock::${password}`;
  let hash = 2166136261;
  for (let i = 0; i < text.length; i += 1) {
    hash ^= text.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return (`00000000${(hash >>> 0).toString(16)}`).slice(-8);
}

const minutesAgo = (minutes) => new Date(Date.now() - minutes * 60000).toISOString();
const daysAgo = (days) => minutesAgo(days * 1440);

const SEED_USERS = [
  {
    id: 'u-aurora',
    username: 'Aurora',
    email: 'aurora@alpha.dev',
    role: 'owner',
    status: 'active',
    online: true,
    muted: false,
    createdAt: '2025-02-11T10:00:00.000Z',
    lastSeen: minutesAgo(2),
    emailVerified: true,
    twoFactor: true,
  },
  {
    id: 'u-steve',
    username: 'Steve',
    email: 'steve@alpha.dev',
    role: 'player',
    status: 'active',
    online: true,
    muted: false,
    createdAt: '2025-04-02T16:20:00.000Z',
    lastSeen: minutesAgo(1),
    emailVerified: true,
    twoFactor: false,
  },
  {
    id: 'u-nova',
    username: 'Nova',
    email: 'nova@alpha.dev',
    role: 'moderator',
    status: 'active',
    online: true,
    muted: false,
    createdAt: '2025-03-21T08:45:00.000Z',
    lastSeen: minutesAgo(4),
    emailVerified: true,
    twoFactor: true,
  },
  {
    id: 'u-kobold',
    username: 'Kobold',
    email: 'kobold@alpha.dev',
    role: 'support',
    status: 'active',
    online: false,
    muted: false,
    createdAt: '2025-06-08T12:10:00.000Z',
    lastSeen: minutesAgo(95),
    emailVerified: true,
    twoFactor: false,
  },
  {
    id: 'u-ender',
    username: 'Ender',
    email: 'ender@alpha.dev',
    role: 'player',
    status: 'active',
    online: true,
    muted: false,
    createdAt: '2025-01-19T19:05:00.000Z',
    lastSeen: minutesAgo(1),
    emailVerified: true,
    twoFactor: false,
  },
  {
    id: 'u-mira',
    username: 'Mira',
    email: 'mira@alpha.dev',
    role: 'player',
    status: 'active',
    online: false,
    muted: false,
    createdAt: '2025-07-14T14:30:00.000Z',
    lastSeen: minutesAgo(310),
    emailVerified: true,
    twoFactor: false,
  },
  {
    id: 'u-bruno',
    username: 'Bruno',
    email: 'bruno@alpha.dev',
    role: 'player',
    status: 'active',
    online: false,
    muted: true,
    createdAt: '2025-09-03T09:15:00.000Z',
    lastSeen: daysAgo(2),
    emailVerified: false,
    twoFactor: false,
  },
  {
    id: 'u-pixel',
    username: 'Pixel',
    email: 'pixel@alpha.dev',
    role: 'player',
    status: 'banned',
    online: false,
    muted: false,
    createdAt: '2025-08-22T17:40:00.000Z',
    lastSeen: daysAgo(5),
    emailVerified: true,
    twoFactor: false,
    banReason: 'Csapatás a farm szerveren',
  },
];

const SEED_STATS = {
  'u-aurora': { playtimeHours: 412.6, blocksMined: 184320, mobsKilled: 12944, deaths: 87, balance: 128450, earned: 342000, level: 42, harvested: 92311 },
  'u-steve': { playtimeHours: 268.4, blocksMined: 96540, mobsKilled: 8122, deaths: 143, balance: 64200, earned: 210500, level: 31, harvested: 61240 },
  'u-nova': { playtimeHours: 351.2, blocksMined: 142880, mobsKilled: 10451, deaths: 62, balance: 98750, earned: 288100, level: 38, harvested: 80450 },
  'u-kobold': { playtimeHours: 198.7, blocksMined: 71400, mobsKilled: 6420, deaths: 201, balance: 41200, earned: 154000, level: 24, harvested: 43210 },
  'u-ender': { playtimeHours: 512.9, blocksMined: 228100, mobsKilled: 15320, deaths: 118, balance: 183500, earned: 410000, level: 47, harvested: 121400 },
  'u-mira': { playtimeHours: 143.5, blocksMined: 52300, mobsKilled: 3980, deaths: 96, balance: 28900, earned: 98200, level: 19, harvested: 28450 },
  'u-bruno': { playtimeHours: 87.2, blocksMined: 31800, mobsKilled: 2410, deaths: 64, balance: 15800, earned: 64000, level: 12, harvested: 15120 },
  'u-pixel': { playtimeHours: 64.8, blocksMined: 21400, mobsKilled: 1820, deaths: 88, balance: 9400, earned: 41000, level: 9, harvested: 9800 },
};

function statRow(raw) {
  return {
    core: {
      playtimeHours: raw.playtimeHours,
      blocksMined: raw.blocksMined,
      mobsKilled: raw.mobsKilled,
      deaths: raw.deaths,
    },
    economy: { balance: raw.balance, earned: raw.earned },
    farm: { level: raw.level, harvested: raw.harvested },
  };
}

export function createSeedDb() {
  return {
    version: DB_VERSION,
    users: SEED_USERS.map((user) => ({
      ...user,
      passwordHash: mockHash(SEED_PASSWORD),
      skin: DEMO_SKIN,
    })),
    stats: {
      'alpha-main': Object.fromEntries(
        Object.entries(SEED_STATS).map(([userId, raw]) => [userId, statRow(raw)]),
      ),
    },
    skinsQueue: [
      {
        id: 'sk-1001',
        userId: 'u-mira',
        fileName: 'mira_sakura.png',
        submittedAt: minutesAgo(48),
        status: 'pending',
        skin: DEMO_SKIN,
      },
      {
        id: 'sk-1002',
        userId: 'u-bruno',
        fileName: 'bruno_miner.png',
        submittedAt: minutesAgo(210),
        status: 'pending',
        skin: DEMO_SKIN,
      },
      {
        id: 'sk-1003',
        userId: 'u-pixel',
        fileName: 'pixel_neon.png',
        submittedAt: daysAgo(1),
        status: 'pending',
        skin: DEMO_SKIN,
      },
    ],
    audit: [
      {
        id: 'au-1',
        actor: 'Aurora',
        action: 'roles.assign',
        target: 'Nova',
        detail: 'Moderátor rang kiosztva',
        at: minutesAgo(112),
      },
      {
        id: 'au-2',
        actor: 'Nova',
        action: 'players.ban',
        target: 'Pixel',
        detail: 'Csapatás a farm szerveren',
        at: minutesAgo(310),
      },
      {
        id: 'au-3',
        actor: 'Aurora',
        action: 'skins.approve',
        target: 'Mira',
        detail: 'mira_sakura.png jóváhagyva',
        at: daysAgo(1),
      },
      {
        id: 'au-4',
        actor: 'Kobold',
        action: 'players.mute',
        target: 'Bruno',
        detail: 'Csevegési korlátozás 24 órára',
        at: daysAgo(2),
      },
    ],
    activity: [
      { id: 'ac-1', kind: 'achievement', user: 'Ender', text: 'elérte a 47-es farm szintet', at: minutesAgo(18) },
      { id: 'ac-2', kind: 'join', user: 'Steve', text: 'csatlakozott az Alpha Main szerverhez', at: minutesAgo(34) },
      { id: 'ac-3', kind: 'skin', user: 'Mira', text: 'új skint küldött jóváhagyásra', at: minutesAgo(48) },
      { id: 'ac-4', kind: 'farm', user: 'Nova', text: '80 000 betakarított növénynél jár', at: minutesAgo(76) },
      { id: 'ac-5', kind: 'join', user: 'Aurora', text: 'belépett az irányítópultra', at: minutesAgo(112) },
    ],
  };
}

let cached = null;

export function loadDb() {
  if (cached) return cached;
  const stored = storage.getJSON(DB_KEY);
  if (stored && stored.version === DB_VERSION && Array.isArray(stored.users)) {
    cached = stored;
    return cached;
  }
  cached = createSeedDb();
  storage.setJSON(DB_KEY, cached);
  return cached;
}

export function saveDb(db) {
  cached = db;
  storage.setJSON(DB_KEY, db);
  return db;
}

/** Wipes local demo mutations and restores the seeded alpha dataset. */
export function resetMockDb() {
  cached = createSeedDb();
  storage.setJSON(DB_KEY, cached);
  return cached;
}
