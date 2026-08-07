// Dashboard script for advanced V2 UI

function updatePositions() {
    fetch('/positions')
        .then(response => response.json())
        .then(data => {
            // Update blocks/tx rate in V2 UI based on some dummy stats derived from positions
            let total_positions = 0;
            for (const [symbol, qty] of Object.entries(data)) {
                total_positions += Math.abs(qty);
            }
            const blocksEl = document.getElementById('blocksTotal');
            if(blocksEl) blocksEl.textContent = Object.keys(data).length;
            const txRateEl = document.getElementById('txRate');
            if(txRateEl) txRateEl.textContent = total_positions;
        })
        .catch(err => console.error('Error fetching positions:', err));
}

function updateTelemetry() {
    fetch('/telemetry')
        .then(response => response.json())
        .then(data => {
            // Re-use some V2 UI elements for telemetry if possible
            const blockHeightEl = document.getElementById('blockHeight');
            if(blockHeightEl) blockHeightEl.textContent = data.cpu_percent + '% CPU';
            
            const propagationEl = document.getElementById('propagation');
            if(propagationEl) propagationEl.textContent = data.engine_latency_ms + 'ms';
            
            const orderLatencyEl = document.getElementById('orderLatency');
            if(orderLatencyEl && data.last_order_latency_ms !== undefined) {
                orderLatencyEl.textContent = Math.round(data.last_order_latency_ms) + 'ms';
            }
        })
        .catch(err => console.error('Error fetching telemetry:', err));
}

function updateEngineStatus() {
    fetch('/health')
        .then(res => res.json())
        .then(data => {
            const statusText = document.getElementById('statusText');
            const statusPulse = document.getElementById('statusPulse');
            if (data && data.status === 'ok') {
                if(statusText) statusText.textContent = 'SYSTEM ACTIVE';
                if(statusPulse) {
                    statusPulse.classList.remove('error');
                    statusPulse.classList.add('live');
                }
            } else {
                if(statusText) statusText.textContent = 'SYSTEM DOWN';
                if(statusPulse) {
                    statusPulse.classList.remove('live');
                    statusPulse.classList.add('error');
                }
            }
        })
        .catch(() => {
            const statusText = document.getElementById('statusText');
            if(statusText) statusText.textContent = 'OFFLINE';
        });
}

// Clock updates
setInterval(() => {
    const clock = document.getElementById('clock');
    if(clock) {
        const d = new Date();
        clock.textContent = d.toISOString().split('T')[1].split('.')[0] + ' UTC';
    }
}, 1000);

// Initial load
updatePositions();
updateEngineStatus();
updateTelemetry();
// Refresh every 2 seconds
setInterval(() => {
    updatePositions();
    updateEngineStatus();
    updateTelemetry();
}, 2000);
