(function() {
    const STORAGE_KEY_AUTH = 'minecraft_launcher_auth';
    const STORAGE_KEY_LANG = 'lang';

    function getLang() {
        return localStorage.getItem(STORAGE_KEY_LANG) || 'hu';
    }

    function setLang(lang) {
        localStorage.setItem(STORAGE_KEY_LANG, lang);
        document.documentElement.setAttribute('lang', lang);
        document.documentElement.setAttribute('data-lang', lang);
        renderAll();
    }

    function t(key, params) {
        const lang = getLang();
        const keys = key.split('.');
        let value = I18N[lang];
        for (const k of keys) {
            if (value && typeof value === 'object' && k in value) {
                value = value[k];
            } else {
                return key;
            }
        }
        if (typeof value !== 'string') {
            return key;
        }
        if (params) {
            for (const [k, v] of Object.entries(params)) {
                value = value.replace(new RegExp(`{${k}}`, 'g'), v);
            }
        }
        return value;
    }

    function getAuth() {
        const raw = localStorage.getItem(STORAGE_KEY_AUTH);
        if (!raw) return null;
        try {
            return JSON.parse(raw);
        } catch (e) {
            return null;
        }
    }

    async function apiFetch(path, options = {}) {
        const auth = getAuth();
        const headers = { ...options.headers };
        if (auth && auth.accessToken) {
            headers['Authorization'] = `Bearer ${auth.accessToken}`;
        }
        const response = await fetch(path, {
            ...options,
            headers
        });
        if (response.status === 401) {
            localStorage.removeItem(STORAGE_KEY_AUTH);
            window.location.reload();
            throw new Error('Unauthorized');
        }
        return response;
    }

    function renderNav() {
        const lang = getLang();
        const auth = getAuth();
        const navDiv = document.getElementById('navLinks');
        if (!navDiv) return;

        let html = `
            <a href="/dashboard" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium">${t('nav.dashboard')}</a>
            <a href="/admin" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium">${t('nav.admin')}</a>
        `;
        if (auth) {
            html += `
                <span class="text-gray-400 text-sm mr-4">@${auth.username}</span>
                <button onclick="handleLogout()" class="text-red-400 hover:text-red-300 text-sm font-medium">${t('nav.logout')}</button>
            `;
        } else {
            html += `
                <a href="/skin-manager.html" class="bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded-md text-sm font-medium">Belépés / Regisztráció</a>
            `;
        }
        navDiv.innerHTML = html;
    }

    window.handleLogout = async function() {
        const auth = getAuth();
        if (auth && auth.refreshToken) {
            try {
                await apiFetch('/api/v1/auth/logout', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ refresh_token: auth.refreshToken })
                });
            } catch (e) {
                console.error(e);
            }
        }
        localStorage.removeItem(STORAGE_KEY_AUTH);
        window.location.href = '/';
    };

    window.loadUsers = async function() {
        const token = document.getElementById('adminToken')?.value;
        const search = document.getElementById('searchUser')?.value;
        if (!token) {
            alert(t('common.error'));
            return;
        }

        const tbody = document.getElementById('userRows');
        if (!tbody) return;

        tbody.innerHTML = `<tr><td colspan="4" class="text-center text-muted">${t('common.loading')}</td></tr>`;

        try {
            const response = await fetch(`/api/admin/users?q=${encodeURIComponent(search || '')}`, {
                headers: { 'X-Admin-Token': token }
            });
            if (!response.ok) {
                throw new Error('API error');
            }
            const data = await response.json();
            if (!data.users || data.users.length === 0) {
                tbody.innerHTML = `<tr><td colspan="4" class="text-center text-muted">${t('admin.noUsersFound')}</td></tr>`;
                return;
            }
            tbody.innerHTML = data.users.map(user => `
                <tr>
                    <td>${user.id}</td>
                    <td>${user.username}</td>
                    <td class="text-muted">${user.created_at ? new Date(user.created_at).toLocaleString(getLang() === 'hu' ? 'hu-HU' : 'en-US') : '-'}</td>
                    <td>
                        <button onclick="deleteUser(${user.id})" class="btn btn-danger text-xs">${t('common.delete')}</button>
                    </td>
                </tr>
            `).join('');
        } catch (e) {
            tbody.innerHTML = `<tr><td colspan="4" class="text-center text-red-400">${t('common.error')}: ${e.message}</td></tr>`;
        }
    };

    window.deleteUser = async function(id) {
        if (!confirm(t('admin.confirmDelete'))) return;
        const token = document.getElementById('adminToken')?.value;
        if (!token) return;

        try {
            await fetch(`/api/admin/users/${id}`, {
                method: 'DELETE',
                headers: { 'X-Admin-Token': token }
            });
            await loadUsers();
        } catch (e) {
            alert(t('common.error'));
        }
    };

    window.loadBans = async function() {
        const token = document.getElementById('adminToken')?.value;
        if (!token) {
            alert(t('common.error'));
            return;
        }

        const tbody = document.getElementById('banRows');
        if (!tbody) return;

        tbody.innerHTML = `<tr><td colspan="2" class="text-center text-muted">${t('common.loading')}</td></tr>`;

        try {
            const response = await fetch('/api/admin/hwid/bans', {
                headers: { 'X-Admin-Token': token }
            });
            if (!response.ok) {
                throw new Error('API error');
            }
            const data = await response.json();
            const bans = data.bans || [];
            if (bans.length === 0) {
                tbody.innerHTML = `<tr><td colspan="2" class="text-center text-muted">${t('admin.noBansFound')}</td></tr>`;
                return;
            }
            tbody.innerHTML = bans.map(deviceId => `
                <tr>
                    <td class="font-mono text-sm">${deviceId}</td>
                    <td><button onclick="unbanHWID('${deviceId}')" class="btn btn-primary text-xs">${t('common.unban')}</button></td>
                </tr>
            `).join('');
        } catch (e) {
            tbody.innerHTML = `<tr><td colspan="2" class="text-center text-red-400">${t('common.error')}: ${e.message}</td></tr>`;
        }
    };

    window.unbanHWID = async function(deviceId) {
        const token = document.getElementById('adminToken')?.value;
        if (!token) return;

        try {
            await fetch('/api/admin/hwid/unban', {
                method: 'POST',
                headers: { 'X-Admin-Token': token, 'Content-Type': 'application/json' },
                body: JSON.stringify({ device_id: deviceId })
            });
            await loadBans();
        } catch (e) {
            alert(t('common.error'));
        }
    };

    window.loadDashboard = async function() {
        const auth = getAuth();
        const content = document.getElementById('dashboardContent');

        if (!auth) {
            content.innerHTML = `
                <div class="text-center py-12">
                    <p class="text-muted mb-4">${t('dashboard.title')}: ${t('common.loading')}</p>
                    <a href="/skin-manager.html" class="btn btn-primary">Belépés</a>
                </div>
            `;
            return;
        }

        content.innerHTML = `
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="card">
                    <div class="card-header">${t('dashboard.welcome', { username: auth.username })}</div>
                    <div class="space-y-2 text-secondary">
                        <p>${t('dashboard.accountCreated')}: ${auth.createdAt ? new Date(auth.createdAt).toLocaleString(getLang() === 'hu' ? 'hu-HU' : 'en-US') : '-'}</p>
                    </div>
                </div>
                <div class="card">
                    <div class="card-header">${t('dashboard.skins')}</div>
                    <div id="skinsList" class="text-muted">${t('dashboard.noSkins')}</div>
                </div>
                <div class="card md:col-span-2">
                    <div class="card-header">${t('dashboard.sessions')}</div>
                    <div id="sessionsList">${t('dashboard.noSessions')}</div>
                </div>
            </div>
        `;

        try {
            const [skinsRes, sessionsRes] = await Promise.all([
                apiFetch('/api/v1/skins'),
                apiFetch('/api/v1/account/sessions')
            ]);

            const skinsData = await skinsRes.json();
            const skinsList = document.getElementById('skinsList');
            if (skinsData.skins && skinsData.skins.length > 0) {
                skinsList.innerHTML = skinsData.skins.map(skin => `
                    <div class="flex items-center justify-between py-2 border-b border-gray-700 last:border-0">
                        <span class="font-medium">${skin.variant}</span>
                        <span class="text-xs text-muted">${skin.content_type}</span>
                    </div>
                `).join('');
            } else {
                skinsList.innerHTML = `<p class="text-muted">${t('dashboard.noSkins')}</p>`;
            }

            const sessionsData = await sessionsRes.json();
            const sessionsList = document.getElementById('sessionsList');
            if (sessionsData.sessions && sessionsData.sessions.length > 0) {
                sessionsList.innerHTML = sessionsData.sessions.map(s => `
                    <div class="flex items-center justify-between py-2 border-b border-gray-700 last:border-0">
                        <div>
                            <span class="font-medium">${s.device_id || 'N/A'}</span>
                            <span class="text-xs text-muted"> | Utolsó használat: ${s.last_used_at ? new Date(s.last_used_at).toLocaleString(getLang() === 'hu' ? 'hu-HU' : 'en-US') : '-'}</span>
                        </div>
                        <span class="tag ${s.revoked ? 'tag-danger' : 'tag-success'}">${s.revoked ? 'Törölve' : 'Aktív'}</span>
                    </div>
                `).join('');
            } else {
                sessionsList.innerHTML = `<p class="text-muted">${t('dashboard.noSessions')}</p>`;
            }
        } catch (e) {
            console.error(e);
            const skinsList = document.getElementById('skinsList');
            const sessionsList = document.getElementById('sessionsList');
            if (skinsList) skinsList.innerHTML = `<p class="text-red-400">Hiba</p>`;
            if (sessionsList) sessionsList.innerHTML = `<p class="text-red-400">Hiba</p>`;
        }
    };

    function renderAll() {
        renderNav();
        const dashboardContent = document.getElementById('dashboardContent');
        if (dashboardContent) {
            loadDashboard();
        }
    }

    document.addEventListener('DOMContentLoaded', function() {
        renderAll();

        const langSwitcher = document.getElementById('langSwitcher');
        if (langSwitcher) {
            langSwitcher.value = getLang();
            langSwitcher.addEventListener('change', function() {
                setLang(this.value);
            });
        }
    });
})();
