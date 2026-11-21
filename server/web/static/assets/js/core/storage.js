import { config } from '../config.js';

const prefix = config.storagePrefix || '';

function namespaced(key) {
  return `${prefix}${key}`;
}

/**
 * Qt WebEngine profiles can have storage disabled or full. Every access is
 * guarded so a locked-down profile degrades to in-memory defaults instead of
 * breaking navigation.
 */
export const storage = {
  get(key, fallback = null) {
    try {
      const value = window.localStorage.getItem(namespaced(key));
      return value === null ? fallback : value;
    } catch (error) {
      return fallback;
    }
  },

  set(key, value) {
    try {
      window.localStorage.setItem(namespaced(key), String(value));
      return true;
    } catch (error) {
      return false;
    }
  },

  remove(key) {
    try {
      window.localStorage.removeItem(namespaced(key));
    } catch (error) {
      /* storage unavailable */
    }
  },

  getJSON(key, fallback = null) {
    const raw = this.get(key);
    if (!raw) return fallback;
    try {
      return JSON.parse(raw);
    } catch (error) {
      this.remove(key);
      return fallback;
    }
  },

  setJSON(key, value) {
    try {
      return this.set(key, JSON.stringify(value));
    } catch (error) {
      return false;
    }
  },
};

export default storage;
