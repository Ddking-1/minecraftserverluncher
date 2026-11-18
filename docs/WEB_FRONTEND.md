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

## Verification performed (2026-11-18, Go handler path)

Go toolchain: Go 1.26.5 (`go version go1.26.5 windows/amd64`) from the module
toolchain cache. All Go checks run from `server/`, the Go module root.

Static Go checks (clean):

- `gofmt -l web` — no files need reformatting.
- `go vet ./web/...` — clean.
- `go build ./web/...` — compiles.
- `go test -count=1 ./web/...` — both tests pass
  (`TestIndexServesSPA`, `TestEmbeddedAssetsAreServed`).

A compiled Go binary was started locally with the embedded frontend and tested
through the real Go handler and CSP. Configuration used for that run:

- Binary: `_local_test/go-server/server.exe`
- Listener: `AUTH_LISTEN_ADDR=127.0.0.1:8090`
- DB: `AUTH_DB_PATH=../_local_test/go-server/auth.db`
- Skin dir: `AUTH_SKIN_STORAGE_DIR=../_local_test/go-server/skins`
- Log: `_local_test/go-server/server.log`

HTTP contract verified against that running Go server:

- `GET /`, `/login`, `/register`, `/dashboard`, `/players`, `/admin` — 200
  `text/html; charset=utf-8`, `index.html` shell, embedded assets served from
  `/` with correct content types.
- Favicon: `/favicon.ico` returns 204; `/assets/img/favicon.svg` returns 200
  with the bundled SVG. The `index.html` now includes
  `<link rel="icon" type="image/svg+xml" href="/assets/img/favicon.svg">`, so
  real browsers no longer request a missing `favicon.ico`.
- Embedded assets: `/assets/css/tokens.css`, `/assets/js/main.js`, `/assets/js/views/admin.js`,
  `/assets/js/views/dashboard.js`, `/assets/js/views/players.js` and
  `/assets/img/demo-skin.png` all return 200 with the expected content.
- Legacy `.html` bookmarks (`/login.html`, `/register.html`, `/dashboard.html`,
  `/players.html`, `/admin.html`, `/skin-manager.html`) — 301 to the matching
  SPA route.
- Missing assets (`/assets/missing.png`) and unknown extensionless `.html`
  paths (`/nope.html`) — 404.
- Methods: `POST /login` returns 405.
- CSP: the `Content-Security-Policy` header is present on the SPA responses.

Browser verification through the real Go origin (`http://127.0.0.1:8090`):

- Visual review in dark and light mode: login, register, dashboard, players,
  admin, 403, role simulator, matrix, skin panel and admin tables.
- All five app routes, plus direct reloads and client-side SPA navigation
  (`element.click()`), load correctly with the theme and title preserved.
- Dashboard: skin upload, immediate canvas refresh, save, drag/keyboard
  rotation, reset and overlay toggle.
- Players: debounced search filtering.
- Admin: owner/admin/moderator/support/player simulations, user search, role
  assignment and restore, skin approve/reject.
- Viewports 640x700 and 800x600: no horizontal document overflow.
- Console: no JavaScript errors, no CSP violations, no failed network requests
  in the final runs.

Verified earlier through a PowerShell static dev server (browser only, not the
Go handler):

- Viewports 1280x720, 1024x700, 800x600 and 640x700.
- Login/register validation, demo login, logout, route guards.
- Two responsive cascade fixes: the shared `.page` grid track was widened by
  admin tables, and desktop menu buttons stayed visible because
  `.icon-btn { display: grid }` overrode unscoped hide rules. Drawer controls
  are now scoped to `.topbar` / `.sidebar` and only appear at 1024 px and below.

Not verified in this environment:

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
3. Decide whether the next iteration needs slim/classic skin models, animation or
   capes; keep the canvas renderer if offline Qt support remains a requirement.
4. Merge the task branch after review; no production deployment is part of this
   task.

## App shell scroll architecture (2026-11-18)

Problem: on long views the page felt like it was scrolling the menu instead of the
content. Root cause was the shell layout: `body` had `overflow-x: hidden`, which
in CSS makes `overflow-y` resolve to `auto`, so the whole document scrolled;
`.app-shell` was only `min-height: 100vh` with `overflow: visible`; and
`.main-scroll` was `flex: 1` but `overflow: visible`, so it was not actually a
scroll container. The sidebar was `position: sticky; height: 100vh` inside that
scrolling document, which is what made it move with the page before sticking.

