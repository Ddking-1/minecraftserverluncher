import { api, passwordStrength, passwordStrengthScore } from '../data/api.js';
import { icon } from '../core/icons.js';
import { toast } from '../core/toast.js';
import { authLayout, fieldError, mountAuthLayout, setButtonLoading } from '../components/auth-layout.js';

export const meta = {
  title: 'Jelszó módosítása',
  layout: 'auth',
};

const CRITERIA = [
  { key: 'minLength8', label: 'Legalább 8 karakter' },
  { key: 'hasNumber', label: 'Tartalmaz számot' },
  { key: 'hasUpper', label: 'Tartalmaz nagybetűt' },
  { key: 'minLength10', label: 'Legalább 10 karakter' },
  { key: 'hasSpecial', label: 'Tartalmaz speciális karaktert' },
];

export function render() {
  return authLayout(
    'change-password',
    `<div class="auth-head">
      <span class="eyebrow">Biztonság</span>
      <h2>Jelszó módosítása</h2>
      <p>Állíts be egy új, erős jelszót a fiókodhoz.</p>
    </div>

    <form id="change-password-form" class="auth-form" novalidate>
      <div class="field">
        <label for="old-password">Jelenlegi jelszó</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('key', 17)}</span>
          <input id="old-password" name="oldPassword" type="password" autocomplete="current-password" placeholder="••••••••" required>
          <button type="button" class="reveal-btn" data-action="toggle-password" data-target="old-password" aria-label="Jelszó megjelenítése">${icon('eye', 17)}</button>
        </div>
        <p class="field-error" data-error-for="oldPassword" hidden></p>
      </div>

      <div class="field">
        <label for="new-password">Új jelszó</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('key', 17)}</span>
          <input id="new-password" name="newPassword" type="password" autocomplete="new-password" placeholder="Legalább 8 karakter" required>
          <button type="button" class="reveal-btn" data-action="toggle-password" data-target="new-password" aria-label="Jelszó megjelenítése">${icon('eye', 17)}</button>
        </div>
        <ul class="password-criteria" id="password-criteria" aria-live="polite">
          ${CRITERIA.map((c) => `<li data-criterion="${c.key}">${icon('x', 13)} ${c.label}</li>`).join('')}
        </ul>
        <div class="password-meter" aria-hidden="true">
          <span></span><span></span><span></span><span></span><span></span>
          <em>Erősség: <b data-strength-label>Túl rövid</b></em>
        </div>
        <p class="field-error" data-error-for="newPassword" hidden></p>
      </div>

      <div class="field">
        <label for="confirm-password">Új jelszó megerősítése</label>
        <div class="input-wrap">
          <span class="field-icon">${icon('lock', 17)}</span>
          <input id="confirm-password" name="confirmPassword" type="password" autocomplete="new-password" placeholder="Írd be újra" required>
        </div>
        <p class="field-error" data-error-for="confirmPassword" hidden></p>
      </div>

      <div class="form-alert" id="change-password-error" role="alert" hidden></div>

      <button type="submit" class="btn btn-primary btn-block btn-lg" id="change-password-submit">Jelszó mentése</button>
    </form>

    <p class="auth-switch"><a href="/dashboard" data-link>Vissza az irányítópultra</a></p>`,
  );
}

export function mount(root, ctx) {
  const cleanupViewer = mountAuthLayout(root);
  const form = root.querySelector('#change-password-form');
  const alertBox = root.querySelector('#change-password-error');
  const submit = root.querySelector('#change-password-submit');
  const newPassword = root.querySelector('#new-password');
  const meter = root.querySelector('.password-meter');
  const strengthLabel = root.querySelector('[data-strength-label]');

  const showAlert = (message) => {
    alertBox.textContent = message;
    alertBox.hidden = !message;
  };

  const updateCriteria = () => {
    const s = passwordStrength(newPassword.value);
    CRITERIA.forEach((c) => {
      const li = root.querySelector(`[data-criterion="${c.key}"]`);
      if (!li) return;
      const ok = s[c.key];
      li.classList.toggle('is-ok', ok);
      li.innerHTML = `${icon(ok ? 'check' : 'x', 13)} ${c.label}`;
    });
    const score = passwordStrengthScore(newPassword.value);
    meter.dataset.strength = String(score);
    meter.querySelectorAll('span').forEach((bar, index) => {
      bar.classList.toggle('is-filled', index < score && Boolean(newPassword.value));
    });
    const labels = ['Túl rövid', 'Gyenge', 'Közepes', 'Jó', 'Erős', 'Kiváló'];
    strengthLabel.textContent = labels[score] || labels[0];
  };

  newPassword.addEventListener('input', updateCriteria);

  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    showAlert('');
    ['oldPassword', 'newPassword', 'confirmPassword'].forEach((f) => fieldError(root, f));

    const data = new FormData(form);
    const oldPass = String(data.get('oldPassword') || '');
    const newPass = String(data.get('newPassword') || '');
    const confirm = String(data.get('confirmPassword') || '');
    let hasError = false;

    if (!oldPass) { fieldError(root, 'oldPassword', 'Add meg a jelenlegi jelszót.'); hasError = true; }
    const s = passwordStrength(newPass);
    if (!Object.values(s).every(Boolean)) {
      fieldError(root, 'newPassword', 'Az új jelszó nem felel meg a biztonsági követelményeknek.');
      hasError = true;
    }
    if (confirm !== newPass) { fieldError(root, 'confirmPassword', 'A két jelszó nem egyezik.'); hasError = true; }
    if (hasError) return;

    setButtonLoading(submit, true, 'Mentés…');
    try {
      await api.changePassword(oldPass, newPass);
      toast('A jelszó sikeresen módosítva.', { tone: 'success', title: 'Jelszó frissítve' });
      ctx.navigate('/dashboard');
    } catch (error) {
      showAlert(error.message || 'A jelszó módosítása nem sikerült.');
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
  updateCriteria();

  return () => {
    cleanupViewer();
    root.removeEventListener('click', onClick);
    newPassword.removeEventListener('input', updateCriteria);
  };
}