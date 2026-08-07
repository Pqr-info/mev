import express from 'express';

const app = express();
app.use(express.json());

const PORT = 4060;

// Internal state of telemetry
const telemetryState = {
  agents: {},
  l6_anomalies: [],
  drift_events: []
};

// Ingest telemetry from Zeta
app.post('/api/telemetry/ingest', (req, res) => {
  const { type, agent, data, timestamp } = req.body;
  
  if (!telemetryState.agents[agent]) {
    telemetryState.agents[agent] = { trust_history: [], status: 'UNKNOWN', last_tamper: 0 };
  }
  
  if (type === 'TRUST_UPDATE') {
    telemetryState.agents[agent].status = data.status;
    telemetryState.agents[agent].trust_history.push({ score: data.score, ts: timestamp });
  } else if (type === 'L6_TAMPERING') {
    telemetryState.l6_anomalies.push({ agent, height: data.height, ts: timestamp });
    telemetryState.agents[agent].last_tamper = timestamp;
  }
  
  res.json({ ok: true, msg: 'Telemetry ingested' });
});

// The Advisory Loop
setInterval(async () => {
  console.log('[Semantic Brain] Analyzing telemetry patterns...');
  const now = Date.now();
  
  for (const [agent, state] of Object.entries(telemetryState.agents)) {
    // 1. Quarantine Relaxation Recommendation
    // If agent is in QUARANTINE, and hasn't tampered recently (simulated 15s cooldown for tests)
    console.log(`[Semantic Brain] State for ${agent}: status=${state.status}, last_tamper=${state.last_tamper}, now=${now}`);
    if (state.status === 'QUARANTINE') {
      const timeSinceTamper = now - state.last_tamper;
      console.log(`[Semantic Brain] timeSinceTamper for ${agent}: ${timeSinceTamper}`);
      if (timeSinceTamper >= 25000 && state.status !== 'PENDING_RELAXATION') {
        console.log(`[Semantic Brain] 💡 Proposing Quarantine Relaxation for ${agent}`);
        const proposal = {
          proposal_id: `prop-${now}-relax`,
          type: 'RECOMMEND_QUARANTINE_RELAXATION',
          target_agent: agent,
          risk_level: 'LOW',
          reasoning: 'Agent has served initial quarantine period.',
          suggested_policy: {
            disabledTiers: [],
            quarantinedAgents: []
          }
        };
        
        try {
          await fetch('http://localhost:4052/api/governance/propose', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(proposal)
          });
          // Reset to avoid spamming
          state.status = 'PENDING_RELAXATION';
        } catch (e) {
          console.warn('[Semantic Brain] Failed to submit proposal to Zeta:', e.message);
        }
      }
    }
  }
}, 5000); // Run every 5s for responsiveness in tests

// Predictive Governance Loop
setInterval(async () => {
  console.log('[Semantic Brain] Running predictive governance model...');
  const now = Date.now();
  
  for (const [agent, state] of Object.entries(telemetryState.agents)) {
    let confidence = 0.0;
    let reasoning = [];
    
    // 1. TAMPERING_RISK
    const timeSinceTamper = now - state.last_tamper;
    if (state.last_tamper > 0 && timeSinceTamper < 60000) {
      confidence += 0.4;
      reasoning.push('recent tampering');
    }
    
    // 2. Trust Volatility
    if (state.trust_history.length >= 2) {
      const recent = state.trust_history.slice(-3);
      const drops = recent.filter((h, i) => i > 0 && h.score < recent[i-1].score);
      if (drops.length > 0) {
        confidence += 0.2;
        reasoning.push('trust volatility');
      }
    }
    
    // 3. Status penalty
    if (state.status === 'QUARANTINE') {
      confidence += 0.2;
      reasoning.push('quarantine isolation');
    } else if (state.status === 'PROBATION') {
      confidence += 0.1;
      reasoning.push('probation instability');
    }
    
    if (confidence > 1.0) confidence = 1.0;
    
    // Throttle duplicate high-confidence spam by adding a last_forecast timestamp
    if (confidence >= 0.5 && (!state.last_forecast || (now - state.last_forecast > 15000))) {
      const type = state.status === 'QUARANTINE' ? 'TRUST_COLLAPSE_RISK' : 'TAMPERING_RISK';
      const reasonStr = `Agent exhibiting ${reasoning.join(' and ')}`;
      console.log(`[Semantic Brain] 🔮 FORECAST GENERATED: ${agent} -> ${type} (${confidence.toFixed(2)})`);
      
      const forecast = {
        agent_id: agent,
        type: type,
        confidence: confidence,
        window_ms: 30000,
        reasoning: reasonStr
      };
      
      try {
        await fetch('http://localhost:4052/api/governance/forecast', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(forecast)
        });
        state.last_forecast = now;
      } catch (e) {
        console.warn('[Semantic Brain] Failed to submit forecast to Zeta:', e.message);
      }
    }
  }
}, 8000);

app.listen(PORT, () => {
  console.log(`[Semantic Brain] Active and listening for telemetry on port ${PORT}`);
});
