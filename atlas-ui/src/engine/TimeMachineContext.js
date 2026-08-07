/**
 * TimeMachineContext — Central temporal mode tracker.
 * 
 * Tracks whether the system is operating in LIVE mode (normal) or 
 * TEMPORAL mode (historical replay / counterfactual analysis).
 * 
 * When in TEMPORAL mode, all side-effecting operations (quarantine, tighten)
 * MUST be replaced with no-op stubs.
 */
export const TimeMachineContext = {
  mode: 'LIVE',              // 'LIVE' | 'TEMPORAL'
  targetTimestamp: null,      // null when LIVE, epoch ms when TEMPORAL
  traversalPath: ['live'],   // provenance chain
  checkpointWindow: 72 * 60 * 60 * 1000, // 72 hours default
  _listeners: new Set(),

  enterTemporal(timestamp) {
    if (!this.validateTimestamp(timestamp)) {
      console.warn('TimeMachine: timestamp outside checkpoint window, clamping.');
      timestamp = Math.max(timestamp, Date.now() - this.checkpointWindow);
    }
    this.mode = 'TEMPORAL';
    this.targetTimestamp = timestamp;
    this.traversalPath = ['historical', timestamp];
    this._notify();
  },

  exitTemporal() {
    this.mode = 'LIVE';
    this.targetTimestamp = null;
    this.traversalPath = ['live'];
    this._notify();
  },

  isTemporalMode() {
    return this.mode === 'TEMPORAL';
  },

  validateTimestamp(ts) {
    const oldest = Date.now() - this.checkpointWindow;
    return ts >= oldest && ts <= Date.now();
  },

  getTraversalMeta() {
    return {
      mode: this.mode,
      targetTimestamp: this.targetTimestamp,
      traversalPath: [...this.traversalPath],
      temporal: this.isTemporalMode()
    };
  },

  subscribe(fn) {
    this._listeners.add(fn);
    return () => this._listeners.delete(fn);
  },

  _notify() {
    this._listeners.forEach(fn => fn(this.getTraversalMeta()));
  },

  async persistSnapshot(contextData) {
    if (this.mode !== 'TEMPORAL' || !this.targetTimestamp) {
      console.warn('TimeMachine: Cannot persist snapshot outside of TEMPORAL mode.');
      return null;
    }
    try {
      const res = await fetch('http://localhost:4052/api/temporal/snapshot/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          timestamp: this.targetTimestamp,
          context_data: contextData
        })
      });
      const data = await res.json();
      if (data.ok) {
        console.log(`[TimeMachine] Snapshot persisted at epoch ${this.targetTimestamp}`);
        return data.snapshot_id;
      }
    } catch (e) {
      console.error('[TimeMachine] Failed to persist snapshot:', e);
    }
    return null;
  },

  async loadSnapshot() {
    try {
      const res = await fetch('http://localhost:4052/api/temporal/snapshot/load');
      const data = await res.json();
      if (data.ok && data.snapshot) {
        console.log(`[TimeMachine] Snapshot loaded from epoch ${data.snapshot.timestamp}`);
        return data.snapshot;
      }
    } catch (e) {
      console.error('[TimeMachine] Failed to load snapshot:', e);
    }
    return null;
  },

  async initiateSelfRepair(targetAgent, rollbackTimestamp = null) {
    if (!targetAgent) return false;
    try {
      const res = await fetch('http://localhost:4052/api/temporal/repair', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: targetAgent,
          rollback_timestamp: rollbackTimestamp
        })
      });
      const data = await res.json();
      if (data.ok) {
        console.log(`[TimeMachine] Temporal Self-Repair initiated successfully for ${targetAgent}.`);
        return true;
      }
    } catch (e) {
      console.error('[TimeMachine] TSRE invocation failed:', e);
    }
    return false;
  }
};