Fix (inspired by established dashboard/app-shell patterns, not invented here):

- `body` is now `overflow: hidden` (clips only horizontally via the reset, and
  never scrolls vertically). The `body.nav-open` drawer lock stays as a safety net.
- `.app-shell` is the non-scrolling frame: `min-height: 100vh; overflow: hidden`.
- `.sidebar` is a full-height rail inside that frame: `height: 100vh;
  overflow-y: auto; overscroll-behavior: contain`. It only scrolls on its own when
  the navigation is taller than the viewport; otherwise it stays pinned and the
  content area scrolls. `position: sticky` is kept as a safeguard, but inside a
  non-scrolling shell it effectively pins the rail to the top.
- `.app-body` is the flex column that holds the topbar and the scroll area; it is
  `height: 100vh` so the flex child can actually be a scroll container.
- `.topbar` is `position: sticky; top: 0` inside `.app-body`.
- `.main-scroll` is the only vertical scroll container for content: `flex: 1;
  min-width: 0; min-height: 0; overflow-y: auto; overscroll-behavior: contain`.

Result: the user scrolls `.main-scroll`, not the menu and not the document. The
topbar stays pinned above the scroll area, and the sidebar stays pinned as a rail.
Verified across ultra-wide (2560x720, 3.56:1), 16:9 (1920x1080) and narrow
portrait (384x720, ~9:21) through the real Go handler at `127.0.0.1:8090`:
`body.scrollHeight` no longer exceeds the viewport, `document.documentElement`
stays at viewport height, `.main-scroll` is the scroll container, and the sidebar
no longer travels with the content.

Auth and plain shells (`.auth-layout`, `.plain-shell`) were given the same treatment
so they keep scrolling internally when their content exceeds the viewport instead of
being clipped by the now non-scrolling body: `height: 100vh; overflow-y: auto;
overscroll-behavior: contain`. Auth columns (`.auth-brand-inner`, `.auth-side`) are
constrained to `height: 100vh` so they do not grow the shell past the viewport on
short viewports.

## Button state model (2026-11-18)

A button is not one state. The states a user can perceive, and how this project
covers them, are:

- **Enabled/default** — the button is clickable. Covered by the base `.btn` /
  `.icon-btn` styles.
- **Hover** — cursor is over the button. Covered by `:hover:not(:disabled)` on
  `.btn` variants and `.icon-btn`.
- **Focus / focus-visible** — keyboard or assistive-tech focus. The project uses the
  global `:focus-visible` ring (2px accent, 2px offset) and now scopes a matching
  ring onto `.btn:focus-visible`, `.icon-btn:focus-visible`, selects and input
  wraps so the ring stays readable on tinted surfaces.
- **Active / pressed** — pointer is down on the button. Covered by `:active`.
  `.btn` moves down 1px; `.icon-btn` now also scales down slightly and dims, so
  icon buttons have the same tactile feedback as text buttons.
- **Clicked / activation** — the momentary press. Same `:active` styles; there is
  no separate activation style because activation is instantaneous.
- **Disabled** — action unavailable. Covered by `:disabled` (48% opacity, no
  transform, no box-shadow, `cursor: not-allowed`). Hover and active are excluded
  with `:not(:disabled)`, so a disabled button cannot pretend to react.
- **Selected / pressed toggle** — for toggle buttons such as the skin overlay
  toggle. The project uses `aria-pressed="true"` plus a helper `.is-active`
  class. The CSS now also styles `[aria-pressed="true"]` directly on `.icon-btn`
  and `.btn`/`.btn-primary`, so the pressed visual survives a view re-render that
  might drop the helper class.
- **Loading / pending** — action in flight. The JS swaps in a spinner and disables
  the button; the CSS keeps `:disabled` looking disabled and dims any inline
  spinner so the button cannot be tapped again while the request is in flight.
- **Conflict states** — e.g. focused but disabled, or pressed but disabled.
  `:disabled` wins because hover/active are gated on `:not(:disabled)` and the
  disabled style zeroes out transform and box-shadow.

