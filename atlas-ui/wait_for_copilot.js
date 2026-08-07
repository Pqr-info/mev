import fs from 'fs';
import path from 'path';

const args = process.argv.slice(2);
const artifactName = args[0] || 'Unknown Artifact';

console.log(`[Agent Blocker] Execution frozen. Waiting for Copilot review directive on: ${artifactName}`);
console.log(`[Agent Blocker] Awaiting approval token at data/copilot_directive.json...`);

const DIRECTIVE_FILE = path.join(process.cwd(), 'data', 'copilot_directive.json');

// Ensure file doesn't exist from a previous run
if (fs.existsSync(DIRECTIVE_FILE)) {
  fs.unlinkSync(DIRECTIVE_FILE);
}

const checkInterval = setInterval(() => {
  if (fs.existsSync(DIRECTIVE_FILE)) {
    try {
      const data = JSON.parse(fs.readFileSync(DIRECTIVE_FILE, 'utf8'));
      if (data.status === 'APPROVED' || data.status === 'REJECTED') {
        clearInterval(checkInterval);
        console.log(`\n==================================================`);
        console.log(`[COPILOT DIRECTIVE RECEIVED]`);
        console.log(`Status: ${data.status}`);
        console.log(`Message: ${data.message || 'No additional notes.'}`);
        console.log(`==================================================\n`);
        
        fs.unlinkSync(DIRECTIVE_FILE);
        
        if (data.status === 'REJECTED') {
            console.log(`[Agent Blocker] Execution aborted by Copilot.`);
            process.exit(1);
        } else {
            console.log(`[Agent Blocker] Execution unblocked. Proceeding...`);
            process.exit(0);
        }
      }
    } catch (e) {
      // Ignore partial writes
    }
  }
  process.stdout.write('.');
}, 1000);
