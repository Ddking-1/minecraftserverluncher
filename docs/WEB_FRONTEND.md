# Web frontend foundation

Status: ALPHA, task branch `feat/web-ui-foundation`, not merged into `master`.
Entry point: `server/web/static/index.html`.
The Go server embeds the whole `server/web/static` directory and serves it as an
SPA (see `server/web/web.go`). The same files can be opened through a normal
static server during UI development; no Node, bundler or build step is needed.

## Goals and constraints

- Premium, dark-first gaming UI with a working light theme.
- Dedicated views and URLs, not one crowded long page.
- Qt 6 WebEngine friendly: no CDN requests, no remote fonts, no `eval`, no
  service worker, no build-time tooling, and a layout that survives launcher
  window sizes down to roughly 640 px wide.
- Alpha-friendly structure: vanilla ES modules, explicit registries and a
  replaceable data provider.
- Client-side checks are presentation only. The Go backend is the authority for
  identity and every mutation.

## Directory layout

```
server/web/static/
  index.html                 boot screen, early theme restore, module entry
  assets/css/
    tokens.css               light/dark design tokens (CSS variables)
    base.css                 reset, typography, accessibility helpers
    layout.css               app shell, sidebar, topbar, auth split
    components.css           buttons, cards, forms, tables, badges
    pages.css                view-specific grids and responsive rules
  assets/js/
    main.js                  route table, guards, shell lifecycle
    config.js                brand, servers, feature flags, data source
    core/
      router.js              history router, link interception, cleanup
      theme.js               dark/light switch + localStorage persistence
      storage.js             namespaced, failure-tolerant localStorage wrapper
      server-context.js      active server + launcher handoff + switcher
      permissions.js         roles and granular permission keys
      dom.js, icons.js, toast.js, modal.js
    data/
      mock-db.js             seeded alpha data, local persistence, demo hashes
      api.js                 async provider with the same shape as an HTTP client
    stats/
      registry.js            stat-module registration and leaderboards
      modules.js             core, economy and farm modules
    components/
      app-shell.js           sidebar, topbar, mobile drawer, user menu
      auth-layout.js         shared split layout for login/register
      skin-viewer.js         interactive isometric canvas skin renderer
      ui.js                  badges, avatars, empty states, small primitives
    views/
      login.js, register.js, dashboard.js, players.js, admin.js, not-found.js
  assets/img/                demo skins and SVG favicon
```

## Routes

| Route        | View          | Guard                                                     |
| ------------ | ------------- | --------------------------------------------------------- |
| `/login`     | Login         | redirects an authenticated session to `/dashboard`        |
| `/register`  | Registration  | guest only; separate form and page from login             |
| `/dashboard` | Dashboard     | requires a session                                       |
| `/players`   | Player search | requires a session                                       |
| `/admin`     | Admin panel   | requires `admin.access`; otherwise a 403 view             |
| other        | Not found     | rendered inside the shell                                 |

The Go handler returns `index.html` for `/`, the five app routes and unknown
extensionless paths, while missing files with an extension return 404. Legacy
`.html` bookmarks receive a 301 to the matching SPA route. Only `GET` and `HEAD`
are accepted on the SPA entry points.

## Theme and design

- `data-theme="dark|light"` is set on `<html>`.
- The choice is stored under `alpha.web.theme`; all frontend storage keys use the
  `alpha.web.` prefix.
- A tiny inline script in `index.html` restores the theme before CSS paints, so
  there is no light flash. It is wrapped in `try/catch` because a locked-down Qt
  profile can disable storage.
- Without a saved choice the OS `prefers-color-scheme` decides; the fallback is
  dark.
- The topbar button exposes the **action** ("Sötét mód bekapcsolása" /
  "Világos mód bekapcsolása"). It deliberately has no `aria-pressed` state,
  because the label already describes the next state.
- Tokens cover text, surfaces, borders, accent, semantic tones, shadows and
  radii. Views should use tokens instead of hard-coded colors.

## Multi-server preparation

`config.servers` currently contains exactly one enabled server, so the switcher
is hidden and every request is scoped to `alpha-main`. The context is already
implemented:

