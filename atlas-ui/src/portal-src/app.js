/**
 * Sovereign-27 Cognitive Portal Application Controller
 * Live Integration to Real Node.js Backend API (http://localhost:4000)
 */

import { whitelabelEngine, PRESET_THEMES } from './config/whitelabel.js';
import { dolphinApi } from './services/api.js';
import { locationEngine } from './services/location.js';
import { celestialEngine } from './services/celestial.js';

import * as realGmi from './services/realGmiClient.js';

import { renderHeader, renderBottomNav } from './components/Navigation.js';
import { renderSovereignPortal } from './components/SovereignPortal.js';
import { renderStoriesCarousel, renderStoryViewerModal } from './components/Stories.js';
import { renderFeedPosts, renderCreatePostModal } from './components/Feed.js';
import { renderRadarView } from './components/Radar.js';
import { renderChatList } from './components/Chat.js';
import { renderGroups } from './components/Groups.js';
import { renderProfile } from './components/Profile.js';
import { renderNotifications } from './components/Notifications.js';
import { renderWhitelabelStudioModal } from './components/WhitelabelStudio.js';
import { renderAdminDashboard } from './components/Admin.js';
import { renderCelestialModal } from './components/CelestialModal.js';
import { renderBodygraphSection } from './components/Bodygraph.js';

class DolphinAppController {
  constructor() {
    this.activeTab = 'portal';
    this.portalSubTab = 'portal';
    this.telemetryData = null;
    this.bleDevices = [];

    this.bodygraphMode = 'single';
    this.selectedPartnerId = 'peer_1';
    this.selectedGroupIds = ['u101', 'peer_1', 'peer_2'];

    this.activeChatId = null;
    this.activeStory = null;
    this.activeCelestialObject = null;
    this.currentTunedFreq = 432;
    this.showStudioModal = false;
    this.showNotificationsModal = false;
    this.showCreatePostModal = false;
    this.proximityRadiusMeters = 5000;

    this.posts = [];
    this.stories = [];
    this.chats = [];
    this.groups = [];
    this.notifications = [];
  }

  async init() {
    whitelabelEngine.applyTheme();
    whitelabelEngine.subscribe(() => this.render());
    locationEngine.subscribe(() => this.render());

    await this.fetchData();
    await this.loadTelemetry();

    this.render();
    this.attachEventListeners();
    this.pingRealBackend();
  }

  async loadTelemetry() {
    try {
      this.telemetryData = await realGmi.fetchTelemetryInspector();
    } catch (e) {
      console.warn('Could not fetch telemetry inspector:', e.message);
    }
  }

  async pingRealBackend() {
    try {
      const health = await realGmi.health();
      const chip = document.getElementById('backendStatusChip');
      if (chip) {
        chip.innerHTML = `<i class="fa-solid fa-circle-check"></i> Sovereign-27 Substrate (${health.mode})`;
        chip.style.background = '#10b98122';
        chip.style.color = '#10b981';
      }
    } catch (e) {
      const chip = document.getElementById('backendStatusChip');
      if (chip) {
        chip.innerHTML = `<i class="fa-solid fa-triangle-exclamation"></i> Stack Offline (${e.message})`;
        chip.style.background = '#ef444422';
        chip.style.color = '#ef4444';
      }
    }
  }

  async fetchData() {
    this.posts = await dolphinApi.getPosts();
    this.stories = await dolphinApi.getStories();
    this.chats = await dolphinApi.getChats();
    this.groups = await dolphinApi.getGroups();
    this.notifications = await dolphinApi.getNotifications();
  }

  render() {
    const appEl = document.getElementById('app');
    if (!appEl) return;

    const config = whitelabelEngine.config;

    appEl.innerHTML = `
      <div class="android-device-frame">
        <div class="android-status-bar">
          <span>9:41</span>
          <div class="android-status-notch">
            <div class="android-camera-lens"></div>
          </div>
          <div style="display: flex; gap: 6px;">
            <i class="fa-solid fa-wifi"></i>
            <i class="fa-solid fa-signal"></i>
            <i class="fa-solid fa-battery-full"></i>
          </div>
        </div>

        ${renderHeader(config, this.activeTab, () => this.openStudio(), () => this.toggleNotifications())}

        <main class="app-content" id="appContent">
          ${this.renderActiveTabContent(config)}
        </main>

        ${renderBottomNav(this.activeTab, config)}

        ${renderCreatePostModal(config)}
        ${renderWhitelabelStudioModal(config)}
        ${renderStoryViewerModal(this.activeStory)}
        ${renderCelestialModal(this.activeCelestialObject, this.currentTunedFreq)}
      </div>
    `;

    this.bindDynamicEvents();
  }

  renderActiveTabContent(config) {
    if (this.showNotificationsModal) {
      return renderNotifications(this.notifications);
    }

    switch (this.activeTab) {
      case 'portal':
        return renderSovereignPortal(this.portalSubTab, this.telemetryData, this.bleDevices);

      case 'radar':
        return renderRadarView(this.proximityRadiusMeters, this.bleDevices);

      case 'design':
        return renderBodygraphSection(this.bodygraphMode, this.selectedPartnerId, this.selectedGroupIds);

      case 'admin':
        return renderAdminDashboard(config);

      case 'feed':
        return `
          ${config.modules.stories ? renderStoriesCarousel(this.stories) : ''}
          <div id="feedPostsContainer">
            ${renderFeedPosts(this.posts)}
          </div>
        `;

      case 'groups':
        return renderGroups(this.groups);

      case 'chat':
        return renderChatList(this.chats, this.activeChatId);

      case 'profile':
        return renderProfile(config.auth.user, config);

      default:
        return renderSovereignPortal(this.portalSubTab, this.telemetryData, this.bleDevices);
    }
  }

