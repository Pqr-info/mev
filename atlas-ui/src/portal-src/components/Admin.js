/**
 * 5D Celestial Mesh Admin Control Interface
 * Real Sovereign-27 Backend Connection (http://localhost:4050) - Zero Mocks
 */

import { locationEngine } from '../services/location.js';
import { celestialEngine, CELESTIAL_TYPES } from '../services/celestial.js';

export function renderAdminDashboard(config) {
  const celestials = celestialEngine.getCelestialObjects(10000);

  return `
    <div class="admin-section">
      <!-- Admin Header Banner -->
      <div class="glass-card" style="margin-bottom: 14px; background: linear-gradient(135deg, rgba(14,165,233,0.15) 0%, rgba(99,102,241,0.15) 100%); border-color: var(--border-highlight);">
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div>
            <h2 style="font-family: var(--font-display); font-size: 1.15rem; font-weight: 800; display: flex; align-items: center; gap: 8px;">
              <i class="fa-solid fa-shield-halved" style="color: var(--brand-primary);"></i> Sovereign-27 Real Backend
            </h2>
            <span style="font-size: 0.75rem; color: var(--text-muted);">API Server: <strong>http://localhost:4050</strong></span>
          </div>

          <span id="backendStatusChip" class="location-chip" style="font-size: 0.7rem; background: #10b98122; color: #10b981; border-color: #10b98144;">
            <i class="fa-solid fa-server"></i> Connecting to Backend...
          </span>
        </div>

        <div style="font-size: 0.72rem; color: var(--text-muted); margin-top: 10px; background: rgba(0,0,0,0.25); padding: 8px 12px; border-radius: 8px; font-family: monospace;">
          Disk DB Path: <span id="diskDbPathDisplay" style="color: #10b981;">data/pqlite_gmi_mesh.db</span> (SQLite WAL)
        </div>
      </div>

      <!-- Sovereign-27 Real Stack Pipeline Execution -->
      <div class="glass-card" style="margin-bottom: 14px; border-color: rgba(16,185,129,0.4);">
        <div class="section-header" style="margin-bottom: 10px;">
          <span class="section-title" style="font-size: 0.95rem;">
            <i class="fa-solid fa-play" style="color: #10b981;"></i> Sovereign-27 Real Pipeline Execution
          </span>

          <button id="btnRunRealSovereign27" class="btn-primary" style="width: auto; padding: 4px 12px; font-size: 0.72rem; background: linear-gradient(135deg, #10b981 0%, #06b6d4 100%);">
            <i class="fa-solid fa-code" style="margin-right: 4px;"></i> Execute Real API Calls
          </button>
        </div>

        <div style="font-size: 0.75rem; color: var(--text-muted); margin-bottom: 10px;">
          Calls real Express/SQLite backend endpoints (/api/gmi/...) and writes persistent records on disk.
        </div>

        <div id="realPipelineLog" style="display: flex; flex-direction: column; gap: 6px; font-family: monospace; font-size: 0.72rem; background: rgba(0,0,0,0.4); padding: 10px; border-radius: 8px; max-height: 180px; overflow-y: auto;">
          <span style="color: var(--text-muted);">Click "Execute Real API Calls" to run live server operations.</span>
        </div>
      </div>

      <!-- Real Memory Search (http://localhost:4050/api/gmi/searchMemory) -->
      <div class="glass-card" style="margin-bottom: 14px;">
        <div class="section-header" style="margin-bottom: 10px;">
          <span class="section-title" style="font-size: 0.95rem;">
            <i class="fa-solid fa-magnifying-glass" style="color: var(--brand-primary);"></i> Real Disk Memory Search
          </span>
        </div>

        <div class="form-group" style="margin-bottom: 8px;">
          <div style="display: flex; gap: 8px;">
            <input type="text" id="realSearchInput" class="form-input" placeholder="Search disk memory..." value="Sovereign-27" style="font-size: 0.8rem;" />
            <button id="btnRealSearchMemory" class="btn-primary" style="width: auto; padding: 8px 14px; font-size: 0.78rem;">
              Search Server
            </button>
          </div>
        </div>

        <div id="realSearchResult" style="display: flex; flex-direction: column; gap: 6px; font-family: monospace; font-size: 0.72rem;"></div>
      </div>

      <!-- Real PQLite SQL Console -->
      <div class="glass-card" style="margin-bottom: 14px; border-color: rgba(6,182,212,0.4);">
        <div class="section-header" style="margin-bottom: 10px;">
          <span class="section-title" style="font-size: 0.95rem;">
            <i class="fa-solid fa-database" style="color: #06b6d4;"></i> Real PQLite SQL Query Engine
          </span>
        </div>

        <div class="form-group" style="margin-bottom: 10px;">
          <div style="display: flex; gap: 8px;">
            <input type="text" id="realSqlInput" class="form-input" value="SELECT * FROM memory_page" style="font-family: monospace; font-size: 0.8rem;" />
            <button id="btnRunRealSql" class="btn-primary" style="width: auto; padding: 8px 14px; font-size: 0.78rem; background: linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%);">
              Execute SQL
            </button>
          </div>
        </div>

        <div id="realSqlResult" style="display: flex; flex-direction: column; gap: 6px; font-family: monospace; font-size: 0.72rem;"></div>
      </div>

      <!-- Celestial Anomaly Spawner Control -->
      <div class="glass-card">
        <div class="section-header" style="margin-bottom: 10px;">
          <span class="section-title" style="font-size: 0.95rem;">
            <i class="fa-solid fa-wand-magic-sparkles" style="color: #fb923c;"></i> Spawn Celestial Anomaly
          </span>
        </div>

        <div class="form-group">
          <select id="adminSpawnTypeSelect" class="form-select">
            ${Object.keys(CELESTIAL_TYPES).map(key => `
              <option value="${key}">${CELESTIAL_TYPES[key].name}</option>
            `).join('')}
          </select>
        </div>

        <button id="btnAdminTriggerSpawn" class="btn-primary" style="background: linear-gradient(135deg, #fb923c 0%, #f43f5e 100%);">
          <i class="fa-solid fa-bolt" style="margin-right: 6px;"></i> Broadcast & Spawn Celestial Object
        </button>
      </div>
    </div>
  `;
}
