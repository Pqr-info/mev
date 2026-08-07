/**
 * Sovereign-27 Official Specification & Substrate Architecture Portal
 * Authoritative Node: zeta.mh (46.224.219.174) | 5D IPv6: fd5d:2700:4900::5
 */

export function renderSovereignPortal(subTab = 'portal', telemetryData = null, bleDevices = []) {
  const telemetryObj = telemetryData || {
    timestamp: new Date().toISOString(),
    active_endpoints_count: 108,
    master_node: "max",
    remote_node: "zeta.mh (46.224.219.174)",
    five_d_ipv6: "fd5d:2700:4900::5",
    t_now_authoritative_state: {
      tnt_id: "tnt_max_8",
      agent_id: "max",
      t_now_sequence: 8,
      cumulative_work: 92880000000,
      active_epoch: 4,
      tnt_state_hash: "0xd4d45ef0e25f",
      status: "T_NOW_ACTIVE",
      timestamp: 1785487635129
    },
    pqr_latest_record: {
      pqr_id: "pqr_max_7_to_8",
      agent_id: "max",
      alpha_t_now_seq: 7,
      omega_t_next_seq: 8,
      delta_work_seu: 11610000000,
      qualification_score: 1,
      pqr_sha256_hash: "0xd1ff5936c872",
      status: "PRE_QUALIFIED_RECORD_VALID",
      timestamp: 1785487635129
    },
    pqr_root_chain_latest: {
      root_height: 5,
      agent_id: "max",
      pqr_id: "pqr_max_7_to_8",
      previous_root_hash: "0xe400f793a6d38b86a58d6c106d278da5f7e5d9b5c7a7c8fc152814b68ad6cf75",
      current_root_hash: "0x4c3e4435795a64032f010133ae8faf5d61a35453274d33b1e335752e36ea3980",
      pqr_sha256_hash: "0xd1ff5936c872",
      status: "PQR_ROOT_BOUND_VALID",
      timestamp: 1785487635129
    },
    pqr_oro_latest_cycle: {
      oro_cycle_id: "oro_max_cycle_8",
      agent_id: "max",
      alpha_t_now_seq: 7,
      omega_t_next_seq: 8,
      committed_work_w: 92880000000,
      oro_root_hash: "0x4c3e4435795a64032f010133ae8faf5d61a35453274d33b1e335752e36ea3980",
      status: "ORO_CYCLE_COMPLETE_VALID",
      timestamp: 1785487635129
    },
    governance_latest_proposal: {
      proposal_id: "prop_max_q_threshold_1785487635139",
      proposer_agent: "max",
      parameter_key: "Q_THRESHOLD",
      proposed_value: "0.9500",
      votes_for: 5,
      votes_against: 0,
      status: "GOV_ORO_ENACTED_ACTIVE",
      timestamp: 1785487635141
    },
    dolphin_safe_mesh_health: {
      cert_id: "ds_cert_max_1785487635134",
      agent_id: "max",
      dolphin_safe_score: 0.8628,
      efficiency_eta: 0.9804,
      fft_spike_level: 0.12,
      root_height: 5,
      certification_hash: "0x0c3f639946bc",
      status: "CERTIFIED_DOLPHIN_SAFE_NEURAL_MESH_ACTIVE",
      timestamp: 1785487635134
    }
  };

  return `
    <div class="sovereign-portal-container" style="padding: 16px; display: flex; flex-direction: column; gap: 16px; color: var(--text-main);">
      
      <!-- Sub-Tab Navigation Header -->
      <div style="display: flex; gap: 6px; overflow-x: auto; padding-bottom: 4px; border-bottom: 1px solid var(--border-color);">
        <button class="portal-subtab-btn ${subTab === 'portal' ? 'active' : ''}" data-subtab="portal" style="padding: 6px 12px; border-radius: 20px; font-size: 0.75rem; font-weight: 700; background: ${subTab === 'portal' ? '#0ea5e9' : 'rgba(255,255,255,0.05)'}; color: ${subTab === 'portal' ? '#fff' : 'var(--text-muted)'}; border: none; cursor: pointer;">
          <i class="fa-solid fa-cloud-bolt"></i> Portal
        </button>
        <button class="portal-subtab-btn ${subTab === 'wiki' ? 'active' : ''}" data-subtab="wiki" style="padding: 6px 12px; border-radius: 20px; font-size: 0.75rem; font-weight: 700; background: ${subTab === 'wiki' ? '#0ea5e9' : 'rgba(255,255,255,0.05)'}; color: ${subTab === 'wiki' ? '#fff' : 'var(--text-muted)'}; border: none; cursor: pointer;">
          <i class="fa-solid fa-book-bookmark"></i> Wiki
        </button>
        <button class="portal-subtab-btn ${subTab === 'control' ? 'active' : ''}" data-subtab="control" style="padding: 6px 12px; border-radius: 20px; font-size: 0.75rem; font-weight: 700; background: ${subTab === 'control' ? '#0ea5e9' : 'rgba(255,255,255,0.05)'}; color: ${subTab === 'control' ? '#fff' : 'var(--text-muted)'}; border: none; cursor: pointer;">
          <i class="fa-solid fa-sliders"></i> Control Room
        </button>
        <button class="portal-subtab-btn ${subTab === 'observatory' ? 'active' : ''}" data-subtab="observatory" style="padding: 6px 12px; border-radius: 20px; font-size: 0.75rem; font-weight: 700; background: ${subTab === 'observatory' ? '#0ea5e9' : 'rgba(255,255,255,0.05)'}; color: ${subTab === 'observatory' ? '#fff' : 'var(--text-muted)'}; border: none; cursor: pointer;">
          <i class="fa-solid fa-satellite-dish"></i> Observatory
        </button>
        <button class="portal-subtab-btn ${subTab === 'radar' ? 'active' : ''}" data-subtab="radar" style="padding: 6px 12px; border-radius: 20px; font-size: 0.75rem; font-weight: 700; background: ${subTab === 'radar' ? '#0ea5e9' : 'rgba(255,255,255,0.05)'}; color: ${subTab === 'radar' ? '#fff' : 'var(--text-muted)'}; border: none; cursor: pointer;">
          <i class="fa-solid fa-compass"></i> Radar
        </button>
      </div>

      <!-- Bluetooth / Wi-Fi Proximity Scan Action Bar -->
      <div class="glass-card" style="padding: 14px; border-radius: 14px; background: linear-gradient(135deg, rgba(16, 185, 129, 0.12) 0%, rgba(14, 165, 233, 0.15) 100%); border: 1px solid rgba(16, 185, 129, 0.35);">
        <div style="display: flex; align-items: center; justify-content: space-between; gap: 10px;">
          <div>
            <div style="font-weight: 800; font-size: 0.9rem; color: #34d399; display: flex; align-items: center; gap: 8px;">
              <i class="fa-brands fa-bluetooth-b"></i> <i class="fa-solid fa-wifi"></i> Bluetooth &amp; Wi-Fi Direct Mesh
            </div>
            <div style="font-size: 0.7rem; color: var(--text-muted); margin-top: 2px;">
              Discover nearby physical mesh neighbors over BLE 5.3 &amp; Wi-Fi Direct P2P.
            </div>
          </div>
          <button id="btnScanBleWifiNeighbors" class="btn-primary" style="padding: 8px 14px; font-size: 0.75rem; font-weight: 800; border-radius: 10px; background: linear-gradient(135deg, #10b981, #0ea5e9); white-space: nowrap; display: flex; align-items: center; gap: 6px;">
            <i class="fa-solid fa-radar"></i> Scan Neighbors
          </button>
        </div>

        ${bleDevices.length > 0 ? `
          <div style="margin-top: 12px; display: flex; flex-direction: column; gap: 8px;" id="bleWifiResultsContainer">
            <div style="font-size: 0.7rem; font-weight: 700; color: #34d399;">
              Discovered ${bleDevices.length} Neighbor Node(s):
            </div>
            ${bleDevices.map(d => `
              <div style="background: rgba(0,0,0,0.4); padding: 8px 12px; border-radius: 8px; display: flex; align-items: center; justify-content: space-between; font-size: 0.72rem; border: 1px solid rgba(255,255,255,0.06);">
                <div>
                  <strong style="color: #fff;">${d.name}</strong>
                  <div style="font-size: 0.65rem; color: var(--text-muted);">MAC/UUID: ${d.mac} | Protocol: ${d.protocol}</div>
                </div>
                <div style="text-align: right;">
                  <span style="color: #10b981; font-weight: 700;">${d.distanceMeters}m away</span>
                  <div style="font-size: 0.65rem; color: #fb923c;">RSSI: ${d.rssi} dBm</div>
                </div>
              </div>
            `).join('')}
          </div>
        ` : ''}
      </div>

      <!-- Authoritative Wiki Header Card -->
      <div class="glass-card" style="background: linear-gradient(135deg, rgba(14, 165, 233, 0.15) 0%, rgba(99, 102, 241, 0.25) 100%); border: 1px solid rgba(14, 165, 233, 0.4); border-radius: 16px; padding: 18px;">
        <div style="font-size: 0.65rem; color: #38bdf8; text-transform: uppercase; font-weight: 800; letter-spacing: 0.5px;">
          OFFICIAL SPECIFICATION &amp; SUBSTRATE ARCHITECTURE
        </div>
        <div style="font-family: var(--font-display); font-size: 1.15rem; font-weight: 800; color: #fff; margin-top: 2px;">
          PQR Architectural Wiki &mdash; <code>wiki.pqr.info</code>
        </div>
        <p style="font-size: 0.75rem; color: var(--text-muted); margin-top: 6px; line-height: 1.4;">
          <strong>PQR = Pre-Qualified Record</strong>. Sovereign-27 is a self-referential, non-destructive, hash-verified temporal logic mesh running across <strong>108 backend REST endpoints</strong> and multi-node Hetzner Threadripper architecture.
        </p>

        <div style="margin-top: 10px; padding: 8px 12px; background: rgba(0,0,0,0.4); border-radius: 8px; display: flex; align-items: center; justify-content: space-between; font-size: 0.7rem; font-family: monospace;">
          <div>
            <span style="color: var(--text-muted);">Authoritative Node:</span>
            <strong style="color: #10b981;">zeta.mh (46.224.219.174)</strong>
          </div>
          <div style="color: #38bdf8;">
            5D IPv6: <code>fd5d:2700:4900::5</code>
          </div>
        </div>
      </div>

      <!-- Wiki Navigation Grid (Interactive Buttons) -->
      <div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px;">
        <button id="btnWikiOverview" class="glass-card wiki-nav-card" style="padding: 10px; border-radius: 10px; text-align: left; background: rgba(255,255,255,0.03); border: 1px solid var(--border-color); cursor: pointer;">
          <div style="font-size: 0.75rem; font-weight: 700; color: #38bdf8;">1. Overview &amp; Philosophy</div>
        </button>
        <button id="btnWikiArchitecture" class="glass-card wiki-nav-card" style="padding: 10px; border-radius: 10px; text-align: left; background: rgba(255,255,255,0.03); border: 1px solid var(--border-color); cursor: pointer;">
          <div style="font-size: 0.75rem; font-weight: 700; color: #818cf8;">2. 5-Layer Stack Architecture</div>
        </button>
        <button id="btnWikiTemporalEconomy" class="glass-card wiki-nav-card" style="padding: 10px; border-radius: 10px; text-align: left; background: rgba(255,255,255,0.03); border: 1px solid var(--border-color); cursor: pointer;">
          <div style="font-size: 0.75rem; font-weight: 700; color: #c084fc;">3. SEU Temporal Economy</div>
        </button>
        <button id="btnWikiOuroborosLoop" class="glass-card wiki-nav-card" style="padding: 10px; border-radius: 10px; text-align: left; background: rgba(255,255,255,0.03); border: 1px solid var(--border-color); cursor: pointer;">
          <div style="font-size: 0.75rem; font-weight: 700; color: #34d399;">4. PQR-ORO Ouroboros Loop</div>
        </button>
      </div>

      <!-- Live Telemetry Inspector Card -->
      <div class="glass-card" style="padding: 16px; border-radius: 16px; border: 1px solid rgba(14, 165, 233, 0.3);">
        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px;">
          <div style="font-size: 0.85rem; font-weight: 700; color: #38bdf8; display: flex; align-items: center; gap: 8px;">
            <i class="fa-solid fa-satellite-dish fa-spin" style="--fa-animation-duration: 6s;"></i> Live Telemetry Inspector
          </div>
          <button id="btnRefreshPortalTelemetry" class="btn-primary" style="padding: 4px 10px; font-size: 0.7rem; border-radius: 8px;">
            <i class="fa-solid fa-arrows-rotate"></i> Refresh Telemetry
          </button>
        </div>

        <p style="font-size: 0.7rem; color: var(--text-muted); margin-bottom: 10px;">
          Live Backend Endpoint Telemetry Inspector (108 REST Endpoints Active)
        </p>

        <div id="telemetryJsonBox" style="background: #050b14; border: 1px solid rgba(255, 255, 255, 0.1); padding: 12px; border-radius: 10px; font-family: monospace; font-size: 0.68rem; color: #10b981; max-height: 280px; overflow-y: auto; white-space: pre-wrap; word-break: break-all; line-height: 1.35;">
${JSON.stringify(telemetryObj, null, 2)}
        </div>
      </div>

    </div>
  `;
}
