export function calculateCreditRating(nodeId, metrics) {
    // PN validity becomes credit rating
    return (metrics.validPnFrames / Math.max(metrics.totalPnFrames, 1)) * 100;
}
