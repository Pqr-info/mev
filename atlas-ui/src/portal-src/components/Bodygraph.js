/**
 * Human Design Bodygraph & Relational Composite UI Component
 */

import { humanDesignEngine, HUMAN_DESIGN_PROFILES, ENERGY_CENTERS } from '../services/humanDesign.js';

export function renderBodygraphSection(activeTabMode = 'single', selectedPartnerId = 'peer_1', selectedGroupIds = ['u101', 'peer_1', 'peer_2']) {
  const currentUser = humanDesignEngine.getProfileByUserId('u101');

  return `
    <div class="bodygraph-section">
      <!-- Top Title & View Mode Selector -->
      <div class="glass-card" style="margin-bottom: 14px;">
        <div class="section-header" style="margin-bottom: 10px;">
          <span class="section-title">
            <i class="fa-solid fa-atom" style="color: var(--brand-primary); font-size: 1.1rem;"></i> Human Design Bodygraph
          </span>
          <span class="location-chip" style="font-size: 0.7rem;">
            <i class="fa-solid fa-dna"></i> Quantum Mechanics
          </span>
        </div>

        <!-- Tab Buttons -->
        <div style="display: flex; gap: 6px; margin-top: 10px;">
          <button class="btn-primary bodygraph-tab-btn" data-mode="single" style="flex: 1; padding: 8px; font-size: 0.78rem; ${activeTabMode === 'single' ? '' : 'background: rgba(255,255,255,0.08); box-shadow: none;'}">
            <i class="fa-solid fa-user"></i> Individual
          </button>
          <button class="btn-primary bodygraph-tab-btn" data-mode="couples" style="flex: 1; padding: 8px; font-size: 0.78rem; ${activeTabMode === 'couples' ? '' : 'background: rgba(255,255,255,0.08); box-shadow: none;'}">
            <i class="fa-solid fa-heart-pulse"></i> Couples Composite
          </button>
          <button class="btn-primary bodygraph-tab-btn" data-mode="group" style="flex: 1; padding: 8px; font-size: 0.78rem; ${activeTabMode === 'group' ? '' : 'background: rgba(255,255,255,0.08); box-shadow: none;'}">
            <i class="fa-solid fa-users-gear"></i> Team Penta
          </button>
        </div>
      </div>

      <!-- Main Bodygraph View Container -->
      ${renderBodygraphTabContent(activeTabMode, currentUser, selectedPartnerId, selectedGroupIds)}
    </div>
  `;
}

function renderBodygraphTabContent(mode, currentUser, selectedPartnerId, selectedGroupIds) {
  if (mode === 'couples') {
    const composite = humanDesignEngine.calculateCoupleComposite('u101', selectedPartnerId);
    return renderCouplesCompositeView(composite, selectedPartnerId);
  }

  if (mode === 'group') {
    const penta = humanDesignEngine.calculateGroupPenta(selectedGroupIds);
    return renderGroupPentaView(penta);
  }

  // Individual View
  return renderIndividualView(currentUser);
}

