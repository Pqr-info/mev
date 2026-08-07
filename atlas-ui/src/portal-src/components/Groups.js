/**
 * Dolphin Communities / Groups Component
 */

export function renderGroups(groups) {
  return `
    <div class="groups-section">
      <div class="section-header" style="margin-bottom: 14px;">
        <span class="section-title">
          <i class="fa-solid fa-layer-group" style="color: var(--brand-primary);"></i> Community Pods
        </span>
        <button class="action-btn" style="color: var(--brand-primary);">
          <i class="fa-solid fa-plus"></i> Create Pod
        </button>
      </div>

      <div class="groups-list">
        ${groups.map(group => `
          <div class="glass-card" style="padding: 0; overflow: hidden;">
            <div style="height: 100px; position: relative;">
              <img src="${group.cover}" style="width: 100%; height: 100%; object-fit: cover;" />
              <span style="position: absolute; top: 10px; right: 10px; background: rgba(0,0,0,0.6); backdrop-filter: blur(4px); font-size: 0.68rem; font-weight: 700; color: #fff; padding: 4px 8px; border-radius: 12px;">
                ${group.category}
              </span>
            </div>

            <div style="padding: 14px;">
              <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 6px;">
                <div>
                  <h3 style="font-weight: 800; font-size: 0.98rem; font-family: var(--font-display);">${group.name}</h3>
                  <div style="font-size: 0.75rem; color: var(--text-muted); margin-top: 2px;">
                    <i class="fa-solid fa-users" style="margin-right: 4px; color: var(--brand-primary);"></i>
                    ${group.membersCount.toLocaleString()} Members
                  </div>
                </div>

                <button class="btn-primary btn-toggle-group" data-group-id="${group.id}" style="width: auto; padding: 6px 14px; font-size: 0.78rem; ${group.isJoined ? 'background: rgba(255,255,255,0.1); box-shadow: none;' : ''}">
                  ${group.isJoined ? 'Joined ✓' : 'Join Pod'}
                </button>
              </div>

              <p style="font-size: 0.82rem; color: var(--text-muted); line-height: 1.4; margin-top: 8px;">
                ${group.description}
              </p>
            </div>
          </div>
        `).join('')}
      </div>
    </div>
  `;
}
