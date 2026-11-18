import { api } from '../data/api.js';
import { icon } from '../core/icons.js';
import { toast } from '../core/toast.js';
import { authLayout, fieldError, mountAuthLayout, setButtonLoading } from '../components/auth-layout.js';

export const meta = {
  title: 'Regisztráció',
  layout: 'auth',
  guestOnly: true,
};

const STRENGTH_LABELS = ['Túl rövid', 'Gyenge', 'Közepes', 'Erős'];

function strengthOf(password) {
  if (!password || password.length < 8) return 0;
  let score = 1;
  if (password.length >= 12) score += 1;
  if (/[A-Z]/.test(password) && /[a-z]/.test(password)) score += 1;
  if (/\d/.test(password) || /[^A-Za-z0-9]/.test(password)) score += 1;
  return Math.min(3, score);
}

export function render() {
  return authLayout(
    'register',
    `<div class="auth-head">
      <span class="eyebrow">Új fiók</span>
      <h2>Regisztráció</h2>
      <p>Hozd létre a saját fiókod az Alpha Networkön.</p>
    </div>

    <form id="register-form" class="auth-form" novalidate>
      <div class="field">
        <label for="username">Felhasználónév</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('user', 17)}</span>
          <input id="username" name="username" type="text" autocomplete="username" placeholder="3–16 karakter" required>
        </div>
        <p class="field-hint">Csak betű, szám és aláhúzás.</p>
        <p class="field-error" data-error-for="username" hidden></p>
      </div>

      <div class="field">
        <label for="email">E-mail cím</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('mail', 17)}</span>
          <input id="email" name="email" type="email" autocomplete="email" placeholder="te@example.com" required>
        </div>
        <p class="field-error" data-error-for="email" hidden></p>
      </div>

      <div class="field">
        <label for="password">Jelszó</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('key', 17)}</span>
          <input id="password" name="password" type="password" autocomplete="new-password" placeholder="Legalább 8 karakter" required>
          <button type="button" class="reveal-btn" data-action="toggle-password" data-target="password" aria-label="Jelszó megjelenítése">${icon('eye', 17)}</button>
        </div>
        <div class="password-meter" aria-hidden="true">
          <span></span><span></span><span></span><span></span>
          <em>Erősség: <b data-strength-label>Túl rövid</b></em>
        </div>
        <p class="field-error" data-error-for="password" hidden></p>
      </div>

      <div class="field">
        <label for="confirm">Jelszó megerősítése</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('lock', 17)}</span>
          <input id="confirm" name="confirm" type="password" autocomplete="new-password" placeholder="Írd be újra" required>
        </div>
        <p class="field-error" data-error-for="confirm" hidden></p>
      </div>

      <label class="checkbox-field">
        <input type="checkbox" id="alpha-terms" name="terms">
        <span>Tudomásul veszem, hogy alpha tesztkörnyezetben vagyok, és az adatok a teszt során változhatnak.</span>
      </label>
      <p class="field-error" data-error-for="terms" hidden></p>

      <div class="form-alert" id="register-error" role="alert" hidden></div>

      <button type="submit" class="btn btn-primary btn-block btn-lg" id="register-submit">Fiók létrehozása</button>
    </form>

    <p class="auth-switch">Már van fiókod? <a href="/login" data-link>Bejelentkezés</a></p>`,
  );
}

export function mount(root, ctx) {
  const cleanupViewer = mountAuthLayout(root);
  const form = root.querySelector('#register-form');
  const alertBox = root.querySelector('#register-error');
  const submit = root.querySelector('#register-submit');
  const password = root.querySelector('#password');
  const meter = root.querySelector('.password-meter');
  const strengthLabel = root.querySelector('[data-strength-label]');

  const showAlert = (message) => {
    alertBox.textContent = message;
    alertBox.hidden = !message;
  };

  const updateStrength = () => {
    const score = strengthOf(password.value);
    meter.dataset.strength = String(score);
    meter.querySelectorAll('span').forEach((bar, index) => {
      bar.classList.toggle('is-filled', index <= score && Boolean(password.value));
    });
    strengthLabel.textContent = STRENGTH_LABELS[score] || STRENGTH_LABELS[0];
  };

  password.addEventListener('input', updateStrength);

  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    showAlert('');
    ['username', 'email', 'password', 'confirm', 'terms'].forEach((field) => fieldError(root, field));

    const data = new FormData(form);
    const username = String(data.get('username') || '').trim();
    const email = String(data.get('email') || '').trim();
    const pass = String(data.get('password') || '');
    const confirm = String(data.get('confirm') || '');
    const terms = data.get('terms') === 'on';
    let hasError = false;

    if (!/^[A-Za-z0-9_]{3,16}$/.test(username)) {
      fieldError(root, 'username', '3–16 karakter, csak betű, szám és _ használható.');
      hasError = true;
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      fieldError(root, 'email', 'Érvényes e-mail címet adj meg.');
      hasError = true;
    }
    if (pass.length < 8) {
      fieldError(root, 'password', 'A jelszó legalább 8 karakter legyen.');
      hasError = true;
    }
    if (confirm !== pass) {
      fieldError(root, 'confirm', 'A két jelszó nem egyezik.');
      hasError = true;
    }
    if (!terms) {
      fieldError(root, 'terms', 'Az alpha teszt feltételeinek elfogadása szükséges.');
      hasError = true;
    }
    if (hasError) return;

    setButtonLoading(submit, true, 'Fiók létrehozása…');
    try {
      const user = await api.register({ username, email, password: pass });
      toast('A fiókod elkészült. Üdv az Alphán!', { tone: 'success', title: 'Sikeres regisztráció' });
      ctx.navigate('/dashboard');
      return user;
    } catch (error) {
      showAlert(error.message || 'A regisztráció nem sikerült.');
      setButtonLoading(submit, false);
    }
  });

  const onClick = (event) => {
    const reveal = event.target.closest('[data-action="toggle-password"]');
    if (!reveal) return;
    const input = root.querySelector(`#${reveal.dataset.target}`);
    if (!input) return;
    const show = input.type === 'password';
    input.type = show ? 'text' : 'password';
    reveal.innerHTML = icon(show ? 'eye-off' : 'eye', 17);
    reveal.setAttribute('aria-label', show ? 'Jelszó elrejtése' : 'Jelszó megjelenítése');
  };

  root.addEventListener('click', onClick);
  updateStrength();

  return () => {
    cleanupViewer();
    root.removeEventListener('click', onClick);
    password.removeEventListener('input', updateStrength);
  };
}
