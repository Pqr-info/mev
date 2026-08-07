/**
 * Dolphin User Profile Component with 5D Mesh Quantum Badges
 */

import { celestialEngine } from '../services/celestial.js';

export function renderProfile(user, config) {
  return `
    <div class="profile-section">
      <!-- Cover & Header -->
      <div class="glass-card" style="padding: 0; overflow: hidden; position: relative;">
        <div style="height: 120px; background: var(--brand-gradient); position: relative;">
          <img src="https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=800&q=80" style="width: 100%; height: 100%; object-fit: cover; opacity: 0.4;" />
          <button id="btnProfileEdit" class="icon-btn" style="position: absolute; top: 10px; right: 10px; background: rgba(0,0,0,0.5);">
            <i class="fa-solid fa-pen-to-square"></i>
          </button>
        </div>

        <div style="padding: 0 16px 16px 16px; margin-top: -36px; position: relative;">
          <div style="display: flex; justify-content: space-between; align-items: flex-end;">
            <img src="${user.avatar}" style="width: 76px; height: 76px; border-radius: 50%; border: 4px solid var(--bg-dark); object-fit: cover; box-shadow: var(--brand-glow);" />
            
            <div style="display: flex; gap: 8px;">
              <button class="btn-primary" id="btnOpenStudioFromProfile" style="width: auto; padding: 6px 14px; font-size: 0.8rem;">
                <i class="fa-solid fa-palette" style="margin-right: 4px;"></i> Whitelabel Studio
              </button>
            </div>
          </div>

          <div style="margin-top: 10px;">
            <h2 style="font-size: 1.15rem; font-weight: 800; font-family: var(--font-display); display: flex; align-items: center; gap: 6px;">
              ${user.name}
              <i class="fa-solid fa-circle-check badge-verified"></i>
            </h2>
            <div style="font-size: 0.8rem; color: var(--text-muted);">@${user.username}</div>
            
            <div style="display: flex; gap: 6px; flex-wrap: wrap; margin-top: 8px;">
              <span style="display: inline-flex; align-items: center; gap: 6px; background: rgba(var(--brand-primary-rgb), 0.15); color: var(--brand-primary); padding: 4px 10px; border-radius: 12px; font-size: 0.72rem; font-weight: 700;">
                <i class="fa-solid fa-shield-halved"></i> 5D Mesh Node Master
              </span>
              <span style="display: inline-flex; align-items: center; gap: 6px; background: rgba(251, 146, 60, 0.15); color: #fb923c; padding: 4px 10px; border-radius: 12px; font-size: 0.72rem; font-weight: 700;">
                <i class="fa-solid fa-atom"></i> ${celestialEngine.userQuantumEnergy} Quantum Energy
              </span>
            </div>

            <p style="font-size: 0.84rem; color: var(--text-main); margin-top: 10px; line-height: 1.4;">
              5D Mesh Participant chasing celestial anomalies, aligning harmonic frequencies, and operating localized autonomous proximity nodes 🌌⚡
            </p>
          </div>

          <!-- Stats Grid -->
          <div style="display: flex; justify-content: space-around; margin-top: 16px; padding-top: 14px; border-top: 1px solid var(--border-color); text-align: center;">
            <div>
              <div style="font-weight: 800; font-size: 1.1rem; font-family: var(--font-display);">${celestialEngine.alignedObjects.length}</div>
              <div style="font-size: 0.72rem; color: var(--text-muted);">Aligned Objects</div>
            </div>
            <div>
              <div style="font-weight: 800; font-size: 1.1rem; font-family: var(--font-display);">2.4k</div>
              <div style="font-size: 0.72rem; color: var(--text-muted);">Mesh Peers</div>
            </div>
            <div>
              <div style="font-weight: 800; font-size: 1.1rem; font-family: var(--font-display); color: #06b6d4;">${user.karma}</div>
              <div style="font-size: 0.72rem; color: var(--text-muted);">Karma 🐬</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Aligned Celestial Catalog -->
      <div class="glass-card">
        <h3 style="font-size: 0.9rem; font-weight: 700; margin-bottom: 10px; display: flex; align-items: center; gap: 6px;">
          <i class="fa-solid fa-sun" style="color: #fb923c;"></i> Aligned Celestial Beacons Catalog (${celestialEngine.alignedObjects.length})
        </h3>
        
        ${celestialEngine.alignedObjects.length === 0 ? `
          <div style="font-size: 0.8rem; color: var(--text-muted); text-align: center; padding: 16px 0;">
            No aligned celestial objects yet. Open the <strong>Radar</strong> tab to find and align nearby celestial beacons!
          </div>
        ` : `
          <div style="display: flex; flex-direction: column; gap: 8px;">
            ${celestialEngine.alignedObjects.map(obj => `
              <div style="display: flex; justify-content: space-between; align-items: center; padding: 10px; background: rgba(255,255,255,0.04); border-radius: 12px; border-left: 3px solid ${obj.color};">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <i class="fa-solid ${obj.icon}" style="color: ${obj.color}; font-size: 1.1rem;"></i>
                  <div>
                    <strong style="font-size: 0.84rem;">${obj.name}</strong>
                    <div style="font-size: 0.68rem; color: var(--text-muted);">${obj.targetFreq}Hz • Aligned at ${obj.alignedAt}</div>
                  </div>
                </div>
                <span style="font-size: 0.75rem; color: #10b981; font-weight: 700;">+${obj.energy} QE</span>
              </div>
            `).join('')}
          </div>
        `}
      </div>
    </div>
  `;
}
