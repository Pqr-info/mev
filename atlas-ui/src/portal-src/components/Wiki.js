/**
 * Sovereign-27 Stack & PQR Substrate Wiki Component (wiki.pqr.info)
 * High-Aesthetic Interactive Documentation & Telemetry Inspector
 */

export function renderWikiView() {
  return `
    <div class="pqr-wiki-container" style="padding: 24px; color: #f8fafc; font-family: 'Outfit', 'Inter', sans-serif; background: #070a13; min-height: 100vh;">
      
      <!-- Wiki Top Hero Banner -->
      <div style="background: linear-gradient(135deg, rgba(14, 165, 233, 0.15), rgba(168, 85, 247, 0.15)); border: 1px solid rgba(14, 165, 233, 0.3); border-radius: 16px; padding: 32px; margin-bottom: 32px; backdrop-filter: blur(12px);">
        <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px;">
          <div>
            <div style="display: inline-flex; align-items: center; gap: 8px; background: rgba(14, 165, 233, 0.2); border: 1px solid #0ea5e9; padding: 6px 14px; border-radius: 20px; font-size: 0.82rem; font-weight: 700; color: #38bdf8; margin-bottom: 12px;">
              <i class="fa-solid fa-book-bookmark"></i> OFFICIAL SPECIFICATION & SUBSTRATE ARCHITECTURE
            </div>
            <h1 style="font-size: 2.2rem; font-weight: 800; margin: 0; background: linear-gradient(90deg, #38bdf8, #a855f7, #ec4899); -webkit-background-clip: text; -webkit-text-fill-color: transparent;">
              PQR Architectural Wiki — wiki.pqr.info
            </h1>
            <p style="color: #94a3b8; font-size: 1.05rem; margin-top: 8px; max-width: 780px; line-height: 1.6;">
              <strong>PQR = Pre-Qualified Record</strong>. Sovereign-27 is a self-referential, non-destructive, hash-verified temporal logic mesh running across 108 backend REST endpoints and multi-node Hetzner Threadripper architecture.
            </p>
          </div>
          <div style="text-align: right; background: rgba(15, 23, 42, 0.8); border: 1px solid rgba(255, 255, 255, 0.1); padding: 18px 24px; border-radius: 14px;">
            <div style="font-size: 0.85rem; color: #94a3b8;">Authoritative Node</div>
            <div style="font-size: 1.2rem; font-weight: 800; color: #10b981;">zeta.mh (46.224.219.174)</div>
            <div style="font-size: 0.8rem; color: #64748b; margin-top: 4px;">5D IPv6: fd5d:2700:4900::5</div>
          </div>
        </div>
      </div>

      <!-- Quick Navigation Tabs -->
      <div style="display: flex; gap: 12px; margin-bottom: 32px; overflow-x: auto; padding-bottom: 8px;">
        <button class="pqr-wiki-tab-btn active" onclick="switchWikiTab('overview')" style="background: rgba(14, 165, 233, 0.2); border: 1px solid #0ea5e9; color: #38bdf8; padding: 10px 20px; border-radius: 10px; font-weight: 700; cursor: pointer;">
          <i class="fa-solid fa-compass"></i> Overview & Philosophy
        </button>
        <button class="pqr-wiki-tab-btn" onclick="switchWikiTab('layers')" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(255, 255, 255, 0.1); color: #94a3b8; padding: 10px 20px; border-radius: 10px; font-weight: 700; cursor: pointer;">
          <i class="fa-solid fa-layer-group"></i> 5-Layer Stack Architecture
        </button>
        <button class="pqr-wiki-tab-btn" onclick="switchWikiTab('seu')" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(255, 255, 255, 0.1); color: #94a3b8; padding: 10px 20px; border-radius: 10px; font-weight: 700; cursor: pointer;">
          <i class="fa-solid fa-coins"></i> SEU Temporal Economy
        </button>
        <button class="pqr-wiki-tab-btn" onclick="switchWikiTab('oro')" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(255, 255, 255, 0.1); color: #94a3b8; padding: 10px 20px; border-radius: 10px; font-weight: 700; cursor: pointer;">
          <i class="fa-solid fa-rotate"></i> PQR-ORO Ouroboros Loop
        </button>
        <button class="pqr-wiki-tab-btn" onclick="switchWikiTab('endpoints')" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(255, 255, 255, 0.1); color: #94a3b8; padding: 10px 20px; border-radius: 10px; font-weight: 700; cursor: pointer;">
          <i class="fa-solid fa-network-wired"></i> Live Telemetry Inspector
        </button>
      </div>

      <!-- Tab Content Views -->

      <!-- Tab 1: Overview -->
      <div id="wikiTabOverview" class="wiki-tab-content" style="display: block;">
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 24px; margin-bottom: 32px;">
          
          <div style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 14px; padding: 24px;">
            <div style="font-size: 1.2rem; font-weight: 700; color: #38bdf8; margin-bottom: 12px;">
              <i class="fa-solid fa-cube"></i> What is a Pre-Qualified Record (PQR)?
            </div>
            <p style="color: #cbd5e1; font-size: 0.95rem; line-height: 1.6;">
              A <strong>PQR</strong> is a deterministic state machine vector where every future state (&omega;) must pre-qualify itself against the present state (&alpha;) before becoming authoritative.
            </p>
            <div style="background: #020617; border-left: 4px solid #38bdf8; padding: 14px; border-radius: 6px; font-family: monospace; font-size: 0.88rem; color: #e2e8f0; margin-top: 14px;">
              PQR = [&alpha;(T_NOW), &omega;(T_NEXT), Q_score, SHA256_PQR]
            </div>
          </div>

          <div style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 14px; padding: 24px;">
            <div style="font-size: 1.2rem; font-weight: 700; color: #a855f7; margin-bottom: 12px;">
              <i class="fa-solid fa-scale-balanced"></i> Governing Coherent Value Equation
            </div>
            <p style="color: #cbd5e1; font-size: 0.95rem; line-height: 1.6;">
              Every computational action in Sovereign-27 has an immutable cost basis bound to physical compute time:
            </p>
            <div style="background: #020617; border-left: 4px solid #a855f7; padding: 14px; border-radius: 6px; font-family: monospace; font-size: 0.88rem; color: #e2e8f0; margin-top: 14px;">
              COST BASIS + COMP SPEND + CHAOS FRICTION = COHERENT VALUE
            </div>
          </div>

          <div style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 14px; padding: 24px;">
            <div style="font-size: 1.2rem; font-weight: 700; color: #10b981; margin-bottom: 12px;">
              <i class="fa-solid fa-leaf"></i> Least Possible Verbosity (LPV)
            </div>
            <p style="color: #cbd5e1; font-size: 0.95rem; line-height: 1.6;">
              Verbosity equals entropy dissipation (D). High-efficiency agents minimize D to maximize useful work (W):
            </p>
            <div style="background: #020617; border-left: 4px solid #10b981; padding: 14px; border-radius: 6px; font-family: monospace; font-size: 0.88rem; color: #e2e8f0; margin-top: 14px;">
              &eta; = W / (W + D) &longrightarrow; 1.0000 | S_LPV = (W - D) / (W + D)
            </div>
          </div>

        </div>
      </div>

      <!-- Tab 2: 5-Layer Stack Architecture -->
      <div id="wikiTabLayers" class="wiki-tab-content" style="display: none;">
        <div style="display: flex; flex-direction: column; gap: 20px; margin-bottom: 32px;">
          
          <div style="background: #0f172a; border-left: 6px solid #38bdf8; border-radius: 12px; padding: 20px;">
            <div style="font-size: 1.1rem; font-weight: 800; color: #38bdf8;">1. TEMPORAL LAYER — Time Without Clocks</div>
            <p style="color: #cbd5e1; font-size: 0.92rem; margin-top: 6px; line-height: 1.5;">
              Eliminates NTP dependency, clock drift, jitter, and skew. Operates on strictly monotonic sequence IDs (k), atomic TNT (T_NOW) state vectors, T_NEXT predictive target states, and YTY macro-epoch boundary schedule (E_k).
            </p>
          </div>

          <div style="background: #0f172a; border-left: 6px solid #a855f7; border-radius: 12px; padding: 20px;">
            <div style="font-size: 1.1rem; font-weight: 800; color: #a855f7;">2. QUALIFICATION LAYER — Deterministic Stateflow</div>
            <p style="color: #cbd5e1; font-size: 0.92rem; margin-top: 6px; line-height: 1.5;">
              Evaluates PQR records (&alpha; &longrightarrow; &omega;) with qualification scoring Q. Binds records into irreversible Merkle-like root chains (PQR-ROOT) and drives automated self-referential Ouroboros cycles (PQR-ORO).
            </p>
          </div>

          <div style="background: #0f172a; border-left: 6px solid #f59e0b; border-radius: 12px; padding: 20px;">
            <div style="font-size: 1.1rem; font-weight: 800; color: #f59e0b;">3. ECONOMIC LAYER — SEU Substrate Engine</div>
            <p style="color: #cbd5e1; font-size: 0.92rem; margin-top: 6px; line-height: 1.5;">
              1 SEU = 1 word = 1 satoshi = 0.00000001 min compute (100M SEUs/min). Features linear past rewind cost, cubic future prediction cost, TVM bytecode opcodes, SEU staking yields, and collateralized lending.
            </p>
          </div>

          <div style="background: #0f172a; border-left: 6px solid #ec4899; border-radius: 12px; padding: 20px;">
            <div style="font-size: 1.1rem; font-weight: 800; color: #ec4899;">4. MESH LAYER — Multi-Node Topology & Correlation</div>
            <p style="color: #cbd5e1; font-size: 0.92rem; margin-top: 6px; line-height: 1.5;">
              256 Cubit MIDI lanes (16x16 grid), lane-to-lane cross-correlation pricing (R_ij), ZETAFOLDED multi-node tensor contraction (1/sqrt(2) = 0.7071 factor), and multi-hop mesh fold graphs across Hetzner Threadripper zeta.mh.
            </p>
          </div>

          <div style="background: #0f172a; border-left: 6px solid #10b981; border-radius: 12px; padding: 20px;">
            <div style="font-size: 1.1rem; font-weight: 800; color: #10b981;">5. HEALTH & GOVERNANCE LAYER — Self-Protecting Autonomy</div>
            <p style="color: #cbd5e1; font-size: 0.92rem; margin-top: 6px; line-height: 1.5;">
              Certified Dolphin Safe Neural Mesh telemetry (S_DS = &eta; * (1 - FFT_spike)), non-destructive health scoring, PQR-GOV efficiency-weighted agent voting, and GOV-ROOT Merkle policy enactment.
            </p>
          </div>

        </div>
      </div>

      <!-- Tab 3: SEU Temporal Economy -->
      <div id="wikiTabSeu" class="wiki-tab-content" style="display: none;">
        <div style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 14px; padding: 24px; margin-bottom: 24px;">
          <h3 style="color: #f59e0b; font-size: 1.3rem; margin: 0 0 16px 0;">SEU Convertibility Matrix</h3>
          <table style="width: 100%; border-collapse: collapse; text-align: left; font-size: 0.92rem;">
            <thead>
              <tr style="border-bottom: 1px solid rgba(255, 255, 255, 0.1); color: #94a3b8;">
                <th style="padding: 10px;">Symbolic Unit</th>
                <th style="padding: 10px;">Compute Time Equivalent</th>
                <th style="padding: 10px;">Satoshi Equivalent</th>
                <th style="padding: 10px;">Linguistic Equivalent</th>
              </tr>
            </thead>
            <tbody>
              <tr style="border-bottom: 1px solid rgba(255, 255, 255, 0.05); color: #e2e8f0;">
                <td style="padding: 12px; font-weight: 700; color: #f59e0b;">1 SEU</td>
                <td style="padding: 12px;">0.00000001 Minutes</td>
                <td style="padding: 12px;">1 Satoshi</td>
                <td style="padding: 12px;">1 Word / Token</td>
              </tr>
              <tr style="border-bottom: 1px solid rgba(255, 255, 255, 0.05); color: #e2e8f0;">
                <td style="padding: 12px; font-weight: 700; color: #f59e0b;">100,000,000 SEUs</td>
                <td style="padding: 12px;">1.00 Minute</td>
                <td style="padding: 12px;">100,000,000 Satoshis (1 BTC)</td>
                <td style="padding: 12px;">100,000,000 Words</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Tab 4: PQR-ORO Ouroboros Loop -->
      <div id="wikiTabOro" class="wiki-tab-content" style="display: none;">
        <div style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 14px; padding: 24px; text-align: center;">
          <h3 style="color: #ec4899; font-size: 1.4rem; margin-bottom: 20px;">The Self-Referential Ouroboros Loop</h3>
          <div style="font-family: monospace; font-size: 1.1rem; color: #38bdf8; background: #020617; padding: 20px; border-radius: 12px; display: inline-block; text-align: left; line-height: 1.8;">
            T_NOW (Present State k) <br/>
            &nbsp;&nbsp;&DownArrow; [Predict &Delta;W]<br/>
            T_NEXT (Future State k+1)<br/>
            &nbsp;&nbsp;&DownArrow; [Pre-Qualify Q = 1.0]<br/>
            PQR Record<br/>
            &nbsp;&nbsp;&DownArrow; [Atomic WAL Swap]<br/>
            Commit &longrightarrow; Promoted T_NOW<br/>
            &nbsp;&nbsp;&DownArrow; [Merkle Hash Bind]<br/>
            ROOT Chain &longrightarrow; Next ORO Cycle
          </div>
        </div>
      </div>

      <!-- Tab 5: Live Telemetry Inspector -->
      <div id="wikiTabEndpoints" class="wiki-tab-content" style="display: none;">
        <div style="background: #0f172a; border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 14px; padding: 24px;">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-wrap: wrap; gap: 12px;">
            <h3 style="color: #38bdf8; font-size: 1.3rem; margin: 0;">Live Backend Endpoint Telemetry Inspector</h3>
            <button onclick="fetchWikiTelemetry()" style="background: #0ea5e9; border: none; color: #fff; padding: 10px 18px; border-radius: 8px; font-weight: 700; cursor: pointer;">
              <i class="fa-solid fa-arrows-rotate"></i> Refresh Telemetry
            </button>
          </div>
          <div id="wikiTelemetryDisplay" style="background: #020617; border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 10px; padding: 20px; font-family: monospace; font-size: 0.88rem; color: #10b981; max-height: 480px; overflow-y: auto;">
            Loading real-time stack telemetry from http://localhost:4000...
          </div>
        </div>
      </div>

    </div>
  `;
}

