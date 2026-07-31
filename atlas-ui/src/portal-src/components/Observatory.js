/**
 * Sovereign-27 Observatory — Deep Introspection Telescope UI
 * Visualizes Lineage Evolution Maps, Skill Influence Propagation, Temporal Heatmaps, Mesh Replay & Config Reconciliation
 */

export class SovereignObservatory {
  constructor(containerId = 'observatoryRoot', apiBase = 'http://localhost:4000') {
    this.containerId = containerId;
    this.apiBase = apiBase;
    this.forecastData = null;
    this.driftData = null;
    this.telemetryData = [];
    this.lineageData = null;
    this.influenceData = null;
    this.heatmapData = null;
    this.anomaliesData = [];
    this.lineageDiffData = null;
    this.replayData = null;
  }

  async loadData() {
    try {
      const [fcRes, dfRes, tmRes, lnRes, infRes, hmRes, anRes, ldRes, rpRes] = await Promise.all([
        fetch(`${this.apiBase}/api/gmi/skills/forecast`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/mesh/drift`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/skills/telemetry`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/observatory/lineage`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/observatory/influence`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/observatory/heatmap`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/observatory/anomalies`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/observatory/lineage/diff`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/observatory/replay`).then(r => r.json()).catch(() => null)
      ]);

      if (fcRes && fcRes.ok) this.forecastData = fcRes;
      if (dfRes && dfRes.ok) this.driftData = dfRes;
      if (tmRes && tmRes.ok) this.telemetryData = tmRes.telemetry;
      if (lnRes && lnRes.ok) this.lineageData = lnRes;
      if (infRes && infRes.ok) this.influenceData = infRes;
      if (hmRes && hmRes.ok) this.heatmapData = hmRes;
      if (anRes && anRes.ok) this.anomaliesData = anRes.anomalies;
      if (ldRes && ldRes.ok) this.lineageDiffData = ldRes;
      if (rpRes && rpRes.ok) this.replayData = rpRes;
    } catch (err) {
      console.warn('[Observatory Load Error]', err.message);
    }
  }

