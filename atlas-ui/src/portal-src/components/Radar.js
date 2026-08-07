/**
 * Dolphin Location Proximity Radar Component
 * Provides real-time radar sweep view of nearby 5D Mesh Peers and Celestial Anomalies.
 */

import { locationEngine } from '../services/location.js';
import { celestialEngine } from '../services/celestial.js';

export function renderRadarView(radiusMeters = 5000, bleDevices = []) {
  const nearbyPeers = locationEngine.getNearbyPeers().filter(p => p.distanceMeters <= radiusMeters);
  const celestials = celestialEngine.getCelestialObjects(radiusMeters);
  const currentLoc = locationEngine.currentLocation;

  return `
    <div class="radar-section">
      <!-- Top Title & Proximity Radius Slider -->
      <div class="glass-card" style="margin-bottom: 14px;">
        <div class="section-header" style="margin-bottom: 10px;">
          <span class="section-title">
            <i class="fa-solid fa-compass" style="color: var(--brand-primary); font-size: 1.1rem;"></i> 5D Mesh Radar
          </span>
          <span class="location-chip">
            <i class="fa-solid fa-location-dot"></i> GPS Active
          </span>
        </div>

        <div style="display: flex; gap: 8px; margin-bottom: 10px;">
          <button id="btnScanBleWifiNeighbors" class="btn-primary" style="flex: 1; padding: 10px; font-size: 0.8rem; font-weight: 800; border-radius: 10px; background: linear-gradient(135deg, #10b981, #0ea5e9); display: flex; align-items: center; justify-content: center; gap: 8px;">
            <i class="fa-brands fa-bluetooth-b"></i> <i class="fa-solid fa-wifi"></i> Scan Neighbors (Bluetooth / Wi-Fi Direct)
          </button>
        </div>

        <div style="margin-top: 10px;">
          <div style="display: flex; justify-content: space-between; font-size: 0.78rem; font-weight: 700; color: var(--text-muted); margin-bottom: 4px;">
            <span>Discovery Radius:</span>
            <strong style="color: var(--brand-primary);">${radiusMeters < 1000 ? radiusMeters + 'm' : (radiusMeters/1000).toFixed(1) + 'km'}</strong>
          </div>
          <input type="range" id="radarRadiusSlider" min="200" max="10000" step="200" value="${radiusMeters}" style="width: 100%; accent-color: var(--brand-primary); cursor: pointer;" />
        </div>
      </div>

      <!-- Discovered BLE / Wi-Fi Direct Peers Banner -->
      ${bleDevices.length > 0 ? `
        <div class="glass-card" style="margin-bottom: 14px; padding: 12px; background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.3);">
          <div style="font-weight: 700; font-size: 0.8rem; color: #34d399; margin-bottom: 6px;">
            <i class="fa-solid fa-signal"></i> Active BLE / Wi-Fi Neighbors (${bleDevices.length}):
          </div>
          <div style="display: flex; flex-direction: column; gap: 6px;">
            ${bleDevices.map(d => `
              <div style="display: flex; justify-content: space-between; font-size: 0.7rem; padding: 4px 8px; background: rgba(0,0,0,0.3); border-radius: 6px;">
                <span><strong>${d.name}</strong> (${d.protocol})</span>
                <span style="color: #10b981; font-weight: 700;">${d.distanceMeters}m | ${d.rssi}dBm</span>
              </div>
            `).join('')}
          </div>
        </div>
      ` : ''}

      <!-- Animated Radar Display Canvas -->
      <div class="glass-card radar-canvas-container" style="text-align: center; padding: 20px 10px; position: relative; overflow: hidden;">
        <div class="radar-sweep-wrapper">
          <div class="radar-ring ring-1"></div>
          <div class="radar-ring ring-2"></div>
          <div class="radar-ring ring-3"></div>
          <div class="radar-crosshair-v"></div>
          <div class="radar-crosshair-h"></div>
          <div class="radar-sweep-beam"></div>

          <!-- Center User Marker -->
          <div class="radar-center-dot" title="You are here (5D Node)">
            <i class="fa-solid fa-dolphin" style="font-size: 0.8rem; color: #fff;"></i>
          </div>

          <!-- Dynamic Mesh Peer Blips on Radar (Green) -->
          ${nearbyPeers.map(peer => {
            const normalizedDist = Math.min((peer.distanceMeters / radiusMeters) * 42, 42);
            const rad = (peer.angle * Math.PI) / 180;
            const topPct = 50 - normalizedDist * Math.cos(rad);
            const leftPct = 50 + normalizedDist * Math.sin(rad);

            return `
              <div class="radar-peer-blip" style="top: ${topPct}%; left: ${leftPct}%;" data-peer-id="${peer.id}">
                <img src="${peer.avatar}" class="blip-avatar" />
                <span class="blip-ping"></span>
              </div>
            `;
          }).join('')}
        </div>
      </div>

      <!-- Nearby Peers List View -->
      <div style="margin-top: 14px;">
        <div style="font-size: 0.85rem; font-weight: 700; color: var(--text-muted); margin-bottom: 8px;">
          Nearby Proximity Mesh Nodes (${nearbyPeers.length})
        </div>
        <div style="display: flex; flex-direction: column; gap: 8px;">
          ${nearbyPeers.map(peer => `
            <div class="glass-card" style="padding: 10px 14px; display: flex; align-items: center; justify-content: space-between;">
              <div style="display: flex; align-items: center; gap: 10px;">
                <img src="${peer.avatar}" style="width: 36px; height: 36px; border-radius: 50%; object-fit: cover;" />
                <div>
                  <div style="font-weight: 700; font-size: 0.85rem;">${peer.name}</div>
                  <div style="font-size: 0.7rem; color: var(--text-muted);">${peer.status}</div>
                </div>
              </div>
              <button class="btn-primary open-direct-chat-btn" data-user-name="${peer.name}" data-user-avatar="${peer.avatar}" style="padding: 6px 10px; font-size: 0.7rem; border-radius: 8px;">
                Message (${peer.distanceMeters}m)
              </button>
            </div>
          `).join('')}
        </div>
      </div>
    </div>
  `;
}
