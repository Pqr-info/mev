/**
 * Dolphin Pods / Stories Component
 */

export function renderStoriesCarousel(stories) {
  return `
    <div class="stories-section">
      <div class="section-header">
        <span class="section-title">
          <i class="fa-solid fa-water" style="color: var(--brand-primary);"></i> Dolphin Pods
        </span>
      </div>

      <div class="stories-scroll">
        ${stories.map(story => {
          if (story.isAdd) {
            return `
              <div class="story-item" id="btnAddStory">
                <div class="story-ring add-story-ring">
                  <i class="fa-solid fa-plus"></i>
                </div>
                <span class="story-name">Add Story</span>
              </div>
            `;
          }

          return `
            <div class="story-item view-story-btn" data-story-id="${story.id}">
              <div class="story-ring ${story.unseen ? 'unseen' : 'seen'}">
                <img src="${story.avatar}" class="story-avatar" alt="${story.name}" />
              </div>
              <span class="story-name">${story.name.split(' ')[0]}</span>
            </div>
          `;
        }).join('')}
      </div>
    </div>
  `;
}

export function renderStoryViewerModal(story) {
  if (!story || !story.slides || story.slides.length === 0) return '';
  const currentSlide = story.slides[0];

  return `
    <div class="modal-overlay active" id="storyModalOverlay" style="background: rgba(0,0,0,0.92);">
      <div style="position: relative; width: 100%; height: 100%; display: flex; flex-direction: column; justify-content: space-between; padding: 20px 16px;">
        
        <!-- Progress Bars -->
        <div style="display: flex; gap: 4px; z-index: 10;">
          <div style="flex: 1; height: 3px; background: rgba(255,255,255,0.4); border-radius: 2px; overflow: hidden;">
            <div style="width: 100%; height: 100%; background: #fff; animation: storyProgress 5s linear forwards;"></div>
          </div>
        </div>

        <!-- Story Author Top Bar -->
        <div style="display: flex; justify-content: space-between; align-items: center; z-index: 10; margin-top: 10px;">
          <div style="display: flex; align-items: center; gap: 10px;">
            <img src="${story.avatar}" style="width: 38px; height: 38px; border-radius: 50%; border: 2px solid var(--brand-primary);" />
            <div>
              <div style="font-weight: 700; color: #fff; font-size: 0.9rem;">${story.name}</div>
              <div style="font-size: 0.72rem; color: rgba(255,255,255,0.7);">Just now</div>
            </div>
          </div>
          <button id="btnCloseStoryModal" style="background: rgba(255,255,255,0.2); border: none; width: 34px; height: 34px; border-radius: 50%; color: #fff; cursor: pointer;">
            <i class="fa-solid fa-xmark"></i>
          </button>
        </div>

        <!-- Story Media Content -->
        <div style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: center;">
          <img src="${currentSlide.url}" style="width: 100%; height: 100%; object-fit: cover;" />
          <div style="position: absolute; bottom: 80px; left: 20px; right: 20px; background: rgba(0,0,0,0.6); backdrop-filter: blur(8px); padding: 12px 16px; border-radius: 16px; color: #fff; font-size: 0.9rem; border: 1px solid rgba(255,255,255,0.1);">
            ${currentSlide.caption}
          </div>
        </div>

        <!-- Bottom Reaction Quick Action -->
        <div style="display: flex; gap: 10px; z-index: 10; align-items: center;">
          <input type="text" placeholder="Send reply to ${story.name}..." style="flex: 1; background: rgba(255,255,255,0.2); border: 1px solid rgba(255,255,255,0.3); padding: 10px 16px; border-radius: 25px; color: #fff; outline: none; font-size: 0.85rem;" />
          <button style="background: var(--brand-gradient); border: none; width: 42px; height: 42px; border-radius: 50%; color: #fff; cursor: pointer;">
            <i class="fa-solid fa-paper-plane"></i>
          </button>
        </div>

      </div>
    </div>
    <style>
      @keyframes storyProgress {
        from { width: 0%; }
        to { width: 100%; }
      }
    </style>
  `;
}
