import { loadDb, saveDb, DEMO_SKIN, SEED_PASSWORD, mockHash } from './mock-db.js';

function skinHash(dataUrl) {
  const text = String(dataUrl || '');
  let hash = 2166136261;
  for (let i = 0; i < text.length; i += 8) {
    hash ^= text.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return `skh-${(hash >>> 0).toString(16)}`;
}

export function passwordStrength(password) {
  const p = String(password || '');
  return {
    minLength8: p.length >= 8,
    hasNumber: /\d/.test(p),
    hasUpper: /[A-Z]/.test(p),
    minLength10: p.length >= 10,
    hasSpecial: /[^A-Za-z0-9]/.test(p),
  };
}

export function passwordStrengthScore(password) {
  const s = passwordStrength(password);
  return Object.values(s).filter(Boolean).length;
}
import { storage } from '../core/storage.js';
import { sleep } from '../core/dom.js';
import { ROLES, ROLE_ORDER, roleDefinition } from '../core/permissions.js';

const SESSION_KEY = 'session';
const LATENCY = 220;
const clone = (value) => JSON.parse(JSON.stringify(value));

function getSession() {
  const session = storage.getJSON(SESSION_KEY);
  if (!session || typeof session.userId !== 'string') return null;
  return session;
}

function setSession(userId) {
  storage.setJSON(SESSION_KEY, { userId, issuedAt: new Date().toISOString() });
}

function findById(db, userId) {
  return db.users.find((user) => user.id === userId) || null;
}

function statsFor(db, userId, serverId) {
  const serverStats = db.stats[serverId] || {};
  return (
    serverStats[userId] || {
      core: { playtimeHours: 0, blocksMined: 0, mobsKilled: 0, deaths: 0 },
      economy: { balance: 0, earned: 0 },
      farm: { level: 0, harvested: 0 },
    }
  );
}

function selfProfile(user) {
  return {
    id: user.id,
    username: user.username,
    email: user.email,
    role: user.role,
    status: user.status,
    online: user.online,
    muted: user.muted,
    createdAt: user.createdAt,
    lastSeen: user.lastSeen,
    emailVerified: user.emailVerified,
    twoFactor: user.twoFactor,
    skin: user.skin || DEMO_SKIN,
    mustChangePassword: Boolean(user.mustChangePassword),
  };
}

function publicProfile(db, user, serverId) {
  return {
    id: user.id,
    username: user.username,
    role: user.role,
    status: user.status,
    online: user.online,
    muted: user.muted,
    createdAt: user.createdAt,
    lastSeen: user.lastSeen,
    skin: user.skin || DEMO_SKIN,
    stats: clone(statsFor(db, user.id, serverId)),
  };
}

function sortUsers(users) {
  return users.slice().sort((a, b) => {
    if (a.online !== b.online) return a.online ? -1 : 1;
    const rankDiff = roleDefinition(b.role).rank - roleDefinition(a.role).rank;
    if (rankDiff !== 0) return rankDiff;
    return a.username.localeCompare(b.username, 'hu');
  });
}

function pushAudit(db, actor, action, target, detail) {
  db.audit.unshift({
    id: `au-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    actor,
    action,
    target,
    detail,
    at: new Date().toISOString(),
  });
  db.audit = db.audit.slice(0, 60);
}

function currentActor(db) {
  const session = getSession();
  const user = session ? findById(db, session.userId) : null;
  return user ? user.username : 'Rendszer';
}

function mustFindUser(db, userId) {
  const user = findById(db, userId);
  if (!user) throw new Error('A felhasználó nem található.');
  return user;
}

export const api = {
  /** Synchronous so the router can guard routes before the first paint. */
  getCurrentUser() {
    const session = getSession();
    if (!session) return null;
    const db = loadDb();
    const user = findById(db, session.userId);
    if (!user || user.status === 'banned') {
      storage.remove(SESSION_KEY);
      return null;
    }
    return selfProfile(user);
  },

  async login(identifier, password) {
    await sleep(LATENCY);
    const db = loadDb();
    const needle = String(identifier || '').trim().toLowerCase();
    const user = db.users.find(
      (candidate) => candidate.username.toLowerCase() === needle || candidate.email.toLowerCase() === needle,
    );

    if (!user || user.passwordHash !== mockHash(String(password || ''))) {
      throw new Error('Hibás felhasználónév/e-mail vagy jelszó.');
    }
    if (user.status === 'banned') {
      throw new Error('Ez a fiók le van tiltva. Kérd a staff segítségét.');
    }

    user.lastSeen = new Date().toISOString();
    // mustChangePassword flag is preserved on the session profile; the UI checks it.
    saveDb(db);
    setSession(user.id);
    return selfProfile(user);
  },

  async register({ username, email, password }) {
    await sleep(LATENCY + 80);
    const db = loadDb();
    const cleanName = String(username || '').trim();
    const cleanEmail = String(email || '').trim().toLowerCase();

    if (!/^[A-Za-z0-9_]{3,16}$/.test(cleanName)) {
      throw new Error('A felhasználónév 3–16 karakter lehet, csak betű, szám és _ engedélyezett.');
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(cleanEmail)) {
      throw new Error('Érvényes e-mail címet adj meg.');
    }
    if (String(password || '').length < 8) {
      throw new Error('A jelszónak legalább 8 karakter hosszúnak kell lennie.');
    }
    if (db.users.some((user) => user.username.toLowerCase() === cleanName.toLowerCase())) {
      throw new Error('Ez a felhasználónév már foglalt.');
    }
    if (db.users.some((user) => user.email.toLowerCase() === cleanEmail)) {
      throw new Error('Ezzel az e-mail címmel már létezik fiók.');
    }

    const user = {
      id: `u-${Date.now().toString(36)}`,
      username: cleanName,
      email: cleanEmail,
      role: 'player',
      status: 'active',
      online: false,
      muted: false,
      createdAt: new Date().toISOString(),
      lastSeen: new Date().toISOString(),
      emailVerified: false,
      twoFactor: false,
      passwordHash: mockHash(password),
      skin: DEMO_SKIN,
    };
    db.users.push(user);

    const primaryServerId = 'alpha-main';
    db.stats[primaryServerId] = db.stats[primaryServerId] || {};
    db.stats[primaryServerId][user.id] = {
      core: { playtimeHours: 0, blocksMined: 0, mobsKilled: 0, deaths: 0 },
      economy: { balance: 0, earned: 0 },
      farm: { level: 0, harvested: 0 },
    };

    saveDb(db);
    setSession(user.id);
    return selfProfile(user);
  },

  async demoLogin(kind = 'player') {
    const username = kind === 'admin' || kind === 'owner' ? 'Aurora' : 'Steve';
    const user = await this.login(username, SEED_PASSWORD);
    return user;
  },

  async logout() {
    await sleep(120);
    storage.remove(SESSION_KEY);
  },

  async changePassword(oldPassword, newPassword) {
    await sleep(160);
    const db = loadDb();
    const session = getSession();
    if (!session) throw new Error('Nincs aktív munkamenet.');
    const user = mustFindUser(db, session.userId);
    if (user.passwordHash !== mockHash(String(oldPassword || ''))) {
      throw new Error('A jelenlegi jelszó hibás.');
    }
    if (String(newPassword || '').length < 8) {
      throw new Error('Az új jelszónak legalább 8 karakter hosszúnak kell lennie.');
    }
    user.passwordHash = mockHash(newPassword);
    user.mustChangePassword = false;
    saveDb(db);
    return selfProfile(user);
  },

  async listPlayers({ query = '', serverId = 'alpha-main' } = {}) {
    await sleep(140);
    const db = loadDb();
    const needle = String(query || '').trim().toLowerCase();
    const users = db.users.filter((user) => {
      if (user.status !== 'active') return false;
      if (!needle) return true;
      return user.username.toLowerCase().includes(needle);
    });
    return sortUsers(users).map((user) => publicProfile(db, user, serverId));
  },

  async getPlayer(username, serverId = 'alpha-main') {
    await sleep(100);
    const db = loadDb();
    const user = db.users.find(
      (candidate) =>
        candidate.username.toLowerCase() === String(username || '').toLowerCase() && candidate.status === 'active',
    );
    if (!user) return null;
    return publicProfile(db, user, serverId);
  },

  async getActivity() {
    await sleep(80);
    const db = loadDb();
    return clone(db.activity);
  },

  async updateSkin(dataUrl) {
    await sleep(160);
    if (typeof dataUrl !== 'string' || !dataUrl.startsWith('data:image/png;base64,')) {
      throw new Error('A skin feltöltése sikertelen.');
    }
    const db = loadDb();
    const session = getSession();
    if (!session) throw new Error('Nincs aktív munkamenet.');
    const user = mustFindUser(db, session.userId);
    const hash = skinHash(dataUrl);
    // Rejected skins are blocked for everyone: never apply them.
    if (db.skinBlacklist.includes(hash)) {
      // Reset anyone still carrying it (defensive) and refuse the upload.
      db.users.forEach((u) => { if (skinHash(u.skin) === hash) u.skin = DEMO_SKIN; });
      saveDb(db);
      throw new Error('Ez a skin el van utasítva, nem használható.');
    }
    // Auto-accept: the skin is applied immediately to the user and the server.
    user.skin = dataUrl;
    // If the skin is brand new (not yet whitelisted or blacklisted), it enters
    // the admin review queue. Approved skins become auto-accepted forever;
    // rejected skins are blocked for everyone.
    const known = db.skinWhitelist.includes(hash) || db.skinBlacklist.includes(hash);
    if (!known) {
      db.skinsQueue.unshift({
        id: `sk-${Date.now().toString(36)}`,
        userId: user.id,
        fileName: 'feltoltes.png',
        submittedAt: new Date().toISOString(),
        status: 'pending',
        skin: dataUrl,
        skinHash: hash,
      });
    }
    saveDb(db);
    return selfProfile(user);
  },

  admin: {
    async listUsers(serverId = 'alpha-main') {
      await sleep(120);
      const db = loadDb();
      const rank = (role) => roleDefinition(role).rank;
      return db.users
        .slice()
        .sort((a, b) => rank(b.role) - rank(a.role) || a.username.localeCompare(b.username, 'hu'))
        .map((user) => ({
          ...publicProfile(db, user, serverId),
          email: user.email,
          emailVerified: user.emailVerified,
          lastSeen: user.lastSeen,
          banReason: user.banReason || '',
          playtimeHours: statsFor(db, user.id, serverId).core.playtimeHours,
          twoFactor: user.twoFactor,
        }));
    },

    async summary() {
      const db = loadDb();
      return {
        total: db.users.length,
        active: db.users.filter((user) => user.status === 'active').length,
        banned: db.users.filter((user) => user.status === 'banned').length,
        online: db.users.filter((user) => user.online).length,
        pendingSkins: db.skinsQueue.filter((item) => item.status === 'pending').length,
        staff: db.users.filter((user) => user.role !== 'player').length,
      };
    },

    async setRole(userId, role) {
      await sleep(140);
      if (!ROLES[role]) throw new Error('Ismeretlen rang.');
      const db = loadDb();
      const user = mustFindUser(db, userId);
      user.role = role;
      pushAudit(db, currentActor(db), 'roles.assign', user.username, `${ROLES[role].label} rang kiosztva`);
      saveDb(db);
      return this.listUsers();
    },

    async banUser(userId, opts = {}) {
      await sleep(140);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      const duration = Number(opts.durationMinutes) || 0;
      const reason = String(opts.reason || '').trim() || 'Admin által eltiltva';
      const scope = opts.scope || 'both';
      const targets = Array.isArray(opts.targets) && opts.targets.length ? opts.targets : ['account'];
      const now = new Date();
      const expires = duration > 0 ? new Date(now.getTime() + duration * 60000).toISOString() : null;
      db.bans = Array.isArray(db.bans) ? db.bans : [];
      if (targets.includes('account')) {
        user.status = 'banned';
        user.online = false;
        user.banReason = reason;
        db.bans.push({ id: `bn-${Date.now().toString(36)}-a`, type: 'ban', scope, target: String(user.id), targetKind: 'account', reason, durationMinutes: duration, expiresAt: expires, createdAt: now.toISOString(), accountId: user.id, active: true });
      }
      if (targets.includes('hwid')) {
        db.bans.push({ id: `bn-${Date.now().toString(36)}-h`, type: 'ban', scope, target: `hwid:${user.id}`, targetKind: 'hwid', reason, durationMinutes: duration, expiresAt: expires, createdAt: now.toISOString(), accountId: user.id, active: true });
      }
      if (targets.includes('ip')) {
        db.bans.push({ id: `bn-${Date.now().toString(36)}-i`, type: 'ban', scope, target: `ip:${user.id}`, targetKind: 'ip', reason, durationMinutes: duration, expiresAt: expires, createdAt: now.toISOString(), accountId: user.id, active: true });
      }
      pushAudit(db, currentActor(db), 'players.ban', user.username, `${reason} (${scope}, ${targets.join(',')})`);
      saveDb(db);
      return this.listUsers();
    },

    async unbanUser(userId, targets) {
      await sleep(120);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      const t = Array.isArray(targets) && targets.length ? targets : ['account', 'hwid', 'ip'];
      if (t.includes('account')) {
        user.status = 'active';
        delete user.banReason;
      }
      db.bans = (db.bans || []).filter((b) => !(b.accountId === user.id && (b.targetKind === 'account' ? t.includes('account') : b.targetKind === 'hwid' ? t.includes('hwid') : t.includes('ip'))));
      pushAudit(db, currentActor(db), 'players.ban', user.username, 'Tiltás feloldva');
      saveDb(db);
      return this.listUsers();
    },

    async kickUser(userId, opts = {}) {
      await sleep(120);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      if (!user.online) throw new Error('A játékos nincs online.');
      const reason = String(opts.reason || '').trim() || 'Kirúgva a szerverről';
      const scope = opts.scope || 'both';
      const targets = Array.isArray(opts.targets) && opts.targets.length ? opts.targets : ['account'];
      user.online = false;
      db.bans = Array.isArray(db.bans) ? db.bans : [];
      db.bans.push({ id: `kc-${Date.now().toString(36)}`, type: 'kick', scope, target: String(user.id), targetKind: 'account', reason, durationMinutes: 0, expiresAt: null, createdAt: new Date().toISOString(), accountId: user.id, active: false });
      pushAudit(db, currentActor(db), 'players.kick', user.username, `${reason} (${scope}, ${targets.join(',')})`);
      saveDb(db);
      return this.listUsers();
    },

    async muteUser(userId, opts = {}) {
      await sleep(120);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      const duration = Number(opts.durationMinutes) || 0;
      const reason = String(opts.reason || '').trim() || 'Némítás bekapcsolva';
      const scope = opts.scope || 'both';
      const targets = Array.isArray(opts.targets) && opts.targets.length ? opts.targets : ['account'];
      const now = new Date();
      const expires = duration > 0 ? new Date(now.getTime() + duration * 60000).toISOString() : null;
      user.muted = true;
      db.bans = Array.isArray(db.bans) ? db.bans : [];
      db.bans.push({ id: `mu-${Date.now().toString(36)}`, type: 'mute', scope, target: String(user.id), targetKind: 'account', reason, durationMinutes: duration, expiresAt: expires, createdAt: now.toISOString(), accountId: user.id, active: true });
      pushAudit(db, currentActor(db), 'players.mute', user.username, `${reason} (${scope}, ${targets.join(',')})`);
      saveDb(db);
      return this.listUsers();
    },

    async unmuteUser(userId, targets) {
      await sleep(120);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      user.muted = false;
      const t = Array.isArray(targets) && targets.length ? targets : ['account', 'hwid', 'ip'];
      db.bans = (db.bans || []).filter((b) => !(b.accountId === user.id && b.type === 'mute' && (b.targetKind === 'account' ? t.includes('account') : b.targetKind === 'hwid' ? t.includes('hwid') : t.includes('ip'))));
      pushAudit(db, currentActor(db), 'players.mute', user.username, 'Némítás feloldva');
      saveDb(db);
      return this.listUsers();
    },

    async warnUser(userId, opts = {}) {
      await sleep(120);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      const reason = String(opts.reason || '').trim() || 'Figyelmeztetés';
      const scope = opts.scope || 'both';
      const targets = Array.isArray(opts.targets) && opts.targets.length ? opts.targets : ['account'];
      db.bans = Array.isArray(db.bans) ? db.bans : [];
      db.bans.push({ id: `wn-${Date.now().toString(36)}`, type: 'warn', scope, target: String(user.id), targetKind: 'account', reason, durationMinutes: Number(opts.durationMinutes) || 0, expiresAt: null, createdAt: new Date().toISOString(), accountId: user.id, active: false });
      pushAudit(db, currentActor(db), 'players.warn', user.username, `${reason} (${scope}, ${targets.join(',')})`);
      saveDb(db);
      return this.listUsers();
    },

    // Backwards-compatible aliases used by older callers.
    async setBanned(userId, banned) {
      if (banned) return this.banUser(userId, { targets: ['account'] });
      return this.unbanUser(userId, ['account']);
    },
    async setMuted(userId, muted) {
      if (muted) return this.muteUser(userId, { targets: ['account'] });
      return this.unmuteUser(userId, ['account']);
    },

    async deleteUser(userId) {
      await sleep(160);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      db.users = db.users.filter((candidate) => candidate.id !== userId);
      Object.values(db.stats).forEach((server) => {
        delete server[userId];
      });
      pushAudit(db, currentActor(db), 'accounts.delete', user.username, 'Fiók véglegesen törölve');
      saveDb(db);
      return this.listUsers();
    },

    async resetPassword(userId) {
      await sleep(160);
      const db = loadDb();
      const user = mustFindUser(db, userId);
      // Generate a random temporary password.
      const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%';
      let temp = '';
      for (let i = 0; i < 12; i += 1) {
        temp += chars[Math.floor(Math.random() * chars.length)];
      }
      user.passwordHash = mockHash(temp);
      user.mustChangePassword = true;
      pushAudit(db, currentActor(db), 'accounts.password-reset', user.username, 'Ideiglenes jelszó generálva');
      saveDb(db);
      return { tempPassword: temp, username: user.username };
    },

    async skinsQueue() {
      await sleep(100);
      const db = loadDb();
      return db.skinsQueue.map((item) => {
        const user = findById(db, item.userId);
        return { ...clone(item), username: user ? user.username : 'Törölt fiók' };
      });
    },

    async reviewSkin(id, decision) {
      await sleep(150);
      if (!['approved', 'rejected'].includes(decision)) throw new Error('Érvénytelen döntés.');
      const db = loadDb();
      const item = db.skinsQueue.find((entry) => entry.id === id);
      if (!item) throw new Error('A skin nem található.');
      item.status = decision;
      item.reviewedAt = new Date().toISOString();
      item.reviewedBy = currentActor(db);
      const hash = item.skinHash || skinHash(item.skin);
      item.skinHash = hash;
      const user = findById(db, item.userId);
      if (decision === 'approved') {
        if (!db.skinWhitelist.includes(hash)) db.skinWhitelist.push(hash);
        db.skinBlacklist = db.skinBlacklist.filter((h) => h !== hash);
        if (user) user.skin = item.skin;
      } else {
        if (!db.skinBlacklist.includes(hash)) db.skinBlacklist.push(hash);
        db.skinWhitelist = db.skinWhitelist.filter((h) => h !== hash);
        // Rejected: strip the skin from every player and reset to the default.
        db.users.forEach((u) => {
          if (skinHash(u.skin) === hash) u.skin = DEMO_SKIN;
        });
      }
      pushAudit(
        db,
        currentActor(db),
        'skins.approve',
        user ? user.username : 'ismeretlen',
        `${item.fileName} ${decision === 'approved' ? 'jóváhagyva' : 'elutasítva'}`,
      );
      saveDb(db);
      return this.skinsQueue();
    },

    async rejectedSkins() {
      await sleep(80);
      const db = loadDb();
      return db.skinBlacklist.map((hash) => ({ hash, skin: null }));
    },

    async removeFromBlacklist(hash) {
      await sleep(120);
      const db = loadDb();
      db.skinBlacklist = db.skinBlacklist.filter((h) => h !== hash);
      saveDb(db);
      return db.skinBlacklist;
    },

    async removeFromWhitelist(hash) {
      await sleep(120);
      const db = loadDb();
      db.skinWhitelist = db.skinWhitelist.filter((h) => h !== hash);
      saveDb(db);
      return db.skinWhitelist;
    },

    async whitelist() {
      await sleep(60);
      const db = loadDb();
      return [...db.skinWhitelist];
    },
    async blacklist() {
      await sleep(60);
      const db = loadDb();
      return [...db.skinBlacklist];
    },

    async audit() {
      await sleep(90);
      const db = loadDb();
      return clone(db.audit);
    },
  },
};

export const mockMeta = {
  demoPassword: SEED_PASSWORD,
  roleOrder: ROLE_ORDER,
};

export default api;
