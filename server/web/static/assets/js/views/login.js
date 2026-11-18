import { api } from '../data/api.js';
import { icon } from '../core/icons.js';
import { toast } from '../core/toast.js';
import { authLayout, fieldError, mountAuthLayout, setButtonLoading } from '../components/auth-layout.js';

export const meta = {
  title: 'Bejelentkezés',
  layout: 'auth',
  guestOnly: true,
};

function safeNext(query) {
  const next = query.get('next') || '';
  if (!next.startsWith('/') || next.startsWith('//')) return '/dashboard';
  if (next === '/login' || next === '/register') return '/dashboard';
  return next;
}

export function render() {
  return authLayout(
    'login',
    `<div class="auth-head">
      <span class="eyebrow">Üdv újra</span>
      <h2>Bejelentkezés</h2>
      <p>Add meg a fiókod adatait a folytatáshoz.</p>
    </div>

    <form id="login-form" class="auth-form" novalidate>
      <div class="field">
        <label for="identifier">Felhasználónév vagy e-mail</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('user', 17)}</span>
          <input id="identifier" name="identifier" type="text" autocomplete="username" placeholder="pl. Aurora" required>
        </div>
        <p class="field-error" data-error-for="identifier" hidden></p>
      </div>

      <div class="field">
        <label for="password">Jelszó</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('key', 17)}</span>
          <input id="password" name="password" type="password" autocomplete="current-password" placeholder="••••••••" required>
          <button type="button" class="reveal-btn" data-action="toggle-password" data-target="password" aria-label="Jelszó megjelenítése">${icon('eye', 17)}</button>
        </div>
        <p class="field-error" data-error-for="password" hidden></p>
      </div>

      <div class="auth-meta-row">
        <span class="secure-note">${icon('lock', 14)} Titkosított kapcsolat</span>
        <span class="alpha-note">Alpha teszt</span>
      </div>

      <div class="form-alert" id="login-error" role="alert" hidden></div>

      <button type="submit" class="btn btn-primary btn-block btn-lg" id="login-submit">Belépés</button>
    </form>

    <div class="auth-divider"><span>vagy próbáld ki</span></div>

    <div class="demo-row">
      <button type="button" class="btn btn-soft" data-demo="player">${icon('play', 16)} Játékos demó</button>
      <button type="button" class="btn btn-soft" data-demo="admin">${icon('shield', 16)} Admin demó</button>
    </div>
    <p class="demo-hint">Demó jelszó: <code>alpha1234</code></p>

    <p class="auth-switch">Még nincs fiókod? <a href="/register" data-link>Regisztráció</a></p>`,
  );
}

export function mount(root, ctx) {
  const cleanupViewer = mountAuthLayout(root);
  const form = root.querySelector('#login-form');
  const alertBox = root.querySelector('#login-error');
  const submit = root.querySelector('#login-submit');

  const showAlert = (message) => {
    alertBox.textContent = message;
    alertBox.hidden = !message;
  };

  const finishLogin = (user) => {
    toast(`Üdv vissza, ${user.username}!`, { tone: 'success', title: 'Sikeres belépés' });
    ctx.navigate(safeNext(ctx.query));
  };

  const submitLogin = async (identifier, password) => {
    showAlert('');
    fieldError(root, 'identifier');
    fieldError(root, 'password');

    let hasError = false;
    if (!identifier.trim()) {
      fieldError(root, 'identifier', 'Add meg a felhasználóneved vagy az e-mail címed.');
      hasError = true;
    }
    if (!password) {
      fieldError(root, 'password', 'Add meg a jelszavad.');
      hasError = true;
    }
    if (hasError) return;

    setButtonLoading(submit, true, 'Belépés…');
    try {
      const user = await api.login(identifier, password);
      finishLogin(user);
    } catch (error) {
      showAlert(error.message || 'A bejelentkezés nem sikerült.');
      setButtonLoading(submit, false);
    }
  };

  form.addEventListener('submit', (event) => {
    event.preventDefault();
    const data = new FormData(form);
    submitLogin(String(data.get('identifier') || ''), String(data.get('password') || ''));
  });

  const onClick = async (event) => {
    const reveal = event.target.closest('[data-action="toggle-password"]');
    if (reveal) {
      const input = root.querySelector(`#${reveal.dataset.target}`);
      if (!input) return;
      const show = input.type === 'password';
      input.type = show ? 'text' : 'password';
      reveal.innerHTML = icon(show ? 'eye-off' : 'eye', 17);
      reveal.setAttribute('aria-label', show ? 'Jelszó elrejtése' : 'Jelszó megjelenítése');
      return;
    }

    const demo = event.target.closest('[data-demo]');
    if (demo) {
      showAlert('');
      demo.disabled = true;
      try {
        const user = await api.demoLogin(demo.dataset.demo);
        finishLogin(user);
      } catch (error) {
        showAlert(error.message || 'A demó belépés nem sikerült.');
      } finally {
        demo.disabled = false;
      }
    }
  };

  root.addEventListener('click', onClick);

  return () => {
    cleanupViewer();
    root.removeEventListener('click', onClick);
  };
}
