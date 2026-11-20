// Real backend admin API adapter. Mirrors the mock api.admin surface so the
// admin view can switch between mock and real backends without changing its
// call sites. Uses X-Admin-Token header.
export function createRealAdminApi(token) {
  const headers = () => ({ 'X-Admin-Token': token || '' });
  const jsonHeaders = () => ({ 'Content-Type': 'application/json', 'X-Admin-Token': token || '' });

  const req = async (path, opts = {}) => {
    const res = await fetch(path, {
      ...opts,
      headers: { ...(opts.headers || {}), ...headers() },
    });
    const text = await res.text();
    let data = null;
    try { data = text ? JSON.parse(text) : null; } catch (e) { data = text; }
    if (!res.ok) {
      const msg = (data && (data.error || data.message)) || res.statusText || 'request failed';
      throw new Error(msg);
    }
    return data;
  };

  const mapUser = (u) => ({
    id: String(u.id),
    username: u.username || '',
    email: '',
    role: 'player',
    status: u.banned ? 'banned' : (u.minecraft_banned ? 'banned' : 'active'),
    online: true,
    muted: !!u.muted,
    lastSeen: u.created_at || new Date().toISOString(),
    banReason: '',
    playtimeHours: 0,
    _real: true,
    _mcBanned: !!u.minecraft_banned,
    _approved: !!u.approved,
  });

  return {
    async listUsers() {
      const data = await req('/api/admin/users');
      const users = (data && data.users) || [];
      return users.map(mapUser);
    },
    async summary() {
      const users = await this.listUsers();
      const banned = users.filter((u) => u.status === 'banned').length;
      return {
        total: users.length,
        staff: 0,
        active: users.length - banned,
        online: 0,
        banned,
        pendingSkins: 0,
      };
    },
    async skinsQueue() { return []; },
    async blacklist() { return []; },
    async audit() {
      const data = await req('/api/admin/bans?limit=20');
      const bans = (data && data.bans) || [];
      return bans.map((b) => ({
        actor: 'Admin',
        action: b.type || b.ban_type || 'ban',
        target: (b.account_id != null ? String(b.account_id) : (b.target || '')),
        detail: `${b.scope || ''} · ${b.reason || ''}`.trim(),
        at: b.created_at || new Date().toISOString(),
      }));
    },
    async banUser(userId, params) {
      const body = {
        duration_minutes: params.durationMinutes || 0,
        reason: params.reason || '',
        scope: params.scope || 'both',
        targets: params.targets || ['account'],
        lookback_days: params.lookbackDays || 30,
      };
      if (params.ip) body.ip = params.ip;
      if (params.hwid) body.hwid = params.hwid;
      return req(`/api/admin/users/${userId}/ban`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify(body),
      });
    },
    async unbanUser(userId, targets) {
      return req(`/api/admin/users/${userId}/unban`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify({ targets: targets || ['account', 'hwid', 'ip'] }),
      });
    },
    async kickUser(userId, params) {
      return req(`/api/admin/users/${userId}/kick`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify({
          duration_minutes: params.durationMinutes || 0,
          reason: params.reason || '',
          scope: params.scope || 'both',
          targets: params.targets || ['account'],
          lookback_days: params.lookbackDays || 30,
        }),
      });
    },
    async muteUser(userId, params) {
      return req(`/api/admin/users/${userId}/mute`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify({
          duration_minutes: params.durationMinutes || 0,
          reason: params.reason || '',
          scope: params.scope || 'both',
          targets: params.targets || ['account'],
          lookback_days: params.lookbackDays || 30,
        }),
      });
    },
    async unmuteUser(userId, targets) {
      return req(`/api/admin/users/${userId}/unmute`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify({ targets: targets || ['account', 'hwid', 'ip'] }),
      });
    },
    async warnUser(userId, params) {
      return req(`/api/admin/users/${userId}/warn`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify({
          duration_minutes: params.durationMinutes || 0,
          reason: params.reason || '',
          scope: params.scope || 'both',
          targets: params.targets || ['account'],
          lookback_days: params.lookbackDays || 30,
        }),
      });
    },
    async deleteUser(userId) {
      return req(`/api/admin/users/${userId}`, { method: 'DELETE', headers: headers() });
    },
    async setRole() { throw new Error('Role management not available in real backend mode'); },
    async reviewSkin() { return []; },
    async removeFromBlacklist() { return; },
    async resetPassword() { throw new Error('Not available'); },
    async userDevices(userId) {
      const data = await req(`/api/admin/users/${userId}/details`);
      return data;
    },
    async ipAccounts(ip) {
      return req(`/api/admin/ip-accounts?ip=${encodeURIComponent(ip)}`);
    },
  };
}
