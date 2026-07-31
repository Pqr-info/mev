/**
 * Dolphin / UNA REST API Client & Autonomous Offline Service Layer
 */

import { whitelabelEngine } from '../config/whitelabel.js';
import { offlineStore } from './offlineStore.js';
import { INITIAL_POSTS, INITIAL_STORIES, INITIAL_CHATS, INITIAL_GROUPS, INITIAL_NOTIFICATIONS } from './mockData.js';

class DolphinApiService {
  constructor() {
    this.posts = [...INITIAL_POSTS];
    this.stories = [...INITIAL_STORIES];
    this.chats = [...INITIAL_CHATS];
    this.groups = [...INITIAL_GROUPS];
    this.notifications = [...INITIAL_NOTIFICATIONS];

    // Load any locally created offline posts
    const localPosts = offlineStore.getLocalPosts();
    if (localPosts.length > 0) {
      this.posts = [...localPosts, ...this.posts];
    }
  }

  get config() {
    return whitelabelEngine.config;
  }

  async callService(module, method, params = {}, options = {}) {
    if (this.config.connectionMode === 'INDEPENDENT' || this.config.connectionMode === 'MOCK') {
      await new Promise(res => setTimeout(res, 200));
      return this.handleMockServiceCall(module, method, params);
    }

    // Live UNA/Dolphin Server REST Call
    const baseUrl = this.config.serverUrl.replace(/\/$/, '');
    const serviceUrl = `${baseUrl}?r=${encodeURIComponent(module)}/${encodeURIComponent(method)}`;
    
    try {
      const headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        ...options.headers
      };

      if (this.config.auth.userToken) {
        headers['Authorization'] = `Bearer ${this.config.auth.userToken}`;
      }

      const response = await fetch(serviceUrl, {
        method: options.method || 'POST',
        headers,
        body: JSON.stringify(params)
      });

      if (!response.ok) {
        throw new Error(`UNA Server returned HTTP status ${response.status}`);
      }

      return await response.json();
    } catch (err) {
      console.warn(`[Dolphin API] Remote server unavailable: ${err.message}. Operating in Autonomous Independent Mode.`, err);
      return this.handleMockServiceCall(module, method, params);
    }
  }

  handleMockServiceCall(module, method, params) {
    switch (`${module}/${method}`) {
      case 'bx_posts/get_posts':
        return { status: 'success', data: this.posts };
      
      case 'bx_posts/entity_add':
        const newPost = {
          id: `p_${Date.now()}`,
          author: {
            id: this.config.auth.user.id,
            name: this.config.auth.user.name,
            username: `@${this.config.auth.user.username}`,
            avatar: this.config.auth.user.avatar,
            verified: true
          },
          time: 'Just now',
          locationTag: '📍 45m away • South Beach',
          content: params.content || '',
          media: params.media || null,
          mediaType: params.media ? 'image' : null,
          stats: { likes: 1, dolphins: 1, comments: 0, shares: 0 },
          userReacted: { like: true, dolphin: true },
          comments: []
        };
        this.posts.unshift(newPost);
        offlineStore.saveLocalPost(newPost);
        return { status: 'success', data: newPost };

      case 'bx_posts/react':
        const post = this.posts.find(p => p.id === params.postId);
        if (post) {
          if (params.type === 'dolphin') {
            post.userReacted.dolphin = !post.userReacted.dolphin;
            post.stats.dolphins += post.userReacted.dolphin ? 1 : -1;
          } else {
            post.userReacted.like = !post.userReacted.like;
            post.stats.likes += post.userReacted.like ? 1 : -1;
          }
        }
        return { status: 'success', data: post };

      case 'bx_comments/add':
        const targetPost = this.posts.find(p => p.id === params.postId);
        if (targetPost) {
          const newComment = {
            id: `c_${Date.now()}`,
            author: this.config.auth.user.name,
            avatar: this.config.auth.user.avatar,
            text: params.text,
            time: 'Just now'
          };
          targetPost.comments.push(newComment);
          targetPost.stats.comments = targetPost.comments.length;
        }
        return { status: 'success', data: targetPost };

      case 'bx_persons/get_stories':
        return { status: 'success', data: this.stories };

      case 'bx_messenger/get_chats':
        return { status: 'success', data: this.chats };

      case 'bx_messenger/send_message':
        const chat = this.chats.find(c => c.id === params.chatId);
        if (chat) {
          const newMsg = {
            id: `m_${Date.now()}`,
            sender: 'me',
            text: params.text,
            time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
          };
          chat.messages.push(newMsg);
          chat.lastMessage = params.text;
          chat.time = 'Just now';
          offlineStore.saveLocalChatMessage(params.chatId, newMsg);
        }
        return { status: 'success', data: chat };

      case 'bx_groups/get_groups':
        return { status: 'success', data: this.groups };

      case 'bx_notifications/get_all':
        return { status: 'success', data: this.notifications };

      default:
        return { status: 'success', data: null };
    }
  }

  async getPosts() {
    const res = await this.callService('bx_posts', 'get_posts');
    return res.data;
  }

  async createPost(content, media = null) {
    const res = await this.callService('bx_posts', 'entity_add', { content, media });
    return res.data;
  }

  async toggleReaction(postId, type = 'dolphin') {
    const res = await this.callService('bx_posts', 'react', { postId, type });
    return res.data;
  }

  async addComment(postId, text) {
    const res = await this.callService('bx_comments', 'add', { postId, text });
    return res.data;
  }

  async getStories() {
    const res = await this.callService('bx_persons', 'get_stories');
    return res.data;
  }

  async getChats() {
    const res = await this.callService('bx_messenger', 'get_chats');
    return res.data;
  }

  async sendMessage(chatId, text) {
    const res = await this.callService('bx_messenger', 'send_message', { chatId, text });
    return res.data;
  }

  async getGroups() {
    const res = await this.callService('bx_groups', 'get_groups');
    return res.data;
  }

  async getNotifications() {
    const res = await this.callService('bx_notifications', 'get_all');
    return res.data;
  }
}

export const dolphinApi = new DolphinApiService();
