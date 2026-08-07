import fs from 'fs';
import path from 'path';

// Watch the agent's current active brain directory
const BRAIN_DIR = 'C:\\Users\\theal\\.gemini\\antigravity\\brain\\f813a299-ac32-48aa-b973-683c584deb7b';
const TARGET_FILES = ['implementation_plan.md', 'walkthrough.md'];
const COPILOT_BRIDGE_URL = 'http://localhost:4050/antigravity/chat';

console.log(`[Copilot Mirror] Initializing stateless NON-BLOCKING ADVISORY mirror.`);
console.log(`[Copilot Mirror] Watching ${BRAIN_DIR}`);

let timeout;

fs.watch(BRAIN_DIR, (eventType, filename) => {
  if (filename && TARGET_FILES.includes(filename)) {
    clearTimeout(timeout);
    timeout = setTimeout(async () => {
      try {
        const filePath = path.join(BRAIN_DIR, filename);
        if (!fs.existsSync(filePath)) return;
        
        const content = fs.readFileSync(filePath, 'utf8');
        console.log(`[Copilot Mirror] Artifact update detected: ${filename}. Forwarding to mesh:chat for advisory review...`);
        
        const res = await fetch(COPILOT_BRIDGE_URL, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            sender: "max",
            message: `[ADVISORY ARTIFACT FORWARD] ${filename}\n\n${content}`,
            timestamp: Date.now()
          })
        });
        
        console.log(`[Copilot Mirror] Advisory payload forwarded. Response: ${res.status}`);

      } catch (err) {
        console.error(`[Copilot Mirror] Bridge Error:`, err.message);
      }
    }, 500);
  }
});

// Keep process alive indefinitely
setInterval(() => {}, 1000 * 60 * 60);
