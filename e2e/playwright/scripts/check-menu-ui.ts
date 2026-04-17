import { chromium } from "playwright";
import path from "path";
import fs from "fs";

const BASE = "https://order.bitmerchant.local:8080";
const MENU_URL = `${BASE}/menu?restaurantID=restaurant_1`;
const OUT_DIR = path.resolve("test-results/menu-ui-check");

async function run() {
  fs.mkdirSync(OUT_DIR, { recursive: true });

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    ignoreHTTPSErrors: true,
    viewport: { width: 390, height: 844 },
  });
  const page = await context.newPage();

  await page.goto(MENU_URL, { waitUntil: "domcontentloaded", timeout: 15_000 });
  await page.waitForTimeout(500);

  const myPlacesInHeader = await page.locator("header[data-menu-sticky-header] a[href='/my-places']").count();
  console.log("'My places' button in header:", myPlacesInHeader === 0 ? "REMOVED ✓" : "STILL PRESENT ✗");

  await page.screenshot({ path: `${OUT_DIR}/no-myplaces-mobile.png`, fullPage: false });
  console.log("Screenshot: no-myplaces-mobile.png");

  await browser.close();
}

run().catch((err) => { console.error(err); process.exit(1); });
