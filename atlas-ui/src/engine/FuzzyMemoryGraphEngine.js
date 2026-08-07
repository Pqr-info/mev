/**
 * FuzzyMemoryGraphEngine.js — 49-Position Fuzzy Logic Relevance & Adaptive Recall Engine
 * 
 * Computes fuzzy logic relevance scores (\mu) across all 49 ticket slots in the MemoryGraph:
 * - Recency (R): Time elapsed since last access/update (exponential decay)
 * - Frequency (F): Access count & cross-agent references
 * - Criticality (C): Auditor anomaly flags, dispute count, and L6 commit volatility
 * 
 * Provides Adaptive Recall Formatting (Level 1 Shallow, Level 2 Standard, Level 3 Deep).
 */

export const FuzzyMemoryGraphEngine = {
  weights: {
    recency: 0.35,
    frequency: 0.35,
    criticality: 0.30
  },

  // Calculate exponential recency decay R \in [0, 1]
  calculateRecency(lastAccessedTs, halfLifeHours = 12) {
    if (!lastAccessedTs) return 0.1;
    const elapsedMs = Date.now() - lastAccessedTs;
    const elapsedHours = elapsedMs / (1000 * 60 * 60);
    return Math.max(0.01, Math.pow(0.5, elapsedHours / halfLifeHours));
  },

  // Calculate normalized frequency F \in [0, 1]
  calculateFrequency(accessCount, maxAccessCount = 50) {
    if (!accessCount) return 0.05;
    return Math.min(1.0, accessCount / maxAccessCount);
  },

  // Calculate criticality C \in [0, 1]
  calculateCriticality(auditorFlags = 0, disputes = 0, l6Volatility = 0) {
    const raw = (auditorFlags * 0.4) + (disputes * 0.3) + (l6Volatility * 0.3);
    return Math.min(1.0, Math.max(0.0, raw));
  },

  // Compute overall Fuzzy Score \mu(T_i) \in [0.0, 1.0]
  computeFuzzyScore(ticket) {
    const r = this.calculateRecency(ticket.last_accessed);
    const f = this.calculateFrequency(ticket.access_count || 1);
    const c = this.calculateCriticality(ticket.auditor_flags || 0, ticket.disputes || 0, ticket.l6_volatility || 0);

    const score = (this.weights.recency * r) + (this.weights.frequency * f) + (this.weights.criticality * c);
    return parseFloat(Math.min(1.0, Math.max(0.0, score)).toFixed(3));
  },

  // Determine Adaptive Recall Level
  getRecallLevel(score) {
    if (score >= 0.80) return { level: 3, name: 'Deep Recall', color: '#10b981' };
    if (score >= 0.40) return { level: 2, name: 'Standard Recall', color: '#3b82f6' };
    return { level: 1, name: 'Shallow Recall', color: '#6b7280' };
  },

  // Format Adaptive Recall Context Injection Payload
  formatAdaptivePayload(ticket) {
    const score = ticket.fuzzy_score ?? this.computeFuzzyScore(ticket);
    const recall = this.getRecallLevel(score);

    if (recall.level === 3) {
      // Level 3: Deep Recall
      return {
        ticket_id: ticket.ticket_id,
        slot: ticket.slot_index,
        title: ticket.title,
        fuzzy_score: score,
        recall_level: recall.name,
        payload: ticket.full_payload || ticket.summary || 'Full telemetry available.',
        l6_commit_history: ticket.l6_commits || [],
        raw_telemetry: ticket.telemetry || { status: ticket.status, trust: ticket.trust_score },
        timestamp: ticket.last_accessed
      };
    } else if (recall.level === 2) {
      // Level 2: Standard Recall
      return {
        ticket_id: ticket.ticket_id,
        slot: ticket.slot_index,
        title: ticket.title,
        fuzzy_score: score,
        recall_level: recall.name,
        summary: ticket.summary || ticket.title,
        status: ticket.status,
        primary_params: ticket.params || {},
        timestamp: ticket.last_accessed
      };
    } else {
      // Level 1: Shallow Recall
      return {
        ticket_id: ticket.ticket_id,
        slot: ticket.slot_index,
        fuzzy_score: score,
        recall_level: recall.name,
        status: ticket.status,
        last_seen_offset: ticket.last_accessed ? `${Math.round((Date.now() - ticket.last_accessed) / 60000)}m ago` : 'N/A'
      };
    }
  }
};
