/**
 * Dolphin Whitelabel Configuration Engine
 * Manages dynamic app branding, color tokens, location proximity, and GCP Redis Memorystore.
 */

export const PRESET_THEMES = {
  ocean: {
    name: 'Dolphin Ocean (Default)',
    primary: '#0ea5e9',
    secondary: '#6366f1',
    accent: '#06b6d4',
    bgDark: '#090d16',
    cardBg: 'rgba(18, 25, 41, 0.75)'
  },
  neon: {
    name: 'Cyber Neon',
    primary: '#10b981',
    secondary: '#06b6d4',
    accent: '#3b82f6',
    bgDark: '#051114',
    cardBg: 'rgba(15, 32, 39, 0.75)'
  },
  royal: {
    name: 'Royal Violet',
    primary: '#8b5cf6',
    secondary: '#ec4899',
    accent: '#a855f7',
    bgDark: '#0d0714',
    cardBg: 'rgba(28, 18, 41, 0.75)'
  },
  sunset: {
    name: 'Crimson Sunset',
    primary: '#f43f5e',
    secondary: '#fb923c',
    accent: '#f43f5e',
    bgDark: '#12070a',
    cardBg: 'rgba(38, 16, 23, 0.75)'
  }
};

const DEFAULT_CONFIG = {
  appName: 'Dolphin 5D Mesh',
  appTagline: 'Google Cloud Redis Memorystore Mesh Platform',
  gcpProjectId: 'pqr-info-5d-mesh',
  gcpRedisHost: '10.140.0.8:6379',
  serverUrl: 'https://community.dolphin.app/api.php',
  connectionMode: 'GCP_REDIS',
  themeMode: 'dark',
  presetKey: 'ocean',
  primaryColor: '#0ea5e9',
  secondaryColor: '#6366f1',
  accentColor: '#06b6d4',
  proximityRadiusMeters: 5000,
  modules: {
    radar: true,
    stories: true,
    chat: true,
    groups: true,
    notifications: true,
    whitelabelStudio: true
  },
  auth: {
    isLoggedIn: true,
    userToken: 'jwt_mock_token_98432',
    user: {
      id: 'u101',
      name: 'Antigravity Dev',
      username: 'ag_dev',
      avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=250&q=80',
      badge: '5D Mesh Master',
      karma: 2450
    }
  }
};

class WhitelabelEngine {
  constructor() {
    this.config = this.loadSavedConfig();
    this.listeners = [];
  }

  loadSavedConfig() {
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('dolphin_whitelabel_config');
      if (saved) {
        try {
          return { ...DEFAULT_CONFIG, ...JSON.parse(saved) };
        } catch (e) {
          console.error('Failed to parse saved whitelabel config', e);
        }
      }
    }
    return { ...DEFAULT_CONFIG };
  }

  saveConfig(newConfig) {
    this.config = { ...this.config, ...newConfig };
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('dolphin_whitelabel_config', JSON.stringify(this.config));
    }
    this.applyTheme();
    this.notifyListeners();
  }

  applyTheme() {
    if (typeof document === 'undefined') return;
    const root = document.documentElement;
    const { primaryColor, secondaryColor, accentColor, themeMode } = this.config;

    const hexToRgb = (hex) => {
      const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
      return result ? `${parseInt(result[1], 16)}, ${parseInt(result[2], 16)}, ${parseInt(result[3], 16)}` : '14, 165, 233';
    };

    root.style.setProperty('--brand-primary', primaryColor);
    root.style.setProperty('--brand-primary-rgb', hexToRgb(primaryColor));
    root.style.setProperty('--brand-secondary', secondaryColor);
    root.style.setProperty('--brand-accent', accentColor);
    root.style.setProperty('--brand-gradient', `linear-gradient(135deg, ${primaryColor} 0%, ${secondaryColor} 100%)`);
    root.style.setProperty('--brand-glow', `0 8px 32px rgba(${hexToRgb(primaryColor)}, 0.35)`);

    root.setAttribute('data-theme', themeMode);
  }

  subscribe(callback) {
    this.listeners.push(callback);
    return () => {
      this.listeners = this.listeners.filter(cb => cb !== callback);
    };
  }

  notifyListeners() {
    this.listeners.forEach(cb => cb(this.config));
  }

  resetToDefault() {
    this.saveConfig(DEFAULT_CONFIG);
  }
}

export const whitelabelEngine = new WhitelabelEngine();
