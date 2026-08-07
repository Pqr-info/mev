/**
 * Dolphin Notifications Component
 */

export function renderNotifications(notifications) {
  return `
    <div class="notifications-section">
      <div class="section-header" style="margin-bottom: 14px;">
        <span class="section-title">
          <i class="fa-solid fa-bell" style="color: var(--brand-primary);"></i> Activity Center
        </span>
        <button class="action-btn" id="btnMarkAllRead" style="font-size: 0.75rem; color: var(--brand-primary);">
          Mark all read
        </button>
      </div>

      <div class="notifications-list">
        ${notifications.map(item => `
          <div class="glass-card" style="padding: 12px; margin-bottom: 8px; display: flex; align-items: center; gap: 12px; ${!item.read ? 'border-left: 3px solid var(--brand-primary);' : ''}">
            <img src="${item.avatar}" style="width: 40px; height: 40px; border-radius: 50%; object-fit: cover;" />
            <div style="flex: 1; font-size: 0.82rem;">
              <div>
                <strong>${item.user}</strong> ${item.text}
              </div>
              <div style="font-size: 0.7rem; color: var(--text-muted); margin-top: 2px;">
                ${item.time}
              </div>
            </div>
            ${item.type === 'dolphin' ? '<i class="fa-solid fa-dolphin" style="color: #06b6d4; font-size: 1.1rem;"></i>' : ''}
          </div>
        `).join('')}
      </div>
    </div>
  `;
}
