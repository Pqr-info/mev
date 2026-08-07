import db15Spec from './db15_backplane.json' assert { type: "json" };

/**
 * Phase 18/19 - Proximity Instruments Manager
 * 
 * Manages the clustering and tier-gating of hardware nodes connected via DB-15 
 * or purely virtual channels.
 */

export class ProximityInstrumentCluster {
    constructor(tier) {
        this.tier = tier; // 1, 2, or 3
        this.instruments = [];
        this.backplaneActive = false;
    }

    addInstrument(instrumentId, transportType) {
        if (this.tier === 1 && this.instruments.length >= 0) {
            throw new Error("Tier 1: No global sequencing Proximity Instruments permitted outside of 10-minute trial.");
        }
        if (this.tier === 2 && this.instruments.length >= 2) {
            throw new Error("Tier 2: Maximum of 2 Proximity Instruments permitted.");
        }
        
        // Tier 3 is unlimited and requires DB-15 backplane
        if (this.tier === 3) {
            this.backplaneActive = true;
            console.log(`[Hardware] Routing instrument ${instrumentId} through DB-15 port...`);
        }

        this.instruments.push({
            id: instrumentId,
            transport: transportType,
            status: "online"
        });
    }

    getBackplaneSpec() {
        return this.backplaneActive ? db15Spec : null;
    }
}
