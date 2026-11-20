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
      // Resolve backend RBAC roles per user so the role selector reflects reality.
      const withRoles = await Promise.all(users.map(async (u) => {
        try {
          const r = await req(`/api/admin/users/${u.id}/roles`);
          const roles = (r && r.roles) || [];
          return { ...u, _roles: roles };
        } catch (e) {
          return { ...u, _roles: [] };
        }
      }));
      return withRoles.map((u) => {
        const mapped = mapUser(u);
        if (Array.isArray(u._roles) && u._roles.includes('admin')) mapped.role = 'admin';
        return mapped;
      });
    },
    async summary() {
      const users = await this.listUsers();
      const banned = users.filter((u) => u.status === 'banned').length;
      let pendingSkins = 0;
      try { const q = await this.skinsQueue(); pendingSkins = q.length; } catch (e) {}
      return {
        total: users.length,
        staff: users.filter((u) => u.role !== 'player').length,
        active: users.length - banned,
        online: 0,
        banned,
        pendingSkins,
      };
    },
    async skinsQueue() {
      const data = await req('/api/admin/skins/pending');
      const pending = (data && data.skins) || [];
      let users = [];
      try { users = await this.listUsers(); } catch (e) {}
      const byId = new Map(users.map((u) => [String(u.id), u]));
      return pending.map((p) => {
        const ownerId = String(p.skin_id || '').split('-')[0];
        const user = byId.get(ownerId);
        return {
          id: p.skin_id,
          userId: ownerId,
          username: user ? user.username : 'ismeretlen',
          fileName: 'skin.png',
          submittedAt: p.reviewed_at || new Date().toISOString(),
          status: 'pending',
          skin: `/api/v1/skin-files/${p.skin_id}`,
        };
      });
    },
    async blacklist() { return []; },
    async audit() {
      const data = await req('/api/admin/audit?limit=50');
      const entries = (data && data.entries) || [];
      return entries.map((e) => ({
        actor: e.actor || 'Admin',
        action: e.action || 'audit',
        target: e.target != null ? String(e.target) : '',
        detail: e.detail || '',
        at: e.created_at || new Date().toISOString(),
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
    async setRole(userId, role) {
      const isAdminRole = role !== 'player';
      if (isAdminRole) {
        return req(`/api/admin/users/${userId}/roles`, {
          method: 'POST',
          headers: jsonHeaders(),
          body: JSON.stringify({ role: 'admin' }),
        });
      }
      try {
        await req(`/api/admin/users/${userId}/roles/admin`, { method: 'DELETE', headers: headers() });
      } catch (e) { /* ignore */ }
      return { status: 'revoked' };
    },
    async reviewSkin(id, decision) {
      return req(`/api/admin/skins/${id}/approve`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify({ approved: decision === 'approved' }),
      });
    },
    async removeFromBlacklist() { return; },
    async resetPassword(userId) {
      const data = await req(`/api/admin/users/${userId}/reset-password`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify({}),
      });
      return { username: data.username, tempPassword: data.temp_password };
    },
    async userDevices(userId) {
      const data = await req(`/api/admin/users/${userId}/details`);
      return data;
    },
    async ipAccounts(ip) {
      return req(`/api/admin/ip-accounts?ip=${encodeURIComponent(ip)}`);
    },
  };
}