  attachEventListeners() {
    document.addEventListener('input', (e) => {
      if (e.target.id === 'radarRadiusSlider') {
        this.proximityRadiusMeters = parseInt(e.target.value, 10);
        this.render();
      }

      if (e.target.id === 'celestialFreqSlider') {
        this.currentTunedFreq = parseInt(e.target.value, 10);
        this.render();
      }
    });

    document.addEventListener('click', async (e) => {
      const navItem = e.target.closest('.nav-item');
      if (navItem) {
        const tab = navItem.dataset.tab;
        if (tab) {
          this.activeTab = tab;
          this.showNotificationsModal = false;
          this.activeChatId = null;
          this.render();
        }
        return;
      }

      // Bluetooth & Wi-Fi Direct Neighbor Scan Trigger
      if (e.target.closest('#btnScanBleWifiNeighbors')) {
        const btn = e.target.closest('#btnScanBleWifiNeighbors');
        btn.innerHTML = `<i class="fa-solid fa-spinner fa-spin"></i> Scanning Nearby...`;
        this.bleDevices = await locationEngine.scanBluetoothWifiNeighbors();
        this.render();
        return;
      }

      // Portal SubTab Buttons
      const portalSubTabBtn = e.target.closest('.portal-subtab-btn');
      if (portalSubTabBtn) {
        this.portalSubTab = portalSubTabBtn.dataset.subtab;
        if (this.portalSubTab === 'radar') {
          this.activeTab = 'radar';
        }
        this.render();
        return;
      }

      // Refresh Telemetry Inspector
      if (e.target.closest('#btnRefreshPortalTelemetry')) {
        await this.loadTelemetry();
        this.render();
        return;
      }

      // Wiki Section Detail Cards
      if (e.target.closest('#btnWikiOverview')) {
        alert('📖 [1. Overview & Philosophy]\nPQR = Pre-Qualified Record. Sovereign-27 is a self-referential, non-destructive, hash-verified temporal logic mesh running across 108 backend REST endpoints.');
        return;
      }
      if (e.target.closest('#btnWikiArchitecture')) {
        alert('🏗️ [2. 5-Layer Stack Architecture]\nLayer 1: GMI API\nLayer 2: NBEP Substrate\nLayer 3: rqlite Consensus\nLayer 4: PQLite WAL Database\nLayer 5: Shared Brain Mesh');
        return;
      }
      if (e.target.closest('#btnWikiTemporalEconomy')) {
        alert('⚡ [3. SEU Temporal Economy]\nSystem Efficiency Units (SEU) quantify delta work performed per cycle hash validation across Hetzner Threadripper compute nodes.');
        return;
      }
      if (e.target.closest('#btnWikiOuroborosLoop')) {
        alert('♾️ [4. PQR-ORO Ouroboros Loop]\nContinuous self-referential cycle verification linking alpha sequence states to omega target states with SHA-256 root chain integrity.');
        return;
      }

      // Direct Message to Radar Peer
      const openDirectChatBtn = e.target.closest('.open-direct-chat-btn');
      if (openDirectChatBtn) {
        const userName = openDirectChatBtn.dataset.userName || 'Marcus Chen';
        const userAvatar = openDirectChatBtn.dataset.userAvatar || 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=250&q=80';

        let existingChat = this.chats.find(c => c.user.name === userName);
        if (!existingChat) {
          existingChat = {
            id: `chat_geo_${Date.now()}`,
            user: { name: userName, avatar: userAvatar, online: true },
            lastMessage: 'Sovereign-27 Proximity Mesh connection established ⚡',
            time: 'Just now',
            unread: 0,
            messages: [
              { id: 'm_init', sender: 'them', text: `Greetings 5D Node! Connected via local proximity radar! (488m away) 📍`, time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }
            ]
          };
          this.chats.unshift(existingChat);
        }

        this.activeTab = 'chat';
        this.activeChatId = existingChat.id;
        this.render();
        return;
      }

      // Brand logo header click -> Reset to Portal
      if (e.target.closest('#headerBrandClick')) {
        this.activeTab = 'portal';
        this.showNotificationsModal = false;
        this.render();
        return;
      }
    });
  }

  bindDynamicEvents() {
    if (this.showStudioModal) {
      const modal = document.getElementById('whitelabelStudioModalOverlay');
      if (modal) modal.classList.add('active');
    }
  }

  openStudio() {
    this.showStudioModal = true;
    this.render();
  }

  closeStudio() {
    this.showStudioModal = false;
    this.render();
  }

  toggleNotifications() {
    this.showNotificationsModal = !this.showNotificationsModal;
    this.render();
  }
}

function bootApp() {
  const app = new DolphinAppController();
  app.init();
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bootApp);
} else {
  bootApp();
}