- `?server=<id>`, `?instance=<id>` or `?instanceId=<id>` preselects a server for
  a launcher handoff.
- The last valid choice is persisted as `alpha.web.server`.
- Unknown or disabled ids fall back to the first enabled server.
- `serverSwitcherHtml()` renders automatically when two or more servers are
  enabled. `onServerChange()` lets a view refresh its data.
- Stats, leaderboards and the mock database are keyed by server id.

Add a second object to `config.servers` with `enabled: true` to expose the
switcher. No dashboard markup change is required.

## Modular statistics

`stats/registry.js` is the extension point. A module declares its metadata,
metric definitions and optional leaderboard value:

```js
registerStatModule({
  id: 'fishing',
  title: 'Horgászat',
  icon: 'fish',
  description: 'Fogások és rekordok.',
  order: 40,
  metrics: [
    { key: 'caught', label: 'Kifogott hal', icon: 'fish', value: (stats) => stats.fishing.caught, format: number, hint: 'összesen' },
  ],
  leaderboard: { label: 'Fogások', value: (stats) => stats.fishing.caught, format: number, icon: 'fish' },
});
```

Views iterate over `getStatModules()` and `leaderboardModules()`, so a new plugin
dataset needs one registration plus the corresponding data field, not new page
markup. The seeded modules are core (playtime/mining/mobs/deaths), economy and
farm.

## Permissions

Roles and permissions live in `core/permissions.js`:

- `owner` — every permission, including `accounts.delete`.
- `admin` — account view, ban/kick/mute, skin approval, role assignment, audit;
  **cannot delete accounts** or manage the server.
- `moderator` — account view, ban/kick/mute, skin approval.
- `support` — account view and mute only.
- `player` — no staff permissions.

The admin view derives every button, select, panel and matrix column from
`hasPermission()` / `canAssignRole()`. Disallowed self-actions are disabled with
a reason. The rank simulator changes only the *effective presentation role*; it
does not grant server access. The actual user-role selector writes the mock role
and is also only a UI prototype.

**Security:** this model must be mirrored and enforced by the Go backend for
every endpoint. Hiding a button is not authorization.

## Data provider and alpha sessions

`data/api.js` exposes async methods for login, registration, profile updates,
skin updates, player search, server stats, leaderboards, activity and admin
operations. It currently reads and writes a local mock database in
`localStorage` (`alpha.web.mock-db`) and keeps a mock session under
`alpha.web.session`.

