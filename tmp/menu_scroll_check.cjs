const { chromium, devices } = require('@playwright/test');
(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ ...devices['iPhone 13'], ignoreHTTPSErrors: true });
  const page = await context.newPage();
  const url = 'https://order.bitmerchant.local:8080/menu?restaurantID=rest_1776224541045085547';
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.waitForTimeout(1200);
  await page.screenshot({ path: '/Users/kjkusap/dev/bit-merchant/tmp/nav-audit/menu_tabs_before_scroll.png' });
  await page.mouse.wheel(0, 1200);
  await page.waitForTimeout(700);
  await page.screenshot({ path: '/Users/kjkusap/dev/bit-merchant/tmp/nav-audit/menu_tabs_after_scroll.png' });
  await browser.close();
})();
