/**
 * Dolphin Whitelabel Studio Component
 * Interactive in-app brand customizer, proximity mesh config editor & GCP Redis Memorystore switcher
 */

import { PRESET_THEMES } from '../config/whitelabel.js';

export function renderWhitelabelStudioModal(config) {
  return `
    <div class="modal-overlay" id="whitelabelStudioModalOverlay">
      <div class="modal-sheet">
        <div class="sheet-handle"></div>

        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
          <div>
            <h2 style="font-family: var(--font-display); font-size: 1.15rem; font-weight: 800; display: flex; align-items: center; gap: 8px;">
              <i class="fa-solid fa-wand-magic-sparkles" style="color: var(--brand-primary);"></i> Whitelabel Studio
            </h2>
            <p style="font-size: 0.75rem; color: var(--text-muted);">Customize your Dolphin Android App live</p>
          </div>

          <button id="btnCloseStudioModal" style="background: transparent; border: none; color: var(--text-muted); font-size: 1.2rem; cursor: pointer;">
            <i class="fa-solid fa-xmark"></i>
          </button>
        </div>

        <!-- Section 1: App Identity -->
        <div class="form-group">
          <label class="form-label">Application Name</label>
          <input type="text" id="studioAppNameInput" class="form-input" value="${config.appName}" />
        </div>

        <!-- Section 2: Connection & Proximity Mode -->
        <div class="form-group">
          <label class="form-label">Backend Connection Mode</label>
          <select id="studioConnectionModeSelect" class="form-select">
            <option value="GCP_REDIS" ${config.connectionMode === 'GCP_REDIS' ? 'selected' : ''}>Google Cloud Redis Memorystore (pqr-info-5d-mesh)</option>
            <option value="INDEPENDENT" ${config.connectionMode === 'INDEPENDENT' ? 'selected' : ''}>Autonomous Independent Mode (Offline Mesh / Serverless)</option>
            <option value="MOCK" ${config.connectionMode === 'MOCK' ? 'selected' : ''}>Standalone Demo Mode (Mock Engine)</option>
            <option value="LIVE_UNA" ${config.connectionMode === 'LIVE_UNA' ? 'selected' : ''}>Live UNA / Dolphin REST API Server</option>
          </select>
        </div>

        <div class="form-group" id="studioServerUrlGroup">
          <label class="form-label">GCP Project & Endpoint</label>
          <input type="text" id="studioServerUrlInput" class="form-input" value="Project: pqr-info-5d-mesh (Host: ${config.gcpRedisHost || '10.140.0.8:6379'})" readonly />
        </div>

        <!-- Section 3: Theme Presets -->
        <div class="form-group">
          <label class="form-label">Brand Color Preset</label>
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px;">
            ${Object.keys(PRESET_THEMES).map(key => {
              const preset = PRESET_THEMES[key];
              const isSelected = config.presetKey === key;
              return `
                <div class="preset-card ${isSelected ? 'selected' : ''}" data-preset="${key}" style="border: 1.5px solid ${isSelected ? 'var(--brand-primary)' : 'var(--border-color)'}; background: rgba(255,255,255,0.04); padding: 10px; border-radius: 12px; cursor: pointer; display: flex; align-items: center; gap: 8px;">
                  <div style="width: 20px; height: 20px; border-radius: 50%; background: linear-gradient(135deg, ${preset.primary}, ${preset.secondary}); flex-shrink: 0;"></div>
                  <span style="font-size: 0.78rem; font-weight: 600;">${preset.name.split(' ')[0]}</span>
                </div>
              `;
            }).join('')}
          </div>
        </div>

        <!-- Section 4: Color Hex Pickers -->
        <div class="form-group">
          <label class="form-label">Custom Brand Colors</label>
          <div style="display: flex; gap: 10px;">
            <div style="flex: 1;">
              <span style="font-size: 0.7rem; color: var(--text-muted);">Primary</span>
              <input type="color" id="studioPrimaryColorPicker" value="${config.primaryColor}" style="width: 100%; height: 36px; border: none; border-radius: 8px; cursor: pointer; background: transparent;" />
            </div>
            <div style="flex: 1;">
              <span style="font-size: 0.7rem; color: var(--text-muted);">Secondary</span>
              <input type="color" id="studioSecondaryColorPicker" value="${config.secondaryColor}" style="width: 100%; height: 36px; border: none; border-radius: 8px; cursor: pointer; background: transparent;" />
            </div>
          </div>
        </div>

        <!-- Section 5: Dark / Light Mode -->
        <div class="form-group">
          <label class="form-label">Appearance Theme</label>
          <div style="display: flex; gap: 10px;">
            <button class="btn-primary studio-theme-btn" data-mode="dark" style="flex: 1; ${config.themeMode === 'dark' ? '' : 'background: rgba(255,255,255,0.1); box-shadow: none;'}">
              <i class="fa-solid fa-moon"></i> Dark Mode
            </button>
            <button class="btn-primary studio-theme-btn" data-mode="light" style="flex: 1; ${config.themeMode === 'light' ? '' : 'background: rgba(255,255,255,0.1); box-shadow: none;'}">
              <i class="fa-solid fa-sun"></i> Light Mode
            </button>
          </div>
        </div>

        <!-- Action Buttons -->
        <div style="display: flex; gap: 10px; margin-top: 20px;">
          <button id="btnSaveWhitelabelConfig" class="btn-primary" style="flex: 2;">
            <i class="fa-solid fa-check" style="margin-right: 6px;"></i> Apply Changes
          </button>
          <button id="btnExportConfigJson" class="btn-primary" style="flex: 1; background: rgba(255,255,255,0.1); box-shadow: none;">
            <i class="fa-solid fa-download"></i> Export JSON
          </button>
        </div>

      </div>
    </div>
  `;
}
