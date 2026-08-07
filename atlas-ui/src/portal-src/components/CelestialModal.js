/**
 * 5D Celestial Harmonic Alignment Minigame Modal
 */

export function renderCelestialModal(targetObject, currentTunedFreq = 432) {
  if (!targetObject) return '';

  const freqDiff = Math.abs(targetObject.targetFreq - currentTunedFreq);
  const accuracyPct = Math.max(0, 100 - (freqDiff * 1.5)).toFixed(0);

  return `
    <div class="modal-overlay active" id="celestialModalOverlay" style="background: rgba(0,0,0,0.88);">
      <div class="modal-sheet" style="border-top-color: ${targetObject.color};">
        <div class="sheet-handle"></div>

        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px;">
          <div style="display: flex; align-items: center; gap: 8px;">
            <div style="width: 36px; height: 36px; border-radius: 50%; background: ${targetObject.color}22; color: ${targetObject.color}; display: flex; align-items: center; justify-content: center; font-size: 1.1rem; border: 1px solid ${targetObject.color};">
              <i class="fa-solid ${targetObject.icon}"></i>
            </div>
            <div>
              <h2 style="font-family: var(--font-display); font-size: 1.1rem; font-weight: 800;">${targetObject.name}</h2>
              <span style="font-size: 0.72rem; color: ${targetObject.color}; font-weight: 700; text-transform: uppercase;">
                ${targetObject.rarity} • ${targetObject.distanceMeters}m away
              </span>
            </div>
          </div>

          <button id="btnCloseCelestialModal" style="background: transparent; border: none; color: var(--text-muted); font-size: 1.2rem; cursor: pointer;">
            <i class="fa-solid fa-xmark"></i>
          </button>
        </div>

        <!-- Resonance Alignment Target Gauge -->
        <div class="glass-card" style="text-align: center; padding: 20px 14px; background: rgba(0,0,0,0.3); border-color: ${targetObject.color}44;">
          <div style="font-size: 0.75rem; color: var(--text-muted); font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 6px;">
            Target Harmonic Resonance
          </div>

          <div style="font-size: 2.2rem; font-weight: 800; font-family: var(--font-display); color: ${targetObject.color}; text-shadow: 0 0 20px ${targetObject.color}aa;">
            ${targetObject.targetFreq} <span style="font-size: 1.1rem;">Hz</span>
          </div>

          <!-- Live Frequency Tuner Readout -->
          <div style="margin-top: 16px; padding: 12px; background: rgba(255,255,255,0.04); border-radius: 14px;">
            <div style="display: flex; justify-content: space-between; font-size: 0.8rem; font-weight: 700; margin-bottom: 6px;">
              <span>Tuned Frequency: <strong style="color: var(--text-main);">${currentTunedFreq} Hz</strong></span>
              <span style="color: ${accuracyPct > 85 ? '#10b981' : '#fb923c'};">${accuracyPct}% Resonance</span>
            </div>

            <input type="range" id="celestialFreqSlider" min="400" max="1000" step="1" value="${currentTunedFreq}" style="width: 100%; accent-color: ${targetObject.color}; cursor: pointer;" />
          </div>

          <!-- Resonance Visual Meter -->
          <div style="height: 6px; background: rgba(255,255,255,0.1); border-radius: 3px; margin-top: 12px; overflow: hidden;">
            <div style="width: ${accuracyPct}%; height: 100%; background: ${targetObject.color}; transition: width 0.1s linear;"></div>
          </div>
        </div>

        <!-- Alignment Reward Info -->
        <div style="display: flex; justify-content: space-between; align-items: center; margin: 14px 0;">
          <div style="font-size: 0.8rem; color: var(--text-muted);">
            Potential Reward: <strong style="color: #06b6d4;">+${targetObject.energy} Quantum Energy</strong>
          </div>
          <span style="font-size: 0.72rem; color: var(--text-muted);">Expires in ${targetObject.expiresIn}</span>
        </div>

        <button id="btnTriggerHarmonicAlign" data-object-id="${targetObject.id}" class="btn-primary" style="background: linear-gradient(135deg, ${targetObject.color} 0%, var(--brand-secondary) 100%); box-shadow: 0 8px 32px ${targetObject.color}55;">
          <i class="fa-solid fa-atom" style="margin-right: 6px;"></i> Align Harmonic Frequency
        </button>

      </div>
    </div>
  `;
}