function renderIndividualView(user) {
  return `
    <!-- User Type Header -->
    <div class="glass-card" style="margin-bottom: 14px; background: linear-gradient(135deg, rgba(14,165,233,0.12) 0%, rgba(99,102,241,0.12) 100%);">
      <div style="display: flex; align-items: center; gap: 12px;">
        <img src="${user.avatar}" style="width: 52px; height: 52px; border-radius: 50%; border: 2px solid var(--brand-primary); object-fit: cover;" />
        <div>
          <h2 style="font-family: var(--font-display); font-size: 1.1rem; font-weight: 800;">${user.name}</h2>
          <div style="font-size: 0.82rem; color: var(--brand-primary); font-weight: 700; margin-top: 2px;">
            ⚡ ${user.type} • ${user.profile}
          </div>
          <div style="font-size: 0.72rem; color: var(--text-muted); margin-top: 2px;">
            Authority: <strong>${user.authority}</strong> • ${user.definition}
          </div>
        </div>
      </div>
    </div>

    <!-- SVG Bodygraph Chart Canvas -->
    <div class="glass-card" style="text-align: center; padding: 20px 10px;">
      <h3 style="font-size: 0.88rem; font-weight: 700; color: var(--text-muted); margin-bottom: 12px;">9 ENERGY CENTERS BODYGRAPH CHART</h3>
      ${renderSvgBodygraph(user.definedCenters)}
    </div>

    <!-- Centers Breakdown -->
    <div class="glass-card">
      <h3 style="font-size: 0.9rem; font-weight: 700; margin-bottom: 10px;">9 Energy Centers Breakdown</h3>
      <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px;">
        ${Object.keys(ENERGY_CENTERS).map(key => {
          const center = ENERGY_CENTERS[key];
          const isDefined = user.definedCenters.includes(key);
          return `
            <div style="padding: 8px 10px; border-radius: 10px; background: ${isDefined ? 'rgba(14,165,233,0.15)' : 'rgba(255,255,255,0.03)'}; border: 1px solid ${isDefined ? 'var(--brand-primary)' : 'var(--border-color)'};">
              <div style="font-weight: 700; font-size: 0.78rem; color: ${isDefined ? 'var(--brand-primary)' : 'var(--text-muted)'};">
                ${center.name}
              </div>
              <div style="font-size: 0.68rem; color: var(--text-dim); margin-top: 2px;">
                ${isDefined ? 'Defined (Active)' : 'Undefined (Open)'}
              </div>
            </div>
          `;
        }).join('')}
      </div>
    </div>
  `;
}

function renderCouplesCompositeView(composite, selectedPartnerId) {
  return `
    <!-- Partner Selection Dropdown -->
    <div class="glass-card" style="margin-bottom: 14px;">
      <label class="form-label">Select Partner / Couple Member</label>
      <select id="couplePartnerSelect" class="form-select">
        ${HUMAN_DESIGN_PROFILES.filter(p => p.userId !== 'u101').map(p => `
          <option value="${p.userId}" ${p.userId === selectedPartnerId ? 'selected' : ''}>
            ${p.name} (${p.type})
          </option>
        `).join('')}
      </select>
    </div>

    <!-- Couples Synergy Banner -->
    <div class="glass-card" style="margin-bottom: 14px; background: linear-gradient(135deg, rgba(244,63,94,0.15) 0%, rgba(168,85,247,0.15) 100%); border-color: rgba(244,63,94,0.4);">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <div>
          <h3 style="font-size: 1.05rem; font-weight: 800; font-family: var(--font-display); color: #f43f5e;">
            ${composite.personA.name} & ${composite.personB.name}
          </h3>
          <div style="font-size: 0.78rem; font-weight: 700; color: var(--text-main); margin-top: 2px;">
            Composite Theme: <span style="color: #a855f7;">${composite.relationshipType}</span>
          </div>
        </div>

        <div style="text-align: right;">
          <div style="font-size: 1.5rem; font-weight: 800; font-family: var(--font-display); color: #f43f5e;">
            ${composite.synergyScore}%
          </div>
          <span style="font-size: 0.65rem; color: var(--text-muted);">Synergy Score</span>
        </div>
      </div>
    </div>

    <!-- Composite SVG Bodygraph Chart -->
    <div class="glass-card" style="text-align: center; padding: 20px 10px;">
      <h3 style="font-size: 0.88rem; font-weight: 700; color: var(--text-muted); margin-bottom: 12px;">RELATIONAL COMPOSITE CHART OVERLAY</h3>
      ${renderSvgBodygraph(composite.combinedCenters, true)}
    </div>

    <!-- Electromagnetic & Connection Channels -->
    <div class="glass-card">
      <h3 style="font-size: 0.9rem; font-weight: 700; margin-bottom: 10px; display: flex; align-items: center; gap: 6px;">
        <i class="fa-solid fa-bolt" style="color: #f43f5e;"></i> Electromagnetic Connection Channels (${composite.electromagneticChannels.length})
      </h3>

      ${composite.electromagneticChannels.length === 0 ? `
        <div style="font-size: 0.78rem; color: var(--text-muted); padding: 10px 0;">
          No direct electromagnetic channel activations. Your relationship brings open flexibility!
        </div>
      ` : `
        <div style="display: flex; flex-direction: column; gap: 8px;">
          ${composite.electromagneticChannels.map(ch => `
            <div style="padding: 10px; background: rgba(244,63,94,0.1); border-radius: 12px; border-left: 3px solid #f43f5e; display: flex; justify-content: space-between; align-items: center;">
              <div>
                <strong style="font-size: 0.84rem;">${ch.name} (Gates ${ch.gates.join('-')})</strong>
                <div style="font-size: 0.7rem; color: var(--text-muted);">${ch.from} ⚡ ${ch.to}</div>
              </div>
              <span style="font-size: 0.72rem; color: #f43f5e; font-weight: 700;">Electromagnetic 🔥</span>
            </div>
          `).join('')}
        </div>
      `}
    </div>
  `;
}