To connect the real backend, keep the method signatures and replace the bodies
with `fetch` calls. Planned endpoints are described in `docs/API.md`; the server
must issue/validate the session and must never trust a client-supplied user id,
role or permission list. The alpha demo credentials are intentionally public
(`Aurora` / `alpha1234`, and the login page's demo buttons) and must not ship as
real accounts.

## Skin handling and 3D preview

- Upload accepts a file picker or drag-and-drop on the dashboard dropzone.
- Client-side validation: PNG signature, 64x64 or 64x32 dimensions, maximum
  512 KB. The server must repeat the validation and decode the image safely.
- `components/skin-viewer.js` is a dependency-free isometric canvas renderer. It
  reads the skin's base, overlay, hat and jacket regions and paints a rotatable
  character with shading. This avoids a CDN/WebGL dependency that would break
  the `script-src 'self'` policy and offline Qt usage.
- Interaction: pointer drag, arrow keys when the canvas is focused, `R` reset,
  reset button, and an outer-layer toggle. Pointer capture is guarded because a
  synthetic or already-released pointer must not produce a console exception.
- A successful save writes the data URL into the mock profile and adds a pending
  item to the admin skin queue. The queue can be approved or rejected by staff;
  approval applies the skin to the mock user.

This is an alpha preview, not a full skinview3d replacement. Animation, capes
and slim/classic body switching are future work.

## Responsive behavior

Breakpoints in `pages.css` / `layout.css` cover 1320, 1180, 980, 760 and 540 px.
The sidebar becomes an off-canvas drawer at 1024 px and below; the topbar gets a
menu button and the user text collapses to an avatar. Auth split becomes a
single card. Dashboard stat modules go from four columns to two and then one. The skin
panel stacks, the admin simulator becomes one column and wide tables scroll
inside their own `.table-wrap` instead of widening the page.

The shared `.page` grid uses `grid-template-columns: minmax(0, 1fr)` so wide
tables and chip rows cannot force the document wider than the viewport. This was
found and fixed during 640 px testing.

## Verification performed (2026-11-18)

Browser: bundled Chromium through the preview panel, dev server on
`http://localhost:4173` (PowerShell static SPA fallback, because the old
`auth-server.exe` embeds the previous frontend).

- Visual review in dark and light mode: login, register, dashboard, players,
  admin, 403, role simulator, matrix, skin panel and admin tables.
- Viewports: 1280x720, 1024x700, 800x600 and 640x700. Checked the sidebar
  drawer, topbar, dashboard cards, skin panel, auth cards, players and admin
  tables. No horizontal document overflow after the `.page` fix.
- Found and fixed two responsive cascade issues: the shared `.page` grid track
  could be widened by admin tables, and the desktop menu buttons stayed visible
  because `.icon-btn { display: grid }` overrode their unscoped hide rules.
  Drawer controls are now scoped to `.topbar` / `.sidebar` and only appear at
  1024 px and below.
- Navigation through every route, guest/session redirects, logout, legacy URL
  redirects on the static dev server.
- Login: invalid password shows an inline alert; valid `Aurora` login creates a
  session and opens `/dashboard`.
- Registration: empty submit marks all fields; a valid registration creates the
  mock account, logs in and shows zeroed stats. Demo login works.
- Dashboard: upload of the alternate demo skin refreshes the preview immediately;
  save enables after upload and writes a queue item; drag/keyboard rotation,
  reset and outer-layer toggle change the rendered canvas.
- Players: debounced search (`ste`) filters the list; selecting a player renders
  details and a second interactive viewer.
- Admin: owner/admin/moderator/support/player simulations hide or disable
  actions as defined by the permission model; user search filters rows; role
  assignment and restore work; skin approve/reject removes pending items and
  writes audit entries.
- Console: no JavaScript exceptions or failed network requests in the final
  runs. Earlier synthetic pointer-capture exceptions were removed by guarding
  `setPointerCapture` / `releasePointerCapture`.

Not verified in this environment:

- `go build ./...` / `go test ./web/...` — the `go` toolchain is not on PATH on
  this machine, so `web.go` and `web_test.go` were reviewed but not compiled.
  The browser tests used the PowerShell static server, so the Go SPA handler,
  CSP middleware, 301/404/405 behavior and `go:embed` bundle were not exercised.
- Real 64x32 rendering and server-side PNG processing; the interactive test used
  the bundled 64x64 demo skins.
- Firefox/Safari and older Qt WebEngine versions; the available browser is
  Chromium-based.

## Security notes

- The alpha mock database, password hashes and session are localStorage-only and
  are **not** a security boundary. They must be removed or made dev-only before
  any production exposure.
- All rendering helpers escape untrusted text (`esc`). Do not interpolate raw
  usernames, file names or audit strings into markup.
- `safeNext()` only permits same-origin absolute paths and rejects `//host`
  redirects. Keep this check when adding post-login redirects.
- No external scripts, fonts or images are loaded. The existing Go CSP allows
  `'unsafe-inline'` for the early theme script and inline SVG/style attributes;
  a future hardening step can replace the inline script with a hash.
- Admin actions are UI simulations. The backend must authenticate the actor,
  check the granular permission, validate the target and record the audit entry.
- File uploads are untrusted input: enforce size, MIME/PNG signature, dimensions,
  decompression limits and output normalization server-side.

## Next steps

1. Wire `data/api.js` to the real Go endpoints and replace mock sessions with the
   backend's authenticated session.
2. Add backend authorization for the granular permission keys and audit logging.
3. Run `gofmt`, `go build ./...` and `go test ./web/...` from `server/` on a
   machine with Go installed; then verify the embedded SPA through the real Go
   handler and CSP.
4. Decide whether the next iteration needs slim/classic skin models, animation or
   capes; keep the canvas renderer if offline Qt support remains a requirement.
5. Merge the task branch after review; no production deployment is part of this
   task.
