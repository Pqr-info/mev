import puppeteer from 'puppeteer';

(async () => {
  console.log('[QA] Launching Puppeteer...');
  const browser = await puppeteer.launch({
    headless: "new",
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  const errors = [];

  // Capture console errors
  page.on('console', msg => {
    const text = msg.text();
    if (text.includes('ERR_NAME_NOT_RESOLVED')) return;
    if (text.includes('404 (Not Found)')) return; // Handled by response event
    if (msg.type() === 'error') {
      console.log(`[QA][Browser Error] ${text}`);
      errors.push(text);
    } else {
      console.log(`[QA][Browser Log] ${text}`);
    }
  });

  page.on('response', response => {
    if (response.status() === 404) {
      console.log(`[QA][Browser Error] 404 Not Found: ${response.url()}`);
      errors.push(`404 Not Found: ${response.url()}`);
    }
  });
  
  page.on('pageerror', err => {
    console.log(`[QA][Page Error] ${err.toString()}`);
    errors.push(err.toString());
  });

  page.on('requestfailed', request => {
    if (request.failure().errorText !== 'net::ERR_NAME_NOT_RESOLVED') {
      console.log(`[QA][Browser Error] Failed to load resource: ${request.url()} - ${request.failure().errorText}`);
      errors.push(`Failed to load resource: ${request.url()}`);
    }
  });

  try {
    console.log('[QA] Navigating to http://localhost:9080');
    await page.goto('http://localhost:9080', { waitUntil: 'domcontentloaded', timeout: 30000 });
    
    // Switch to Portal tab
    console.log('[QA] Clicking Portal tab...');
    await page.evaluate(() => {
      const tabs = Array.from(document.querySelectorAll('button'));
      const portalTab = tabs.find(t => t.textContent.includes('Portal'));
      if (portalTab) portalTab.click();
    });

    // Wait for iframe
    console.log('[QA] Waiting for iframe to load...');
    await page.waitForSelector('iframe[title="Portal View"]', { timeout: 10000 });

    // We must wait for the iframe to load its content
    await new Promise(r => setTimeout(r, 3000));
    
    const iframeElement = await page.$('iframe[title="Portal View"]');
    const frame = await iframeElement.contentFrame();
    
    if (!frame) {
      throw new Error("Could not find content frame for the Portal View iframe.");
    }

    console.log('[QA] Clicking "Scan Bluetooth/WiFi Neighbors" inside Portal...');
    const scanBtn = await frame.$('#btnScanBleWifiNeighbors');
    if (scanBtn) {
      await scanBtn.click();
      await new Promise(r => setTimeout(r, 2000));
      console.log('[QA] Scan triggered successfully.');
    } else {
      console.log('[QA] Scan button not found, is the portal rendered?');
      const html = await frame.evaluate(() => document.body.innerHTML);
      console.log('[QA] Iframe inner HTML length:', html.length);
      throw new Error("Scan button not found.");
    }
    
    if (errors.length > 0) {
      console.error(`[QA] Failed: Found ${errors.length} browser errors.`);
      process.exit(1);
    } else {
      console.log('[QA] Success: No errors detected.');
      process.exit(0);
    }

  } catch (error) {
    console.error('[QA] Exception during test execution:', error);
    process.exit(1);
  } finally {
    await browser.close();
  }
})();
