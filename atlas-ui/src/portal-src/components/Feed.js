/**
 * Dolphin Social Feed & Post Component
 */

export function renderFeedPosts(posts) {
  if (!posts || posts.length === 0) {
    return `
      <div class="glass-card" style="text-align: center; padding: 30px;">
        <i class="fa-solid fa-water" style="font-size: 2.5rem; color: var(--brand-primary); margin-bottom: 10px;"></i>
        <h3 style="font-weight: 700;">No posts in proximity feed</h3>
        <p style="color: var(--text-muted); font-size: 0.85rem; margin-top: 4px;">Be the first to splash a local post in your pod!</p>
      </div>
    `;
  }

  return posts.map(post => {
    const isDolphinReacted = post.userReacted?.dolphin;
    const isLikeReacted = post.userReacted?.like;

    return `
      <article class="glass-card post-card" data-post-id="${post.id}">
        <!-- Post Header -->
        <header class="post-header">
          <div class="post-author">
            <img src="${post.author.avatar}" class="author-avatar" alt="${post.author.name}" />
            <div class="author-info">
              <div class="author-name">
                ${post.author.name}
                ${post.author.verified ? '<i class="fa-solid fa-circle-check badge-verified"></i>' : ''}
              </div>
              <span class="post-time">
                ${post.author.username} • ${post.time}
              </span>
            </div>
          </div>

          <span class="location-chip" style="font-size: 0.68rem; padding: 2px 8px;">
            <i class="fa-solid fa-location-dot"></i> ${post.locationTag || '📍 150m away'}
          </span>
        </header>

        <!-- Post Body Content -->
        <div class="post-body">
          ${post.content.replace(/\n/g, '<br/>')}
        </div>

        <!-- Post Media Attachment -->
        ${post.media ? `
          <div class="post-media">
            <img src="${post.media}" alt="Post media" loading="lazy" />
          </div>
        ` : ''}

        <!-- Post Reaction Actions Bar -->
        <div class="post-actions">
          <div class="reaction-group">
            <button class="action-btn btn-dolphin ${isDolphinReacted ? 'active-react' : ''}" data-action="react-dolphin" data-post-id="${post.id}">
              <i class="fa-solid fa-dolphin"></i>
              <span>${post.stats.dolphins}</span>
            </button>

            <button class="action-btn ${isLikeReacted ? 'active-react' : ''}" data-action="react-like" data-post-id="${post.id}">
              <i class="fa-regular fa-heart"></i>
              <span>${post.stats.likes}</span>
            </button>

            <button class="action-btn btn-toggle-comments" data-post-id="${post.id}">
              <i class="fa-regular fa-comment"></i>
              <span>${post.stats.comments}</span>
            </button>
          </div>

          <button class="action-btn btn-share-post" data-post-id="${post.id}">
            <i class="fa-regular fa-paper-plane"></i>
          </button>
        </div>

        <!-- Nested Interactive Comment Section -->
        <div class="comments-section" id="comments_${post.id}">
          ${(post.comments || []).map(comment => `
            <div class="comment-item">
              <img src="${comment.avatar}" class="comment-avatar" alt="${comment.author}" />
              <div class="comment-bubble">
                <div class="comment-author">${comment.author} <span style="font-weight: normal; color: var(--text-muted); font-size: 0.7rem;">${comment.time}</span></div>
                <div>${comment.text}</div>
              </div>
            </div>
          `).join('')}

          <div class="comment-input-box">
            <input type="text" placeholder="Write a comment..." id="inputComment_${post.id}" />
            <button class="btn-primary btn-submit-comment" data-post-id="${post.id}" style="width: auto; padding: 6px 14px; font-size: 0.8rem;">
              Post
            </button>
          </div>
        </div>
      </article>
    `;
  }).join('');
}

export function renderCreatePostModal(config) {
  return `
    <div class="modal-overlay" id="createPostModalOverlay">
      <div class="modal-sheet">
        <div class="sheet-handle"></div>
        
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
          <h2 style="font-family: var(--font-display); font-size: 1.1rem; font-weight: 800;">Create Location Post</h2>
          <button id="btnCloseCreateModal" style="background: transparent; border: none; color: var(--text-muted); font-size: 1.2rem; cursor: pointer;">
            <i class="fa-solid fa-xmark"></i>
          </button>
        </div>

        <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 14px;">
          <img src="${config.auth.user.avatar}" style="width: 38px; height: 38px; border-radius: 50%;" />
          <div>
            <div style="font-weight: 700; font-size: 0.9rem;">${config.auth.user.name}</div>
            <span class="location-chip" style="font-size: 0.65rem; margin-top: 2px;">
              <i class="fa-solid fa-location-dot"></i> Coastal Mesh • South Beach
            </span>
          </div>
        </div>

        <div class="form-group">
          <textarea id="postContentInput" class="form-input" rows="4" placeholder="What's happening in your local pod?" style="resize: none;"></textarea>
        </div>

        <div class="form-group">
          <label class="form-label">Image URL (Optional)</label>
          <input type="url" id="postMediaInput" class="form-input" placeholder="https://images.unsplash.com/..." />
        </div>

        <div style="display: flex; gap: 10px; margin-top: 20px;">
          <button id="btnPublishPost" class="btn-primary">
            <i class="fa-solid fa-paper-plane" style="margin-right: 6px;"></i> Broadcast to Proximity Mesh
          </button>
        </div>
      </div>
    </div>
  `;
}