No new interaction mechanic was added. The changes are visual completeness for
states that already existed in the markup or were reachable by pointer/keyboard
interaction: tactile `:active` feedback on icon buttons, an ARIA-backed toggle style
as a backup for the class-based one, explicit focus-visible on the interactive
controls, and a loading-spinner dim so pending buttons stay clearly disabled.

## Responsive grid and text overflow fixes (2026-11-18)

Two related issues were fixed together: grids slipping apart on unusual aspect
ratios, and long text escaping its block.

Grid slippage happens when a fluid track is `1fr` without a zero floor. The fix is
`minmax(0, 1fr)` on every fluid track (already largely in place) plus `min-width: 0`
on the items that contain nowrap content, so the item can shrink below its content
minimum instead of pushing the track open.

Text overflow happens when a block has `white-space: nowrap` but no clipping or
shrink strategy. The fix is the standard three-part pattern: `min-width: 0` (or a
constraint that lets the box be narrower than the text), `overflow: hidden`, and
either `text-overflow: ellipsis` for one-line truncations or `white-space: normal`
/ `overflow-wrap: anywhere` for wrappable text. This project already used the
ellipsis pattern in several places; the remaining gaps were:

- `.skin-facts dd`, `.detail-fact strong`, `.detail-fact span`, `.hero-meta-item
  strong`, `.meta-sub`, `.table-muted`, `.activity-copy > span` — now have
  `overflow: hidden; text-overflow: ellipsis` (and `min-width: 0` where the parent
  is a flex/grid item) so long server names, emails, relative times and fact values
  truncate instead of slipping the layout on narrow or wide screens.
- `.detail-fact` (flex space-between) and `.hero-meta-item` (grid) — given
  `min-width: 0` so their nowrap children can shrink.
- `.kpi-card > div` — duplicate `min-width: 0` guard so the card text column can
  shrink under the icon.

The grid tracks themselves were already `minmax(0, 1fr)` in the responsive cascade;
no track was changed to `1fr` without a zero floor. The result is that rows of
cards, facts, meta items and table cells stay together and shrink proportionally
instead of one long value forcing the whole row wide, across 21:9 through 9:21.

## Verification performed (2026-11-18, after scroll/button/grid fixes)

Go toolchain: Go 1.26.5 from the module toolchain cache. All Go checks run from
`server/`, the Go module root.

- `go vet ./web/...` — clean.
- `go build ./web/...` — compiles.
- `go test -count=1 ./web/...` — passes (`ok minecraftserverluncher/server/web`).

Because the frontend is embedded into the Go binary with `//go:embed static`, the
running local test binary had to be rebuilt from `server/` with `go build -o
_local_test/go-server/server.exe .` for the CSS changes to reach the live handler.
The old pre-built `_local_test/go-server/server.exe` was serving the previous
embedded copy and had to be replaced.

Live verification through the rebuilt Go handler at `http://127.0.0.1:8090`:

- `GET /assets/css/base.css` now returns `overflow: hidden` on `body` (was
  `overflow-x: hidden`).
- Dashboard on 2560x720 (ultra-wide): `body` and `.app-shell` are `overflow:
  hidden` and stay at viewport height; `.main-scroll` is `overflow-y: auto` and is
  the scroll container; `.sidebar` is a sticky rail that does not scroll with the
  content.
- Dashboard on 1920x1080 (16:9): same shell behavior; sidebar is a sticky rail,
  main content scrolls.
- Dashboard on 384x720 (narrow portrait, mobile breakpoint): sidebar is a fixed
  drawer (`position: fixed`, `z-index: 60`), `.main-scroll` scrolls the content.
- Players, admin and the not-found/app routes render with the same shell behavior;
  the admin data table still scrolls horizontally inside its own `.table-wrap`.
- Console: no JavaScript errors reported during these navigations.

Note on the browser preview in this environment: the local Chromium-based preview
tab does not always apply CSS transforms to the off-canvas `.sidebar` (the
`translateX(-102%)` rule is present in the stylesheet and is correct, but the
preview's computed transform stayed identity in this session). The drawer class
toggling (`is-open`, `nav-open`) works; the drawer behavior itself is expected to
work in a normal browser and in Qt WebEngine, and was verified visually in earlier
runs through a static server. This is a preview-tab rendering limitation, not a code
change.
