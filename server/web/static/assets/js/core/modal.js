import { esc } from './dom.js';
import { icon } from './icons.js';

let activeModal = null;

/**
 * Promise-based confirmation dialog used for destructive admin actions.
 * @returns {Promise<boolean>}
 */
export function confirmDialog({
  title = 'Megerősítés',
  message = '',
  confirmLabel = 'Megerősítés',
  cancelLabel = 'Mégse',
  tone = 'danger',
  iconName = '',
} = {}) {
  if (activeModal) activeModal.close(false);

  return new Promise((resolve) => {
    const overlay = document.createElement('div');
    overlay.className = 'modal-backdrop';
    const toneIcon = iconName || (tone === 'danger' ? 'alert' : 'info');
    overlay.innerHTML = `
      <div class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title" aria-describedby="modal-message">
        <div class="modal-icon tone-${esc(tone)}">${icon(toneIcon, 22)}</div>
        <h2 id="modal-title">${esc(title)}</h2>
        <p id="modal-message">${esc(message)}</p>
        <div class="modal-actions">
          <button type="button" class="btn btn-ghost" data-modal-cancel>${esc(cancelLabel)}</button>
          <button type="button" class="btn ${tone === 'danger' ? 'btn-danger' : 'btn-primary'}" data-modal-confirm>${esc(confirmLabel)}</button>
        </div>
      </div>
    `;

    const previousFocus = document.activeElement;
    const confirmButton = overlay.querySelector('[data-modal-confirm]');
    const cancelButton = overlay.querySelector('[data-modal-cancel]');

    const close = (result) => {
      if (!overlay.isConnected) return;
      document.removeEventListener('keydown', onKeyDown);
      overlay.classList.remove('is-visible');
      window.setTimeout(() => overlay.remove(), 160);
      activeModal = null;
      if (previousFocus && typeof previousFocus.focus === 'function') previousFocus.focus();
      resolve(result);
    };

    const onKeyDown = (event) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        close(false);
        return;
      }
      if (event.key !== 'Tab') return;
      const focusables = [cancelButton, confirmButton];
      const index = focusables.indexOf(document.activeElement);
      event.preventDefault();
      const next = event.shiftKey ? (index - 1 + focusables.length) % focusables.length : (index + 1) % focusables.length;
      focusables[next].focus();
    };

    cancelButton.addEventListener('click', () => close(false));
    confirmButton.addEventListener('click', () => close(true));
    overlay.addEventListener('click', (event) => {
      if (event.target === overlay) close(false);
    });
    document.addEventListener('keydown', onKeyDown);

    document.body.appendChild(overlay);
    activeModal = { close };
    confirmButton.focus();
    window.requestAnimationFrame(() => overlay.classList.add('is-visible'));
  });
}
