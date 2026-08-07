/**
 * TimeMachineMiddleware — Wraps state-fetching with temporal awareness.
 * 
 * Provides validated temporal state queries and cross-epoch diffing.
 */
import { TimeMachineContext } from './TimeMachineContext';

const API_BASE = '/api/governance';

export const TimeMachineMiddleware = {

  /**
   * Fetch the full mesh state at a given timestamp.
   * Injects traversalPath provenance into the returned state.
   */
  async fetchTemporalState(timestamp) {
    if (!TimeMachineContext.validateTimestamp(timestamp)) {
      return { error: 'Timestamp outside checkpoint window', agents: [], proposals: [], forecasts: [] };
    }

    try {
      const res = await fetch(`${API_BASE}/state-at/${timestamp}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const state = await res.json();

      // Inject provenance metadata
      state.meta = {
        traversalPath: ['historical', timestamp],
        temporal: true,
        fetchedAt: Date.now(),
        targetTimestamp: timestamp
      };

      return state;
    } catch (e) {
      console.error('TimeMachineMiddleware: failed to fetch temporal state', e);
      return { error: e.message, agents: [], proposals: [], forecasts: [] };
    }
  },

  /**
   * Fetch the checkpoint window bounds from the backend.
   */
  async fetchCheckpointWindow() {
    try {
      const res = await fetch(`${API_BASE}/checkpoint-window`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      // Fallback: use local 72h window
      return {
        oldest: Date.now() - TimeMachineContext.checkpointWindow,
        newest: Date.now()
      };
    }
  },

  /**
   * Diff mesh state between two timestamps.
   * Returns structured delta of trust scores, statuses, and rule changes.
   */
  async diffStates(ts1, ts2) {
    try {
      const res = await fetch(`${API_BASE}/state-diff/${ts1}/${ts2}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      // Client-side fallback: fetch both states and diff locally
      const [stateA, stateB] = await Promise.all([
        this.fetchTemporalState(ts1),
        this.fetchTemporalState(ts2)
      ]);
      return this.computeLocalDiff(stateA, stateB);
    }
  },

  /**
   * Client-side structural diff when backend endpoint is unavailable.
   */
  computeLocalDiff(stateA, stateB) {
    const diff = {
      ts1: stateA.meta?.targetTimestamp || null,
      ts2: stateB.meta?.targetTimestamp || null,
      agentDeltas: [],
      summary: ''
    };

    const agentsA = {};
    (stateA.agents || []).forEach(a => { agentsA[a.agent_id] = a; });
    const agentsB = {};
    (stateB.agents || []).forEach(a => { agentsB[a.agent_id] = a; });

    const allIds = new Set([...Object.keys(agentsA), ...Object.keys(agentsB)]);

    for (const id of allIds) {
      const a = agentsA[id] || { trust_score: 100, status: 'UNKNOWN' };
      const b = agentsB[id] || { trust_score: 100, status: 'UNKNOWN' };

      if (a.trust_score !== b.trust_score || a.status !== b.status) {
        diff.agentDeltas.push({
          agent_id: id,
          trustBefore: a.trust_score,
          trustAfter: b.trust_score,
          trustDelta: b.trust_score - a.trust_score,
          statusBefore: a.status,
          statusAfter: b.status
        });
      }
    }

    diff.summary = `${diff.agentDeltas.length} agent(s) changed between epochs.`;
    return diff;
  }
};
