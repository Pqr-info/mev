/**
 * Phase 24 - Temporal Commodities
 * 
 * Defines the tradable timing assets in the SRRK Exchange.
 * Every measurable timing property becomes a tradable asset.
 */

// A. Stability Tokens (STBL)
// Derived from jitter floor, latency floor, PN validity, ensemble consistency.
// Stable nodes mint more STBL. Unstable nodes mint less.
export function mintSTBL(metrics) {
    const jitter = Math.max(metrics.jitter || 0.01, 0.001); // avoid div zero
    const volatility = Math.max(metrics.volatility || 0.01, 0.001);
    return 1.0 / (jitter + volatility);
}

// B. Volatility Futures (VOL-F)
// Predict future instability (BLE interference, UDP congestion).
export function priceVolatilityFuture(expectedVolatility) {
    return expectedVolatility * 100.0;
}

// C. PN Gate Contracts (PN-C)
// Nodes buy/sell PN gate periods. Shorter PN periods = more expensive.
export function pricePnContract(periodMs) {
    const period = Math.max(periodMs, 1);
    return 1000.0 / period;
}

// D. Routing Derivatives (RTE-D)
// Nodes trade routing preferences. High-reliability paths cost more.
export function priceRouteDerivative(pathScore) {
    const consistency = pathScore.consistency || 0;
    const reliability = pathScore.reliability || 0;
    return consistency * reliability;
}

// E. Preset Atomicity Bonds (ATMC-B)
// Preset stability becomes collateral. Atomic preset banks = high-value bonds.
export function priceAtomicityBond(fragmentationIndex) {
    const fragmentation = Math.max(fragmentationIndex, 0.01);
    return 10.0 / fragmentation;
}
