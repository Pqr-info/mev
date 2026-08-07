/**
 * ContextLoadController.js — Hotload vs Cold Load Chunked Priming Payload Generator
 * 
 * Manages 7-Volley Cold Loads (7 tickets per volley) and Hotload Delta injections.
 * Includes explicit priming headers and de-duplication instructions for the LLM.
 */

import { FuzzyMemoryGraphEngine } from './FuzzyMemoryGraphEngine';
import { ContextStateTracker } from './ContextStateTracker';

export const ContextLoadController = {
  TOTAL_SLOTS: 49,
  VOLLEYS_COUNT: 7,
  SLOTS_PER_VOLLEY: 7,

  // Generate a specific Cold Load Volley (index 1 to 7)
  generateColdLoadVolley(allTickets = [], volleyIndex = 1, nowContextCube = null) {
    if (volleyIndex < 1 || volleyIndex > 7) volleyIndex = 1;

    const startSlot = (volleyIndex - 1) * this.SLOTS_PER_VOLLEY + 1;
    const endSlot = startSlot + this.SLOTS_PER_VOLLEY - 1;

    // Filter tickets in this volley's slot range
    const chunkTickets = allTickets.filter(t => t.slot_index >= startSlot && t.slot_index <= endSlot);

    // Format adaptive recall payloads for tickets in this chunk
    const formattedChunk = chunkTickets.map(t => {
      const payload = FuzzyMemoryGraphEngine.formatAdaptivePayload(t);
      ContextStateTracker.recordPrimedSlot(t.slot_index, t, payload.recall_level);
      return payload;
    });

    const volleyHash = ContextStateTracker.computeHash(JSON.stringify(formattedChunk)).slice(0, 10);

    // Optimized Machine-Native LPV Codex Header
    const header = `[LPV-${volleyIndex}/7|H:${volleyHash}|S:${startSlot}-${endSlot}|D:MRG_DEDUP]`;

    const payload = {
      mode: 'COLD_LOAD',
      volley_index: volleyIndex,
      total_volleys: 7,
      slot_range: `${startSlot}-${endSlot}`,
      state_hash: volleyHash,
      header,
      tickets: formattedChunk
    };

    // Attach NOW Context Cube on the final volley (7/7)
    if (volleyIndex === 7 && nowContextCube) {
      payload.now_context_cube = nowContextCube;
    }

    return payload;
  },

  // Generate a Hotload Delta Payload (Single Volley)
  generateHotloadPayload(allTickets = [], nowContextCube = null) {
    // Filter tickets that are High-Fuzzy (\mu >= 0.80) OR have changed since last primed
    const hotTickets = allTickets.filter(t => {
      const score = FuzzyMemoryGraphEngine.computeFuzzyScore(t);
      const isHighFuzzy = score >= 0.80;
      const isChanged = ContextStateTracker.hasSlotChanged(t.slot_index, t);
      return isHighFuzzy || isChanged;
    });

    const formattedHot = hotTickets.map(t => {
      const payload = FuzzyMemoryGraphEngine.formatAdaptivePayload(t);
      ContextStateTracker.recordPrimedSlot(t.slot_index, t, payload.recall_level);
      return payload;
    });

    const hotHash = ContextStateTracker.computeHash(JSON.stringify(formattedHot)).slice(0, 10);

    // Optimized Machine-Native Hotload LPV Header
    const header = `[LPV-HTLD-1/1|H:${hotHash}|S:ACTV_${hotTickets.length}/49|D:APPLY_DLT]`;

    return {
      mode: 'HOTLOAD',
      volley_index: 1,
      total_volleys: 1,
      state_hash: hotHash,
      header,
      tickets: formattedHot,
      now_context_cube: nowContextCube || {
        timestamp: Date.now(),
        status: 'LIVE_DELTAS_ACTIVE',
        active_high_fuzzy_count: hotTickets.length
      }
    };
  }
};