function renderGroupPentaView(penta) {
  return `
    <div class="glass-card" style="margin-bottom: 14px; background: linear-gradient(135deg, rgba(16,185,129,0.15) 0%, rgba(6,182,212,0.15) 100%);">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <div>
          <h3 style="font-size: 1.05rem; font-weight: 800; font-family: var(--font-display); color: #10b981;">
            Group Team Penta Dynamic (${penta.membersCount} Nodes)
          </h3>
          <span style="font-size: 0.75rem; color: var(--text-main); font-weight: 700;">${penta.pentaStatus}</span>
        </div>

        <div style="text-align: right;">
          <div style="font-size: 1.5rem; font-weight: 800; font-family: var(--font-display); color: #10b981;">
            ${penta.synergyScore}%
          </div>
          <span style="font-size: 0.65rem; color: var(--text-muted);">Penta Synergy</span>
        </div>
      </div>
    </div>

    <!-- Group SVG Bodygraph Chart -->
    <div class="glass-card" style="text-align: center; padding: 20px 10px;">
      <h3 style="font-size: 0.88rem; font-weight: 700; color: var(--text-muted); margin-bottom: 12px;">GROUP PENTA COMPOSITE BODYGRAPH</h3>
      ${renderSvgBodygraph(penta.allCenters, false, true)}
    </div>

    <!-- Team Members Role Distribution -->
    <div class="glass-card">
      <h3 style="font-size: 0.9rem; font-weight: 700; margin-bottom: 10px;">Penta Team Composition</h3>
      <div style="display: flex; flex-direction: column; gap: 8px;">
        ${penta.profiles.map(member => `
          <div style="display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; background: rgba(255,255,255,0.04); border-radius: 10px;">
            <div style="display: flex; align-items: center; gap: 8px;">
              <img src="${member.avatar}" style="width: 32px; height: 32px; border-radius: 50%;" />
              <div>
                <div style="font-weight: 700; font-size: 0.84rem;">${member.name}</div>
                <div style="font-size: 0.7rem; color: var(--text-muted);">${member.type} • ${member.profile}</div>
              </div>
            </div>
            <span class="dist-badge" style="background: rgba(16,185,129,0.2); color: #10b981;">${member.definedCenters.length} Centers</span>
          </div>
        `).join('')}
      </div>
    </div>
  `;
}

/**
 * Interactive SVG 9 Energy Centers Bodygraph Renderer
 */