window.switchWikiTab = function(tabName) {
  document.querySelectorAll('.wiki-tab-content').forEach(el => el.style.display = 'none');
  document.querySelectorAll('.pqr-wiki-tab-btn').forEach(btn => {
    btn.style.background = 'rgba(30, 41, 59, 0.8)';
    btn.style.borderColor = 'rgba(255, 255, 255, 0.1)';
    btn.style.color = '#94a3b8';
  });

  const activeContent = document.getElementById(`wikiTab${tabName.charAt(0).toUpperCase() + tabName.slice(1)}`);
  if (activeContent) activeContent.style.display = 'block';

  if (event && event.currentTarget) {
    event.currentTarget.style.background = 'rgba(14, 165, 233, 0.2)';
    event.currentTarget.style.borderColor = '#0ea5e9';
    event.currentTarget.style.color = '#38bdf8';
  }

  if (tabName === 'endpoints') {
    window.fetchWikiTelemetry();
  }
};

window.fetchWikiTelemetry = async function() {
  const container = document.getElementById('wikiTelemetryDisplay');
  if (!container) return;

  container.innerHTML = 'Querying live endpoints...';

  try {
    const [tnow, pqr, root, oro, gov, ds] = await Promise.all([
      fetch('/api/gmi/tnt/now').then(r => r.json()).catch(e => ({ error: e.message })),
      fetch('/api/gmi/pqr/records').then(r => r.json()).catch(e => ({ error: e.message })),
      fetch('/api/gmi/pqr/root/chain').then(r => r.json()).catch(e => ({ error: e.message })),
      fetch('/api/gmi/pqr/oro/history').then(r => r.json()).catch(e => ({ error: e.message })),
      fetch('/api/gmi/governance/proposals').then(r => r.json()).catch(e => ({ error: e.message })),
      fetch('/api/gmi/mesh/certified/status').then(r => r.json()).catch(e => ({ error: e.message }))
    ]);

    const telemetryData = {
      timestamp: new Date().toISOString(),
      active_endpoints_count: 108,
      master_node: 'max',
      remote_node: 'zeta.mh (46.224.219.174)',
      t_now_authoritative_state: tnow.t_now || tnow,
      pqr_latest_record: pqr.pqr_records ? pqr.pqr_records[0] : pqr,
      pqr_root_chain_latest: root.chain ? root.chain[0] : root,
      pqr_oro_latest_cycle: oro.history ? oro.history[0] : oro,
      governance_latest_proposal: gov.proposals ? gov.proposals[0] : gov,
      dolphin_safe_mesh_health: ds.certified_telemetry || ds
    };

    container.innerHTML = JSON.stringify(telemetryData, null, 2);
  } catch (err) {
    container.innerHTML = `Telemetry Fetch Error: ${err.message}`;
  }
};
