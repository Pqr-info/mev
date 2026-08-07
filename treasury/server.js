const express = require('express');
const TemporalTreasury = require('./treasury_engine');

const app = express();
app.use(express.json());

const role = process.env.NODE_ROLE || 'secondary';
const treasury = new TemporalTreasury(role);

app.post('/treasury/tick', (req, res) => {
    const srrkMetrics = req.body;
    if (!srrkMetrics || typeof srrkMetrics.sovereignTime === 'undefined') {
        return res.status(400).json({ error: 'Missing srrkMetrics.sovereignTime' });
    }
    
    treasury.tick(srrkMetrics);
    
    res.json({
        status: 'ok',
        treasuryState: {
            clock: treasury.sovereignClock,
            stbl: treasury.stabilityTokens,
            volatilityFutures: treasury.volatilityFutures,
            interestRate: treasury.pnInterestRate
        }
    });
});

app.post('/mesh/sync', (req, res) => {
    const externalTime = req.body.time || Date.now();
    const syncedTime = treasury.syncClock(externalTime);
    res.json({ status: 'synced', time: syncedTime });
});

const port = process.env.TREASURY_PORT || 8082;
app.listen(port, () => {
    console.log(`[Treasury] Engine listening on port ${port} (Role: ${role})`);
});