  async handleReconciliation(node = 'ted') {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/mesh/reconcile`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ node, referenceNode: 'max' })
      }).then(r => r.json());

      if (res.ok) {
        alert(`⚡ Node '${node}' reconciled with gold-standard '${res.referenceNode}'! Digest aligned: ${res.reconciledDigest.substring(0, 12)}...`);
        await this.loadData();
        this.render();
      } else {
        alert(`[-] Reconciliation failed: ${res.error}`);
      }
    } catch (e) {
      alert(`[-] API Error: ${e.message}`);
    }
  }

  render() {
    const el = document.getElementById(this.containerId);
    if (!el) return;

    const stability = this.driftData?.mesh_stability_index || '99.5%';
    const driftScore = this.driftData?.drift_score ?? 0.05;
    const arbitrationStatus = this.driftData?.arbitration_status || 'STABLE_IN_CONSENSUS';
    const forecasts = this.forecastData?.forecasts || [];
    const cubes = this.driftData?.agent_cube_digests || {};
    const lineageNodes = this.lineageData?.lineage_nodes || [];
    const skillInfluence = this.influenceData?.skill_influence || [];
    const temporalGrid = this.heatmapData?.temporal_grid || [];

    el.innerHTML = `
      <div style="padding: 24px; color: #f8fafc; font-family: system-ui, -apple-system, sans-serif; background: #070a12; min-height: 100vh;">
        
        <!-- Header -->
        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px; border-bottom: 1px solid rgba(168,85,247,0.2); padding-bottom: 16px;">
          <div>
            <h1 style="margin: 0; font-size: 24px; font-weight: 800; background: linear-gradient(135deg, #a855f7, #ec4899); -webkit-background-clip: text; -webkit-text-fill-color: transparent;">
              🔭 Sovereign-27 Observatory Telescope
            </h1>
            <div style="color: #94a3b8; font-size: 13px; margin-top: 4px;">
              Lineage Diffing, Replay Engine & Automatic Config Reconciliation | Stability: <strong style="color: #10b981;">${stability}</strong>
            </div>
          </div>
          <div style="display:flex; gap:10px;">
            <button id="btnTriggerReconcile" style="background: rgba(16,185,129,0.2); border: 1px solid #10b981; color: #10b981; padding: 8px 16px; border-radius: 8px; cursor: pointer; font-weight: 700;">
              ⚡ Reconcile Node Drift
            </button>
            <button id="btnRefreshObservatory" style="background: rgba(168,85,247,0.2); border: 1px solid #a855f7; color: #c084fc; padding: 8px 16px; border-radius: 8px; cursor: pointer; font-weight: 600;">
              🔄 Refresh Telescope
            </button>
          </div>
        </div>

        <!-- Replay Control Bar -->
        <div style="background: rgba(15, 23, 42, 0.9); border: 1px solid rgba(6,182,212,0.3); border-radius: 12px; padding: 12px 20px; margin-bottom: 24px; display:flex; justify-content:space-between; align-items:center;">
          <div style="display:flex; align-items:center; gap:12px;">
            <span style="background:#06b6d4; color:#000; font-weight:800; font-size:11px; padding:3px 8px; border-radius:12px;">
              ${this.replayData?.replay_mode || 'ACTIVE_PLAYBACK'}
            </span>
            <span style="color:#e2e8f0; font-size:13px; font-weight:600;">
              ▶ 24-Hour Mesh Activity Replay Sequence (${this.replayData?.events_count || 24} Events)
            </span>
          </div>
          <div style="color:#06b6d4; font-weight:700; font-size:13px;">Speed: 1x Real-Time</div>
        </div>

        <!-- Section 1: Lineage Evolution Map & Skill Influence -->
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 24px; margin-bottom: 24px;">
          
          <!-- Lineage Evolution Map -->
          <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(168,85,247,0.3); border-radius: 14px; padding: 20px;">
            <h3 style="margin:0 0 16px 0; font-size:18px; color:#a855f7;">🌳 Agent Lineage Evolution Tree</h3>
            <div style="display:flex; flex-direction:column; gap:10px;">
              ${lineageNodes.map(n => `
                <div style="background:rgba(0,0,0,0.5); padding:10px 14px; border-radius:10px; border-left:4px solid ${n.generation === 0 ? '#10b981' : n.generation === 1 ? '#06b6d4' : '#a855f7'}; display:flex; justify-content:space-between; align-items:center;">
                  <div>
                    <div style="font-weight:700; color:#f8fafc; font-size:14px;">${n.agent_name}</div>
                    <div style="font-size:11px; color:#94a3b8;">Gen ${n.generation} | Node: ${n.node_id}</div>
                  </div>
                  <span style="background:rgba(255,255,255,0.05); color:#cbd5e1; padding:3px 8px; border-radius:12px; font-size:11px; font-weight:600;">
                    ${n.capabilities_count} Caps
                  </span>
                </div>
              `).join('')}
            </div>
          </div>

          <!-- Skill Influence Propagation -->
          <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(6,182,212,0.3); border-radius: 14px; padding: 20px;">
            <h3 style="margin:0 0 16px 0; font-size:18px; color:#06b6d4;">🌊 Skill Influence Propagation Graph</h3>
            <div style="display:flex; flex-direction:column; gap:12px;">
              ${skillInfluence.map(s => `
                <div style="background:rgba(0,0,0,0.5); padding:12px; border-radius:10px;">
                  <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
                    <strong style="color:#f8fafc; font-size:14px;">${s.skill_name} v${s.version}</strong>
                    <span style="color:#06b6d4; font-weight:800; font-size:13px;">${s.propagation_reach}</span>
                  </div>
                  <div style="width:100%; background:rgba(255,255,255,0.1); height:8px; border-radius:4px; overflow:hidden;">
                    <div style="width:${s.propagation_reach}; background:linear-gradient(90deg, #06b6d4, #a855f7); height:100%;"></div>
                  </div>
                </div>
              `).join('')}
            </div>
          </div>

        </div>

        <!-- Section 2: Temporal Stability Heatmap (24 Hours) & Anomalies -->
        <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(16,185,129,0.3); border-radius: 14px; padding: 20px; margin-bottom: 24px;">
          <h3 style="margin:0 0 16px 0; font-size:18px; color:#10b981;">📅 24-Hour Temporal Stability Matrix Grid</h3>
          <div style="display:grid; grid-template-columns: repeat(12, 1fr); gap:8px;">
            ${temporalGrid.map(h => `
              <div style="background: ${h.stability_percentage >= 98 ? 'rgba(16,185,129,0.25)' : 'rgba(245,158,11,0.25)'}; border:1px solid ${h.stability_percentage >= 98 ? '#10b981' : '#f59e0b'}; padding:8px; border-radius:8px; text-align:center;">
                <div style="font-size:10px; color:#94a3b8; font-weight:700;">${h.hour_label}</div>
                <div style="font-size:13px; font-weight:800; color:${h.stability_percentage >= 98 ? '#10b981' : '#f59e0b'}; margin-top:2px;">
                  ${h.stability_percentage}%
                </div>
                <div style="font-size:9px; color:#cbd5e1; margin-top:2px;">${h.resolutions} res</div>
              </div>
            `).join('')}
          </div>
        </div>

        <!-- Section 3: Capability Forecasting & Vector Drift -->
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 24px;">
          
          <!-- Forecasting -->
          <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(6,182,212,0.3); border-radius: 14px; padding: 20px;">
            <h3 style="margin:0 0 16px 0; font-size:18px; color:#06b6d4;">🔮 Mesh Capability Forecasting</h3>
            <div style="display:flex; flex-direction:column; gap:10px;">
              ${forecasts.map(f => `
                <div style="background:rgba(0,0,0,0.5); padding:10px; border-radius:8px; border-left:3px solid #06b6d4;">
                  <div style="display:flex; justify-content:space-between;">
                    <strong style="color:#f8fafc; font-size:13px;">${f.capability}</strong>
                    <span style="color:#10b981; font-weight:700; font-size:11px;">${(f.probability * 100).toFixed(0)}%</span>
                  </div>
                  <div style="font-size:11px; color:#94a3b8; margin-top:2px;">Target: ${f.target_skill} | ${f.rationale}</div>
                </div>
              `).join('')}
            </div>
          </div>

          <!-- Drift Arbitration -->
          <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(236,72,153,0.3); border-radius: 14px; padding: 20px;">
            <h3 style="margin:0 0 16px 0; font-size:18px; color:#ec4899;">⚖️ Vector Drift Arbitration</h3>
            <div style="color:#cbd5e1; font-size:13px; font-weight:700; margin-bottom:8px;">Agent Cube Vector Digests:</div>
            <div style="display:flex; flex-direction:column; gap:8px;">
              ${Object.keys(cubes).map(aid => `
                <div style="background:rgba(0,0,0,0.4); padding:10px; border-radius:8px; font-size:11px;">
                  <div style="color:#38bdf8; font-weight:700;">${aid}</div>
                  <code style="color:#cbd5e1; font-size:10px; word-break:break-all;">${cubes[aid]}</code>
                </div>
              `).join('')}
            </div>
          </div>

        </div>

      </div>
    `;

    this.bindEvents();
  }

  bindEvents() {
    const btnRef = document.getElementById('btnRefreshObservatory');
    if (btnRef) {
      btnRef.addEventListener('click', async () => {
        await this.loadData();
        this.render();
      });
    }

    const btnRec = document.getElementById('btnTriggerReconcile');
    if (btnRec) {
      btnRec.addEventListener('click', async () => {
        await this.handleReconciliation('ted');
      });
    }
  }
}
