/**
 * Dolphin Chat Component - Messaging & Direct Messages
 */

export function renderChatList(chats, activeChatId = null) {
  if (activeChatId) {
    const activeChat = chats.find(c => c.id === activeChatId);
    if (activeChat) {
      return renderActiveChatRoom(activeChat);
    }
  }

  return `
    <div class="chat-section">
      <div class="section-header" style="margin-bottom: 14px;">
        <span class="section-title">
          <i class="fa-solid fa-comments" style="color: var(--brand-primary);"></i> Dolphin Chat
        </span>
        <span style="font-size: 0.75rem; color: var(--text-muted); font-weight: 600;">
          Direct Messages
        </span>
      </div>

      <div class="chat-list">
        ${chats.map(chat => `
          <div class="chat-item open-chat-btn" data-chat-id="${chat.id}">
            <div class="chat-avatar-wrap">
              <img src="${chat.user.avatar}" style="width: 48px; height: 48px; border-radius: 50%; object-fit: cover;" />
              ${chat.user.online ? '<span class="online-indicator"></span>' : ''}
            </div>
            
            <div class="chat-details">
              <div class="chat-user">
                <span>${chat.user.name}</span>
                <span style="font-size: 0.7rem; color: var(--text-muted); font-weight: normal;">${chat.time}</span>
              </div>
              <div class="chat-last-msg">
                ${chat.lastMessage}
              </div>
            </div>

            ${chat.unread > 0 ? `
              <span class="badge-count" style="position: static; font-size: 0.7rem;">${chat.unread}</span>
            ` : ''}
          </div>
        `).join('')}
      </div>
    </div>
  `;
}

function renderActiveChatRoom(chat) {
  return `
    <div class="chat-window">
      <!-- Chat Header -->
      <div style="display: flex; align-items: center; justify-content: space-between; padding-bottom: 12px; border-bottom: 1px solid var(--border-color); margin-bottom: 12px;">
        <div style="display: flex; align-items: center; gap: 10px;">
          <button id="btnBackToChatList" class="icon-btn" style="width: 32px; height: 32px;">
            <i class="fa-solid fa-arrow-left"></i>
          </button>
          <img src="${chat.user.avatar}" style="width: 36px; height: 36px; border-radius: 50%;" />
          <div>
            <div style="font-weight: 700; font-size: 0.9rem;">${chat.user.name}</div>
            <div style="font-size: 0.7rem; color: #10b981; font-weight: 600;">${chat.user.online ? 'Active now' : 'Offline'}</div>
          </div>
        </div>

        <div style="display: flex; gap: 8px;">
          <button class="icon-btn" style="width: 32px; height: 32px;"><i class="fa-solid fa-phone"></i></button>
          <button class="icon-btn" style="width: 32px; height: 32px;"><i class="fa-solid fa-video"></i></button>
        </div>
      </div>

      <!-- Messages History Container -->
      <div style="flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: 4px; padding-right: 4px;" id="chatMessagesContainer">
        ${chat.messages.map(msg => `
          <div class="msg-bubble ${msg.sender === 'me' ? 'sent' : 'received'}">
            <div>${msg.text}</div>
            <div style="font-size: 0.65rem; opacity: 0.7; text-align: right; margin-top: 3px;">${msg.time}</div>
          </div>
        `).join('')}
      </div>

      <!-- Message Input Bar -->
      <div style="display: flex; gap: 8px; margin-top: 12px; padding-top: 8px; border-top: 1px solid var(--border-color);">
        <input type="text" id="inputChatMessage" placeholder="Type a message..." class="form-input" style="border-radius: 24px; padding: 10px 16px;" />
        <button id="btnSendChatMessage" data-chat-id="${chat.id}" class="btn-primary" style="width: 44px; height: 44px; border-radius: 50%; padding: 0; display: flex; align-items: center; justify-content: center;">
          <i class="fa-solid fa-paper-plane"></i>
        </button>
      </div>
    </div>
  `;
}
