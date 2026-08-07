/**
 * Android Navigation Component - Top Bar & Bottom Navigation
 */

export function renderHeader(config, activeTab, onOpenStudio, onOpenNotifications) {
  const modeLabel = config.connectionMode === 'INDEPENDENT'
    ? '5D Independent Mesh'
    : (config.connectionMode === 'MOCK' ? 'Demo Mode' : 'Sovereign-27 Cognitive Substrate');
  
  return `
    <header class="app-header">
      <div class="brand-container" id="headerBrandClick">
        <div class="brand-logo" style="background: linear-gradient(135deg, #0ea5e9, #6366f1);">
          <i class="fa-solid fa-atom"></i>
        </div>
        <div>
          <h1 class="brand-name">Sovereign Portal</h1>
          <div class="server-status-chip" id="backendStatusChip">
            <span class="status-dot"></span>
            ${modeLabel}
          </div>
        </div>
      </div>

      <div class="header-actions">
        ${config.modules.whitelabelStudio ? `
          <button class="icon-btn" id="btnOpenStudio" title="Whitelabel Studio">
            <i class="fa-solid fa-wand-magic-sparkles"></i>
          </button>
        ` : ''}
        
        <button class="icon-btn" id="btnOpenNotifications" title="Notifications">
          <i class="fa-regular fa-bell"></i>
          <span class="badge-dot"></span>
        </button>
      </div>
    </header>
  `;
}

export function renderBottomNav(activeTab, config) {
  const tabs = [
    { id: 'portal', label: 'Portal', icon: 'fa-solid fa-cloud-bolt' },
    { id: 'wiki', label: 'Wiki', icon: 'fa-solid fa-book-bookmark' },
    { id: 'control_room', label: 'Control Room', icon: 'fa-solid fa-gauge-high' },
    { id: 'observatory', label: 'Observatory', icon: 'fa-solid fa-binoculars' },
    { id: 'radar', label: 'Radar', icon: 'fa-solid fa-compass' }
  ];




  return `
    <nav class="bottom-nav">
      ${tabs.map(tab => {
        const isActive = activeTab === tab.id;
        return `
          <button class="nav-item ${isActive ? 'active' : ''}" data-tab="${tab.id}">
            <i class="${tab.icon}"></i>
            <span>${tab.label}</span>
          </button>
        `;
      }).join('')}
    </nav>
  `;
}
