/**
 * Tiny history/hash router for dependency-free multi-view apps.
 *
 * - http(s): real URLs via history.pushState (deep links handled by the Go
 *   server, which serves the shell for the known routes).
 * - file:// (e.g. a local Qt WebEngine page without a server): hash fallback.
 */
export class Router {
  constructor({ routes, fallback, render }) {
    this.routes = routes;
    this.fallback = fallback;
    this.render = render;
    this.current = null;
    this.cleanup = null;
    this.started = false;
    this.renderToken = 0;
    this.usedHash = window.location.protocol === 'file:';
  }

  routeUrl(route) {
    if (!this.usedHash) return route.path;
    return `#${route.path}`;
  }

  readLocation() {
    if (!this.usedHash) {
      return { path: window.location.pathname || '/', search: window.location.search || '', hash: window.location.hash || '' };
    }
    const raw = window.location.hash.replace(/^#/, '') || '/';
    const [path, search = ''] = raw.split('?');
    return { path: path || '/', search: search ? `?${search}` : '', hash: '' };
  }

  match(path) {
    return this.routes.find((route) => route.path === path) || null;
  }

  start() {
    if (this.started) return;
    this.started = true;

    document.addEventListener('click', this.handleClick);
    window.addEventListener('popstate', this.handlePopState);
    window.addEventListener('hashchange', this.handlePopState);
    this.go();
  }

  handleClick = (event) => {
    if (event.defaultPrevented || event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    const anchor = event.target instanceof Element ? event.target.closest('a[data-link]') : null;
    if (!anchor) return;
    const href = anchor.getAttribute('href');
    if (!href || href.startsWith('http') || href.startsWith('mailto:') || href.startsWith('#')) return;
    if (anchor.target && anchor.target !== '_self') return;
    event.preventDefault();
    this.navigate(href);
  };

  handlePopState = () => {
    this.go();
  };

  parseUrl(url) {
    const base = this.usedHash ? `${window.location.origin}/` : window.location.origin;
    try {
      return new URL(url, base);
    } catch (error) {
      return new URL('/', base);
    }
  }

  navigate(url, { replace = false } = {}) {
    const target = this.parseUrl(url);
    if (this.usedHash) {
      const raw = `${target.pathname}${target.search}`;
      const nextHash = `#${raw}`;
      if (replace) {
        window.history.replaceState({}, '', nextHash);
        this.go();
      } else if (window.location.hash !== nextHash) {
        window.location.hash = raw;
      } else {
        this.go();
      }
      return;
    }

    const next = `${target.pathname}${target.search}`;
    const current = `${window.location.pathname}${window.location.search}`;
    if (next === current) {
      this.go();
      return;
    }
    window.history[replace ? 'replaceState' : 'pushState']({}, '', next);
    this.go();
  }

  async go() {
    const location = this.readLocation();
    const route = this.match(location.path) || this.fallback;
    const token = ++this.renderToken;

    document.body.dataset.route = (route.path || '/').replace(/\//g, '') || 'root';

    if (typeof this.cleanup === 'function') {
      try {
        this.cleanup();
      } catch (error) {
        console.error('route cleanup failed', error);
      }
      this.cleanup = null;
    }

    const query = new URLSearchParams(location.search.replace(/^\?/, ''));
    try {
      const result = await this.render(route, { path: location.path, query, token, isCurrent: () => token === this.renderToken });
      if (!result) return;
      if (typeof result === 'function') this.cleanup = result;
      this.current = route;
      window.scrollTo({ top: 0, behavior: 'auto' });
    } catch (error) {
      console.error('route render failed', error);
      if (token === this.renderToken) {
        document.getElementById('app').innerHTML = `<div class="fatal-error"><h1>Váratlan hiba</h1><p>${String(error && error.message ? error.message : error)}</p><a class="btn btn-primary" href="/dashboard" data-link>Vissza az irányítópultra</a></div>`;
      }
    }
  }
}

export default Router;
