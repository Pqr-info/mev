/**
 * Sovereign-27 Control Room Cockpit
 * Unified Web UI for Mesh Health, Council Governance, Capability Graph, Telemetry & 49-Ticket Matrix
 */

export class SovereignControlRoom {
  constructor(containerId = 'mainAppContainer', apiBase = 'http://localhost:4000') {
    this.containerId = containerId;
    this.apiBase = apiBase;
    this.activeTab = 'mesh'; // 'mesh' | 'council' | 'graph' | 'telemetry' | 'tickets'
    this.meshData = null;
    this.councilAudit = [];
    this.graphData = null;
    this.telemetryData = [];
    this.ticketMatrix = [];
    this.midiState = null;
    this.skillsRegistry = {};
  }

  async loadData() {
    try {
      const [mRes, cRes, gRes, tRes, tkRes, mdRes, sRes] = await Promise.all([
        fetch(`${this.apiBase}/api/gmi/mesh/health`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/skills/council/audit`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/skills/graph`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/skills/telemetry`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/tickets/matrix`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/midi/state`).then(r => r.json()).catch(() => null),
        fetch(`${this.apiBase}/api/gmi/skills/registry`).then(r => r.json()).catch(() => null)
      ]);

      if (mRes && mRes.ok) this.meshData = mRes;
      if (cRes && cRes.ok) this.councilAudit = cRes.audit;
      if (gRes && gRes.ok) this.graphData = gRes;
      if (tRes && tRes.ok) this.telemetryData = tRes.telemetry;
      if (tkRes && tkRes.ok) this.ticketMatrix = tkRes.tickets;
      if (mdRes && mdRes.ok) this.midiState = mdRes;
      if (sRes && sRes.ok) this.skillsRegistry = sRes.registry;
    } catch (err) {
      console.warn('[Control Room Load Warning]', err.message);
    }
  }

  async handleTransition(name, targetStatus) {
    try {
      const res = await fetch(`${this.apiBase}/api/gmi/skills/council/transition`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, targetStatus, approvedBy: 'council.admin', rationale: `Control Room UI Action: ${targetStatus.toUpperCase()}` })
      }).then(r => r.json());

      if (res.ok) {
        alert(`⚡ Skill '${name}' transitioned to '${targetStatus.toUpperCase()}' (Ticket #${res.ticketId})`);
        await this.loadData();
        this.render();
      } else {
        alert(`[-] Transition failed: ${res.error}`);
      }
    } catch (e) {
      alert(`[-] API Error: ${e.message}`);
    }
  }

  render() {
    const el = document.getElementById(this.containerId);
    if (!el) return;

    el.innerHTML = `
      <div style="padding: 24px; color: #f8fafc; font-family: system-ui, -apple-system, sans-serif; background: #090d16; min-height: 100vh;">
        
        <!-- Header -->
        <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px; border-bottom: 1px solid rgba(255,255,255,0.1); padding-bottom: 16px;">
          <div>
            <h1 style="margin: 0; font-size: 24px; font-weight: 800; background: linear-gradient(135deg, #06b6d4, #a855f7); -webkit-background-clip: text; -webkit-text-fill-color: transparent;">
              👑 Sovereign-27 Control Room Cockpit
            </h1>
            <div style="color: #94a3b8; font-size: 13px; margin-top: 4px;">
              Master Node: <span style="color:#10b981; font-weight:600;">max</span> | 5D IPv6 Subnet: <code style="color:#06b6d4;">fd5d:2700:4900::/48</code>
            </div>
          </div>
          <button id="btnRefreshControlRoom" style="background: rgba(6, 182, 212, 0.2); border: 1px solid #06b6d4; color: #38bdf8; padding: 8px 16px; border-radius: 8px; cursor: pointer; font-weight: 600;">
            🔄 Refresh Data
          </button>
        </div>

        <!-- Navigation Tabs -->
        <div style="display: flex; gap: 8px; margin-bottom: 24px; border-bottom: 1px solid rgba(255,255,255,0.05); padding-bottom: 12px;">
          ${this.renderTabBtn('mesh', '🌐 Mesh Health')}
          ${this.renderTabBtn('council', '⚖️ Council & Governance')}
          ${this.renderTabBtn('graph', '🕸️ Capability Graph')}
          ${this.renderTabBtn('telemetry', '📊 Telemetry & Drift')}
          ${this.renderTabBtn('tickets', '🎹 MIDI & 49-Tickets')}
        </div>

        <!-- Active View Panel -->
        <div id="controlRoomPanel">
          ${this.renderActivePanel()}
        </div>
      </div>
    `;

    this.bindEvents();
  }

  renderTabBtn(tabKey, label) {
    const active = this.activeTab === tabKey;
    return `
      <button class="cr-tab-btn ${active ? 'active' : ''}" data-tab="${tabKey}"
        style="padding: 10px 18px; border-radius: 10px; font-weight: 600; font-size: 14px; cursor: pointer; border: 1px solid ${active ? '#a855f7' : 'rgba(255,255,255,0.1)'}; background: ${active ? 'linear-gradient(135deg, rgba(168,85,247,0.3), rgba(6,182,212,0.3))' : 'rgba(255,255,255,0.03)'}; color: ${active ? '#fff' : '#94a3b8'};">
        ${label}
      </button>
    `;
  }

  renderActivePanel() {
    switch (this.activeTab) {
      case 'mesh': return this.renderMeshPanel();
      case 'council': return this.renderCouncilPanel();
      case 'graph': return this.renderGraphPanel();
      case 'telemetry': return this.renderTelemetryPanel();
      case 'tickets': return this.renderTicketsPanel();
      default: return this.renderMeshPanel();
    }
  }

  renderMeshPanel() {
    const nodes = this.meshData?.nodes || [];
    const tunnel = this.meshData?.tunnel || {};
    const cockroach = this.meshData?.cockroachdb || {};

    return `
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 20px;">
        
        <!-- Node Status Cards -->
        ${nodes.map(n => `
          <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(255,255,255,0.1); border-radius: 14px; padding: 20px; backdrop-filter: blur(10px);">
            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:12px;">
              <h3 style="margin:0; font-size:18px; color:#f8fafc;">${n.id}</h3>
              <span style="background: ${n.status === 'ONLINE' ? 'rgba(16,185,129,0.2)' : 'rgba(245,158,11,0.2)'}; color: ${n.status === 'ONLINE' ? '#10b981' : '#f59e0b'}; padding: 4px 10px; border-radius: 20px; font-size:12px; font-weight:700;">
                ${n.status}
              </span>
            </div>
            <div style="color:#94a3b8; font-size:13px; line-height:1.6;">
              <div>Role: <strong style="color:#e2e8f0;">${n.role}</strong></div>
              <div>IP: <code>${n.ip}</code></div>
              <div>5D Address: <code style="color:#06b6d4;">${n.ipv6_5d}</code></div>
              <div>Latency: <strong style="color:#a855f7;">${n.latency}</strong></div>
            </div>
          </div>
        `).join('')}

        <!-- 5D Route Panel -->
        <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(6,182,212,0.3); border-radius: 14px; padding: 20px;">
          <h3 style="margin:0 0 12px 0; font-size:18px; color:#06b6d4;">🛰️ 5D IPv6 Mesh Tunnel</h3>
          <div style="color:#94a3b8; font-size:13px; line-height:1.6;">
            <div>Subnet: <code>${tunnel.subnet || 'fd5d:2700:4900::/48'}</code></div>
            <div>Active Route: <code style="color:#38bdf8;">${tunnel.active_route || 'fd5d:2700:4900:0002::1 -> fd5d:2700:4900:0001::1'}</code></div>
            <div>Status: <span style="color:#10b981; font-weight:700;">${tunnel.status || 'AUTO_RECONNECT_ENABLED'}</span></div>
          </div>
        </div>

        <!-- CockroachDB Replication Panel -->
        <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(168,85,247,0.3); border-radius: 14px; padding: 20px;">
          <h3 style="margin:0 0 12px 0; font-size:18px; color:#a855f7;">🪳 CockroachDB Cluster (zeta.mh)</h3>
          <div style="color:#94a3b8; font-size:13px; line-height:1.6;">
            <div>Target Host: <code>${cockroach.cluster || 'zeta.mh:26257'}</code></div>
            <div>Database: <code>${cockroach.database || 'substrate27_midi'}</code></div>
            <div>Replication Queue: <strong style="color:#10b981;">${cockroach.replication_queue ?? 0} queued</strong></div>
            <div>Bridge Mode: <span style="color:#fb923c; font-weight:700;">${cockroach.status || 'WAL_BUFFERED'}</span></div>
          </div>
        </div>

      </div>
    `;
  }

  renderCouncilPanel() {
    const skills = Object.keys(this.skillsRegistry);

    return `
      <div style="display:flex; flex-direction:column; gap: 24px;">
        
        <!-- Skill Lifecycle Governance Table -->
        <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(255,255,255,0.1); border-radius: 14px; padding: 20px;">
          <h3 style="margin:0 0 16px 0; font-size:18px; color:#f8fafc;">⚖️ Mesh Skill Lifecycle Governance</h3>
          <table style="width:100%; border-collapse:collapse; color:#cbd5e1; font-size:14px; text-align:left;">
            <thead>
              <tr style="border-bottom: 1px solid rgba(255,255,255,0.1); color:#94a3b8;">
                <th style="padding:10px;">Skill</th>
                <th style="padding:10px;">Version</th>
                <th style="padding:10px;">Status</th>
                <th style="padding:10px;">Owner</th>
                <th style="padding:10px;">Council Action</th>
              </tr>
            </thead>
            <tbody>
              ${skills.map(sName => {
                const m = this.skillsRegistry[sName]?.manifest || {};
                const st = m.status || 'active';
                return `
                  <tr style="border-bottom: 1px solid rgba(255,255,255,0.05);">
                    <td style="padding:12px; font-weight:700; color:#06b6d4;">${sName}</td>
                    <td style="padding:12px;">v${m.version || '1.0.0'}</td>
                    <td style="padding:12px;">
                      <span style="background: ${st === 'active' ? 'rgba(16,185,129,0.2)' : st === 'pending' ? 'rgba(245,158,11,0.2)' : 'rgba(239,68,68,0.2)'}; color: ${st === 'active' ? '#10b981' : st === 'pending' ? '#f59e0b' : '#ef4444'}; padding:3px 8px; border-radius:12px; font-size:12px; font-weight:700;">
                        ${st.toUpperCase()}
                      </span>
                    </td>
                    <td style="padding:12px;">${m.owner || 'council'}</td>
                    <td style="padding:12px; display:flex; gap:6px;">
                      ${st !== 'active' ? `<button class="btn-transition" data-skill="${sName}" data-target="active" style="background:#10b981; color:#000; border:none; padding:4px 10px; border-radius:6px; font-weight:700; cursor:pointer;">Approve (Active)</button>` : ''}
                      ${st !== 'deprecated' ? `<button class="btn-transition" data-skill="${sName}" data-target="deprecated" style="background:#f59e0b; color:#000; border:none; padding:4px 10px; border-radius:6px; font-weight:700; cursor:pointer;">Deprecate</button>` : ''}
                      ${st !== 'revoked' ? `<button class="btn-transition" data-skill="${sName}" data-target="revoked" style="background:#ef4444; color:#fff; border:none; padding:4px 10px; border-radius:6px; font-weight:700; cursor:pointer;">Revoke</button>` : ''}
                    </td>
                  </tr>
                `;
              }).join('')}
            </tbody>
          </table>
        </div>

        <!-- Audit Log Stream -->
        <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(255,255,255,0.1); border-radius: 14px; padding: 20px;">
          <h3 style="margin:0 0 16px 0; font-size:18px; color:#a855f7;">📜 Council Audit Stream (skill_rollout_audit)</h3>
          <div style="max-height: 260px; overflow-y:auto; display:flex; flex-direction:column; gap:8px;">
            ${this.councilAudit.map(a => `
              <div style="background:rgba(0,0,0,0.4); padding:10px; border-radius:8px; border-left:3px solid #a855f7; font-size:13px; color:#e2e8f0;">
                <div style="display:flex; justify-content:space-between; color:#94a3b8; font-size:11px;">
                  <span>Ticket #${a.ticket_id < 10 ? '0' + a.ticket_id : a.ticket_id} | Council: ${a.approved_by}</span>
                  <span>${new Date(a.timestamp).toLocaleTimeString()}</span>
                </div>
                <div style="margin-top:4px; font-weight:600; color:#38bdf8;">${a.skill_name} v${a.version} -> ${a.status}</div>
              </div>
            `).join('')}
          </div>
        </div>

      </div>
    `;
  }

  renderGraphPanel() {
    const nodes = this.graphData?.nodes || [];
    const edges = this.graphData?.edges || [];

    return `
      <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(255,255,255,0.1); border-radius: 14px; padding: 20px;">
        <h3 style="margin:0 0 16px 0; font-size:18px; color:#06b6d4;">🕸️ Capability Graph Matrix</h3>
        <div style="color:#94a3b8; font-size:13px; margin-bottom:16px;">
          Nodes: <strong style="color:#fff;">${nodes.length}</strong> | Edges: <strong style="color:#fff;">${edges.length}</strong>
        </div>

        <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap:16px;">
          ${nodes.map(n => `
            <div style="background:rgba(0,0,0,0.5); padding:14px; border-radius:10px; border:1px solid ${n.type === 'agent' ? '#3b82f6' : n.type === 'skill' ? '#a855f7' : '#10b981'};">
              <div style="font-size:11px; text-transform:uppercase; font-weight:800; color:${n.type === 'agent' ? '#3b82f6' : n.type === 'skill' ? '#a855f7' : '#10b981'};">
                ${n.type}
              </div>
              <div style="font-weight:700; color:#f8fafc; font-size:15px; margin-top:4px;">${n.label}</div>
            </div>
          `).join('')}
        </div>
      </div>
    `;
  }

  renderTelemetryPanel() {
    return `
      <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(255,255,255,0.1); border-radius: 14px; padding: 20px;">
        <h3 style="margin:0 0 16px 0; font-size:18px; color:#10b981;">📊 Resolution Telemetry & Drift Stream</h3>
        <div style="display:flex; flex-direction:column; gap:10px; max-height:400px; overflow-y:auto;">
          ${this.telemetryData.map(t => `
            <div style="background:rgba(0,0,0,0.5); padding:12px; border-radius:8px; border-left:4px solid ${t.status === 'SUCCESS' ? '#10b981' : '#ef4444'}; font-size:13px;">
              <div style="display:flex; justify-content:space-between; color:#94a3b8; font-size:11px;">
                <span>Trace ID: <code>${t.telemetry_id}</code></span>
                <span>Ticket #${t.ticket_id}</span>
              </div>
              <div style="margin-top:6px; color:#f8fafc; font-weight:600;">
                Agent '${t.agent_id}' requested capability <code style="color:#06b6d4;">${t.capability_need}</code> ->
                <span style="color:${t.status === 'SUCCESS' ? '#10b981' : '#ef4444'};">${t.status}</span> (${t.resolved_skill || 'None'})
              </div>
            </div>
          `).join('')}
        </div>
      </div>
    `;
  }

  renderTicketsPanel() {
    const bpm = this.midiState?.bpm || 128.0;
    const channels = this.midiState?.activeChannels || 16;
    const tick = this.midiState?.currentTick || 96;

    return `
      <div style="display:flex; flex-direction:column; gap:24px;">
        
        <!-- MIDI Substrate Telemetry Header -->
        <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(6,182,212,0.3); border-radius: 14px; padding: 20px; display:flex; justify-content:space-between; align-items:center;">
          <div>
            <h3 style="margin:0; font-size:18px; color:#06b6d4;">🎹 MIDI State Machine Substrate</h3>
            <div style="color:#94a3b8; font-size:13px; margin-top:4px;">CockroachDB Replication: <code>zeta.mh:26257</code></div>
          </div>
          <div style="display:flex; gap:16px;">
            <div style="text-align:center;"><div style="color:#94a3b8; font-size:11px;">BPM</div><div style="font-size:20px; font-weight:800; color:#10b981;">${bpm}</div></div>
            <div style="text-align:center;"><div style="color:#94a3b8; font-size:11px;">Channels</div><div style="font-size:20px; font-weight:800; color:#38bdf8;">${channels}</div></div>
            <div style="text-align:center;"><div style="color:#94a3b8; font-size:11px;">Tick Clock</div><div style="font-size:20px; font-weight:800; color:#a855f7;">${tick}</div></div>
          </div>
        </div>

        <!-- 49-Ticket Grid -->
        <div style="background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(255,255,255,0.1); border-radius: 14px; padding: 20px;">
          <h3 style="margin:0 0 16px 0; font-size:18px; color:#f8fafc;">🧠 Active 49-Ticket Context Matrix (Tickets 0..48)</h3>
          <div style="display:grid; grid-template-columns: repeat(7, 1fr); gap:10px;">
            ${this.ticketMatrix.map(t => {
              const isRollout = t.isReservedRollout;
              return `
                <div style="background: ${isRollout ? 'rgba(168,85,247,0.2)' : 'rgba(0,0,0,0.5)'}; border: 1px solid ${isRollout ? '#a855f7' : 'rgba(255,255,255,0.1)'}; padding:10px; border-radius:8px; font-size:11px;">
                  <div style="display:flex; justify-content:space-between; font-weight:700; color:${isRollout ? '#a855f7' : '#06b6d4'};">
                    <span>#${t.ticketId < 10 ? '0' + t.ticketId : t.ticketId}</span>
                    ${isRollout ? '<span>[ROLLOUT]</span>' : ''}
                  </div>
                  <div style="margin-top:4px; color:#cbd5e1; height:32px; overflow:hidden; text-overflow:ellipsis;">
                    ${t.snippet}
                  </div>
                </div>
              `;
            }).join('')}
          </div>
        </div>

      </div>
    `;
  }

  bindEvents() {
    // Refresh button
    const btnRef = document.getElementById('btnRefreshControlRoom');
    if (btnRef) {
      btnRef.addEventListener('click', async () => {
        await this.loadData();
        this.render();
      });
    }

    // Tab buttons
    document.querySelectorAll('.cr-tab-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        this.activeTab = e.currentTarget.dataset.tab;
        this.render();
      });
    });

    // Governance transition buttons
    document.querySelectorAll('.btn-transition').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        const skill = e.currentTarget.dataset.skill;
        const target = e.currentTarget.dataset.target;
        await this.handleTransition(skill, target);
      });
    });
  }
}
