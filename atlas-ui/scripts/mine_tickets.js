import fs from 'fs';
import path from 'path';
import readline from 'readline';

const BRAIN_DIR = 'C:\\Users\\theal\\.gemini\\antigravity\\brain';
const ZETA_API_URL = 'http://127.0.0.1:4052/api/tickets/propose';

// Track duplicates by a simple title hash to avoid spam
const seenTitles = new Set();
let ticketCount = 0;

const isDryRun = process.argv.includes('--dry-run');

// Filter regexes
const EXCLUDE_PATTERN = /typo|whitespace|padding|margin|import|comment|scratch|temporary|test|ephemeral|diagnostic|print|log|health/i;
const INCLUDE_PATTERN = /architecture|security|vault|zeta|mev|manifest\.json|shared_brain|layout|ui|dns|proxy|sentinel|ticket|keystore|topology|environment|simulation|native|risk/i;

async function processFile(filePath) {
  const fileStream = fs.createReadStream(filePath);
  const rl = readline.createInterface({
    input: fileStream,
    crlfDelay: Infinity
  });

  for await (const line of rl) {
    if (!line.trim()) continue;
    try {
      const step = JSON.parse(line);
      if (step.type === 'PLANNER_RESPONSE' && step.tool_calls) {
        for (const tc of step.tool_calls) {
          if (tc.name === 'replace_file_content' || tc.name === 'multi_replace_file_content' || tc.name === 'write_to_file') {
            const args = tc.args || {};
            let title = args.Description || args.toolSummary || args.toolAction;
            const instruction = args.Instruction || args.toolAction || `Historical execution of ${tc.name}`;
            
            // For strings that might be stringified json inside args, handle quotes
            if (title && title.startsWith('"')) {
              title = JSON.parse(title);
            }
            let instStr = instruction;
            if (instStr && instStr.startsWith('"')) {
              instStr = JSON.parse(instStr);
            }

            if (!title) continue;
            
            // Clean title
            title = title.replace(/\s+/g, ' ').trim();
            if (title.length < 5) continue; // too short
            
            // Applying Filtering Rules
            const combinedText = `${title} ${instStr} ${args.TargetFile || ''}`.toLowerCase();
            if (EXCLUDE_PATTERN.test(combinedText)) {
              // skip trivial things
              continue;
            }
            if (!INCLUDE_PATTERN.test(combinedText)) {
              // skip if it doesn't match our core architecture topics
              continue;
            }

            if (seenTitles.has(title)) continue;
            seenTitles.add(title);
            
            ticketCount++;
            
            const payload = {
              title: title,
              step_what: instStr,
              step_how: `File modified via ${tc.name} tool targeting ${args.TargetFile ? path.basename(args.TargetFile.replace(/"/g, '')) : 'unknown'}`,
              step_why_uncaught: 'Architectural action identified via historical JSONL log mining',
              step_auto_catch_heal: 'Transcribed via mine_tickets.js heuristic filters',
              step_doc_update: 'Historical brain entry ingested into s27-ticket-matrix'
            };

            if (isDryRun) {
              console.log(`[DRY RUN] Would create ticket: ${title}`);
            } else {
              // POST to Zeta Master Compute
              try {
                const response = await fetch(ZETA_API_URL, {
                  method: 'POST',
                  headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify(payload)
                });
                if (!response.ok) {
                  const errText = await response.text();
                  console.error(`Failed to submit ${title}:`, errText);
                } else {
                  console.log(`Successfully ingested: ${title}`);
                }
              } catch (e) {
                console.error(`Fetch error for ${title}:`, e.message);
              }
              // Add a small delay so we don't overwhelm Zeta SQLite
              await new Promise(r => setTimeout(r, 50));
            }
          }
        }
      }
    } catch (e) {
      // Ignore JSON parse errors for incomplete lines
    }
  }
}

async function main() {
  console.log(`Starting Photographic Memory Mining (Dry Run: ${isDryRun})...`);
  
  if (!fs.existsSync(BRAIN_DIR)) {
    console.error(`Brain directory not found: ${BRAIN_DIR}`);
    return;
  }

  const dirs = fs.readdirSync(BRAIN_DIR, { withFileTypes: true })
    .filter(dirent => dirent.isDirectory())
    .map(dirent => dirent.name);

  for (const dir of dirs) {
    const transcriptPath = path.join(BRAIN_DIR, dir, '.system_generated', 'logs', 'transcript.jsonl');
    if (fs.existsSync(transcriptPath)) {
      // console.log(`Mining ${dir}...`);
      await processFile(transcriptPath);
    }
  }

  console.log(`\n✅ Mining Complete! Total Unique Tickets Discovered: ${ticketCount}`);
}

main().catch(console.error);
