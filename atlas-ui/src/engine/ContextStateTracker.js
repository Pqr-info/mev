/**
 * ContextStateTracker.js — Session Context State & De-Duplication Tracker
 * 
 * Tracks which slots out of the 49-position MemoryGraph have been primed in the LLM's active context.
 * Computes state hashes to allow the LLM to perform de-duplication and delta updates.
 */

export const ContextStateTracker = {
  primedSlots: new Map(), // slotIndex -> { hash, timestamp, recallLevel }
  isWarm: false,
  lastColdLoadTs: null,

  // Check if session context is warm or requires a 7-volley Cold Load
  isContextWarm(maxAgeMinutes = 30) {
    if (!this.isWarm || !this.lastColdLoadTs) return false;
    const elapsedMinutes = (Date.now() - this.lastColdLoadTs) / (1000 * 60);
    return elapsedMinutes < maxAgeMinutes && this.primedSlots.size >= 49;
  },

  // Simple string hash function for de-duplication tracking
  computeHash(text) {
    let hash = 0;
    if (!text || text.length === 0) return 'hash-0';
    for (let i = 0; i < text.length; i++) {
      const char = text.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash |= 0;
    }
    return `sha256-${Math.abs(hash).toString(16)}`;
  },

  // Record a primed slot volley
  recordPrimedSlot(slotIndex, ticketData, recallLevel) {
    const dataStr = JSON.stringify(ticketData);
    const hash = this.computeHash(dataStr);
    this.primedSlots.set(slotIndex, {
      hash,
      timestamp: Date.now(),
      recallLevel
    });

    if (this.primedSlots.size >= 49) {
      this.isWarm = true;
      this.lastColdLoadTs = Date.now();
    }
  },

  // Check if a slot content has changed since last primed
  hasSlotChanged(slotIndex, currentTicketData) {
    const primed = this.primedSlots.get(slotIndex);
    if (!primed) return true;
    const currentHash = this.computeHash(JSON.stringify(currentTicketData));
    return primed.hash !== currentHash;
  },

  // Reset session state (triggers Cold Load next turn)
  resetSession() {
    this.primedSlots.clear();
    this.isWarm = false;
    this.lastColdLoadTs = null;
  }
};
