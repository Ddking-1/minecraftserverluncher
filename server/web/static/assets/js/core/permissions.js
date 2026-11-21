/**
 * Granular permission model shared by the admin UI.
 *
 * This is a client-side presentation model for the alpha permission
 * simulator. It must never become the authority for real mutations: the
 * backend has to re-check every action server-side (see docs/SECURITY.md).
 */
export const PERMISSIONS = [
  { key: 'admin.access', label: 'Admin panel elérése', group: 'Általános', description: 'Az admin felület megnyitása.' },
  { key: 'accounts.view', label: 'Fiókok megtekintése', group: 'Fiókok', description: 'Felhasználói lista és alapadatok.' },
  { key: 'accounts.delete', label: 'Fiók törlése', group: 'Fiókok', description: 'Végleges, rendszerszintű törlés.' },
  { key: 'players.ban', label: 'Játékos bannolása', group: 'Játékosok', description: 'Fiók letiltása és visszaállítása.' },
  { key: 'players.kick', label: 'Játékos kirúgása', group: 'Játékosok', description: 'Online játékos eltávolítása.' },
  { key: 'players.mute', label: 'Játékos némítása', group: 'Játékosok', description: 'Csevegés korlátozása.' },
  { key: 'skins.approve', label: 'Skin jóváhagyás', group: 'Tartalom', description: 'Feltöltött skinek elfogadása vagy elutasítása.' },
  { key: 'roles.assign', label: 'Rangok kiosztása', group: 'Jogosultságok', description: 'Alacsonyabb rangok kiosztása.' },
  { key: 'server.manage', label: 'Szerver kezelése', group: 'Rendszer', description: 'Szerver-szintű beállítások módosítása.' },
  { key: 'audit.view', label: 'Audit napló', group: 'Rendszer', description: 'Adminisztratív műveletek naplója.' },
];

export const PERMISSION_GROUPS = ['Általános', 'Fiókok', 'Játékosok', 'Tartalom', 'Jogosultságok', 'Rendszer'];

export const ROLES = {
  owner: {
    id: 'owner',
    label: 'Tulajdonos',
    short: 'OWNER',
    rank: 5,
    tone: 'gold',
    description: 'Teljes rendszerhozzáférés, beleértve a fióktörlést.',
    permissions: ['*'],
  },
  admin: {
    id: 'admin',
    label: 'Admin',
    short: 'ADMIN',
    rank: 4,
    tone: 'green',
    description: 'Széles körű kezelés, de fiókot nem törölhet.',
    permissions: [
      'admin.access',
      'accounts.view',
      'players.ban',
      'players.kick',
      'players.mute',
      'skins.approve',
      'roles.assign',
      'audit.view',
    ],
  },
  moderator: {
    id: 'moderator',
    label: 'Moderátor',
    short: 'MOD',
    rank: 3,
    tone: 'blue',
    description: 'Játékosmoderáció és skin jóváhagyás.',
    permissions: ['admin.access', 'accounts.view', 'players.ban', 'players.kick', 'players.mute', 'skins.approve'],
  },
  support: {
    id: 'support',
    label: 'Support',
    short: 'SUPPORT',
    rank: 2,
    tone: 'violet',
    description: 'Ügyfélszolgálati betekintés, némítási joggal.',
    permissions: ['admin.access', 'accounts.view', 'players.mute'],
  },
  player: {
    id: 'player',
    label: 'Játékos',
    short: 'PLAYER',
    rank: 1,
    tone: 'slate',
    description: 'Normál játékosi hozzáférés.',
    permissions: [],
  },
};

export const ROLE_ORDER = ['owner', 'admin', 'moderator', 'support', 'player'];

export function roleDefinition(role) {
  return ROLES[role] || ROLES.player;
}

export function permissionsFor(role) {
  const definition = roleDefinition(role);
  if (definition.permissions.includes('*')) return PERMISSIONS.map((permission) => permission.key);
  return [...definition.permissions];
}

export function hasPermission(actor, permissionKey) {
  const role = typeof actor === 'string' ? actor : actor && actor.role;
  const definition = roleDefinition(role);
  if (definition.permissions.includes('*')) return true;
  return definition.permissions.includes(permissionKey);
}

export function canAssignRole(actorRole, targetRole) {
  if (!hasPermission(actorRole, 'roles.assign')) return false;
  if (!ROLES[targetRole]) return false;
  if (actorRole === 'owner') return true;
  if (actorRole === 'admin') return ['moderator', 'support', 'player'].includes(targetRole);
  return false;
}

export function assignableRoles(actorRole) {
  return ROLE_ORDER.filter((role) => canAssignRole(actorRole, role));
}