function renderSvgBodygraph(definedCenters = [], isComposite = false, isPenta = false) {
  const isDef = (center) => definedCenters.includes(center);

  const activeColor = isComposite ? '#f43f5e' : (isPenta ? '#10b981' : 'var(--brand-primary)');
  const defaultFill = 'rgba(255, 255, 255, 0.05)';
  const borderStroke = 'rgba(255, 255, 255, 0.2)';

  return `
    <svg viewBox="0 0 240 320" style="width: 100%; max-width: 260px; height: auto; margin: 0 auto; filter: drop-shadow(0 0 10px rgba(0,0,0,0.5));">
      <!-- Connecting Channel Lines -->
      <line x1="120" y1="35" x2="120" y2="65" stroke="${isDef('HEAD') && isDef('AJNA') ? activeColor : borderStroke}" stroke-width="3" />
      <line x1="120" y1="95" x2="120" y2="125" stroke="${isDef('AJNA') && isDef('THROAT') ? activeColor : borderStroke}" stroke-width="3" />
      <line x1="120" y1="155" x2="120" y2="175" stroke="${isDef('THROAT') && isDef('G_CENTER') ? activeColor : borderStroke}" stroke-width="3" />
      <line x1="120" y1="215" x2="120" y2="235" stroke="${isDef('G_CENTER') && isDef('SACRAL') ? activeColor : borderStroke}" stroke-width="3" />
      <line x1="120" y1="265" x2="120" y2="285" stroke="${isDef('SACRAL') && isDef('ROOT') ? activeColor : borderStroke}" stroke-width="3" />
      <line x1="120" y1="195" x2="180" y2="195" stroke="${isDef('G_CENTER') && isDef('HEART') ? activeColor : borderStroke}" stroke-width="3" />
      <line x1="120" y1="195" x2="195" y2="235" stroke="${isDef('G_CENTER') && isDef('SOLAR_PLEXUS') ? activeColor : borderStroke}" stroke-width="3" />
      <line x1="120" y1="195" x2="45" y2="235" stroke="${isDef('G_CENTER') && isDef('SPLEEN') ? activeColor : borderStroke}" stroke-width="3" />

      <!-- 1. HEAD CENTER (Triangle Up) -->
      <polygon points="120,15 95,45 145,45" fill="${isDef('HEAD') ? activeColor : defaultFill}" stroke="${isDef('HEAD') ? activeColor : borderStroke}" stroke-width="2" />
      <text x="120" y="34" font-size="8" fill="#fff" text-anchor="middle" font-weight="bold">HEAD</text>

      <!-- 2. AJNA CENTER (Triangle Down) -->
      <polygon points="95,65 145,65 120,95" fill="${isDef('AJNA') ? activeColor : defaultFill}" stroke="${isDef('AJNA') ? activeColor : borderStroke}" stroke-width="2" />
      <text x="120" y="78" font-size="8" fill="#fff" text-anchor="middle" font-weight="bold">AJNA</text>

      <!-- 3. THROAT CENTER (Square) -->
      <rect x="100" y="125" width="40" height="30" rx="4" fill="${isDef('THROAT') ? activeColor : defaultFill}" stroke="${isDef('THROAT') ? activeColor : borderStroke}" stroke-width="2" />
      <text x="120" y="143" font-size="8" fill="#fff" text-anchor="middle" font-weight="bold">THROAT</text>

      <!-- 4. G-CENTER (Diamond) -->
      <polygon points="120,175 140,195 120,215 100,195" fill="${isDef('G_CENTER') ? activeColor : defaultFill}" stroke="${isDef('G_CENTER') ? activeColor : borderStroke}" stroke-width="2" />
      <text x="120" y="198" font-size="8" fill="#fff" text-anchor="middle" font-weight="bold">G</text>

      <!-- 5. HEART / EGO CENTER (Small Triangle) -->
      <polygon points="170,185 190,195 170,205" fill="${isDef('HEART') ? activeColor : defaultFill}" stroke="${isDef('HEART') ? activeColor : borderStroke}" stroke-width="2" />

      <!-- 6. SPLEEN CENTER (Left Triangle) -->
      <polygon points="45,215 45,255 20,235" fill="${isDef('SPLEEN') ? activeColor : defaultFill}" stroke="${isDef('SPLEEN') ? activeColor : borderStroke}" stroke-width="2" />

      <!-- 7. SOLAR PLEXUS CENTER (Right Triangle) -->
      <polygon points="195,215 195,255 220,235" fill="${isDef('SOLAR_PLEXUS') ? activeColor : defaultFill}" stroke="${isDef('SOLAR_PLEXUS') ? activeColor : borderStroke}" stroke-width="2" />

      <!-- 8. SACRAL CENTER (Square) -->
      <rect x="105" y="235" width="30" height="30" rx="4" fill="${isDef('SACRAL') ? activeColor : defaultFill}" stroke="${isDef('SACRAL') ? activeColor : borderStroke}" stroke-width="2" />
      <text x="120" y="253" font-size="7" fill="#fff" text-anchor="middle" font-weight="bold">SACRAL</text>

      <!-- 9. ROOT CENTER (Bottom Square) -->
      <rect x="102" y="285" width="36" height="25" rx="4" fill="${isDef('ROOT') ? activeColor : defaultFill}" stroke="${isDef('ROOT') ? activeColor : borderStroke}" stroke-width="2" />
      <text x="120" y="301" font-size="7" fill="#fff" text-anchor="middle" font-weight="bold">ROOT</text>
    </svg>
  `;
}
