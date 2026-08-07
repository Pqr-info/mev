/**
 * Dolphin Autonomous Serverless Offline Store
 * Manages local storage persistence for posts, stories, chats, and groups.
 */

import { SEED_POSTS, SEED_STORIES, SEED_CHATS, SEED_GROUPS, SEED_NOTIFICATIONS } from './mockData.js';

const STORAGE_KEYS = {
  POSTS: 'dolphin_offline_posts',
  STORIES: 'dolphin_offline_stories',
  CHATS: 'dolphin_offline_chats',
  GROUPS: 'dolphin_offline_groups',
  NOTIFICATIONS: 'dolphin_offline_notifications'
};

class OfflineStore {
  constructor() {
    this.initStore();
  }

  initStore() {
    if (typeof localStorage === 'undefined') return;

    if (!localStorage.getItem(STORAGE_KEYS.POSTS)) {
      localStorage.setItem(STORAGE_KEYS.POSTS, JSON.stringify(SEED_POSTS));
    }
    if (!localStorage.getItem(STORAGE_KEYS.STORIES)) {
      localStorage.setItem(STORAGE_KEYS.STORIES, JSON.stringify(SEED_STORIES));
    }
    if (!localStorage.getItem(STORAGE_KEYS.CHATS)) {
      localStorage.setItem(STORAGE_KEYS.CHATS, JSON.stringify(SEED_CHATS));
    }
    if (!localStorage.getItem(STORAGE_KEYS.GROUPS)) {
      localStorage.setItem(STORAGE_KEYS.GROUPS, JSON.stringify(SEED_GROUPS));
    }
    if (!localStorage.getItem(STORAGE_KEYS.NOTIFICATIONS)) {
      localStorage.setItem(STORAGE_KEYS.NOTIFICATIONS, JSON.stringify(SEED_NOTIFICATIONS));
    }
  }

  getItem(key, fallback) {
    if (typeof localStorage === 'undefined') return fallback;
    const data = localStorage.getItem(key);
    try {
      return data ? JSON.parse(data) : fallback;
    } catch (e) {
      return fallback;
    }
  }

  setItem(key, value) {
    if (typeof localStorage === 'undefined') return;
    localStorage.setItem(key, JSON.stringify(value));
  }

  getPosts() { return this.getItem(STORAGE_KEYS.POSTS, SEED_POSTS); }
  savePosts(posts) { this.setItem(STORAGE_KEYS.POSTS, posts); }

  getLocalPosts() { return this.getPosts(); }
  saveLocalPosts(posts) { this.savePosts(posts); }

  getStories() { return this.getItem(STORAGE_KEYS.STORIES, SEED_STORIES); }
  getLocalStories() { return this.getStories(); }

  getChats() { return this.getItem(STORAGE_KEYS.CHATS, SEED_CHATS); }
  saveChats(chats) { this.setItem(STORAGE_KEYS.CHATS, chats); }
  getLocalChats() { return this.getChats(); }
  saveLocalChats(chats) { this.saveChats(chats); }

  getGroups() { return this.getItem(STORAGE_KEYS.GROUPS, SEED_GROUPS); }
  getLocalGroups() { return this.getGroups(); }

  getNotifications() { return this.getItem(STORAGE_KEYS.NOTIFICATIONS, SEED_NOTIFICATIONS); }
  getLocalNotifications() { return this.getNotifications(); }
}

export const offlineStore = new OfflineStore();
