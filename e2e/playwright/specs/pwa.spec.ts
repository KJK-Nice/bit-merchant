import { expect, test } from "@playwright/test";
import { surfaces } from "../support/surfaces";

// ─── helpers ────────────────────────────────────────────────────────────────

const customerURL = (path: string) => `${surfaces.customer}${path}`;

// Wait for the active service worker on the customer surface.
async function waitForSW(page: import("@playwright/test").Page) {
  await page.waitForFunction(
    () =>
      navigator.serviceWorker.controller !== null ||
      navigator.serviceWorker
        .getRegistration("/")
        .then((r) => r?.active !== null),
    { timeout: 10_000 }
  );
}

// ─── manifest ────────────────────────────────────────────────────────────────

test.describe("PWA manifest", () => {
  test("returns valid JSON with required fields", async ({ request }) => {
    const res = await request.get(
      `${surfaces.customer}/static/pwa/manifest.json`
    );
    expect(res.ok()).toBeTruthy();

    const manifest = await res.json();
    expect(manifest).toMatchObject({
      id: expect.any(String),
      name: expect.any(String),
      short_name: expect.any(String),
      start_url: expect.any(String),
      display: expect.any(String),
      icons: expect.arrayContaining([
        expect.objectContaining({ src: expect.any(String), sizes: expect.any(String) }),
      ]),
    });

    // Must have at least one maskable icon for Android adaptive icons.
    const maskable = manifest.icons.filter(
      (i: { purpose?: string }) => i.purpose?.includes("maskable")
    );
    expect(maskable.length).toBeGreaterThanOrEqual(1);
  });

  test("192px and 512px PNG icons are reachable", async ({ request }) => {
    for (const size of ["192", "512"]) {
      const res = await request.get(
        `${surfaces.customer}/static/pwa/icons/icon-${size}.png`
      );
      expect(res.ok(), `icon-${size}.png not found`).toBeTruthy();
    }
  });
});

// ─── security headers ────────────────────────────────────────────────────────

test.describe("Security headers", () => {
  test("CSP header is present and carries a nonce", async ({ request }) => {
    const res = await request.get(customerURL("/menu?restaurantID=restaurant_1"));
    const csp = res.headers()["content-security-policy"];
    expect(csp, "CSP header missing").toBeTruthy();
    expect(csp).toContain("script-src");
    expect(csp).toMatch(/nonce-[A-Za-z0-9+/]+=*/);
  });

  test("X-Content-Type-Options is nosniff", async ({ request }) => {
    const res = await request.get(customerURL("/menu?restaurantID=restaurant_1"));
    expect(res.headers()["x-content-type-options"]).toBe("nosniff");
  });
});

// ─── service worker ──────────────────────────────────────────────────────────

test.describe("Service worker", () => {
  test("registers successfully and activates", async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();

    await page.goto(customerURL("/menu?restaurantID=restaurant_1"));
    await page.waitForLoadState("networkidle");

    const swState = await page.evaluate(async () => {
      const reg = await navigator.serviceWorker.getRegistration("/");
      return reg?.active?.state ?? null;
    });

    expect(swState).toBe("activated");
    await context.close();
  });

  test("/sw.js carries no-cache headers and Service-Worker-Allowed header", async ({
    request,
  }) => {
    const res = await request.get(`${surfaces.customer}/sw.js`);
    expect(res.ok()).toBeTruthy();
    const cc = res.headers()["cache-control"] ?? "";
    expect(cc).toContain("no-cache");
    expect(res.headers()["service-worker-allowed"]).toBe("/");
  });

  test("/offline page is reachable", async ({ request }) => {
    const res = await request.get(`${surfaces.customer}/offline`);
    // Expect 503 (ServiceUnavailable) — that's intentional.
    expect([200, 503]).toContain(res.status());
    const body = await res.text();
    expect(body).toContain("offline");
  });

  test("/sw-kill.js escape hatch is reachable", async ({ request }) => {
    const res = await request.get(`${surfaces.customer}/sw-kill.js`);
    expect(res.ok()).toBeTruthy();
    expect(await res.text()).toContain("unregister");
  });
});

// ─── offline menu browsing ───────────────────────────────────────────────────

test.describe("Offline menu browsing", () => {
  test("menu page is served from SW cache when offline", async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();

    // First visit — populates the SW runtime cache.
    await page.goto(customerURL("/menu?restaurantID=restaurant_1"));
    await page.waitForLoadState("networkidle");
    await waitForSW(page);

    // Go offline.
    await context.setOffline(true);

    // Reload the same menu URL — SW should serve from cache.
    const [response] = await Promise.all([
      page.waitForResponse(customerURL("/menu?restaurantID=restaurant_1"), {
        timeout: 5_000,
      }).catch(() => null),
      page.reload(),
    ]);

    // The page should still render (not show a browser offline error).
    const title = await page.title();
    expect(title).not.toBe(""); // browser offline page has no title
    expect(title).not.toContain("ERR_"); // Chrome network error pages

    await context.close();
  });
});

// ─── canonical-host surface routing regression ───────────────────────────────

test.describe("Surface routing regression", () => {
  test("navigate to wrong surface redirects correctly with SW registered", async ({
    browser,
  }) => {
    // Register SW on the customer surface first.
    const context = await browser.newContext();
    const page = await context.newPage();

    await page.goto(customerURL("/menu?restaurantID=restaurant_1"));
    await page.waitForLoadState("networkidle");
    await waitForSW(page);

    // Directly navigate to the public surface — the routing middleware
    // should redirect to the appropriate host. The SW must NOT intercept
    // this navigate and break the redirect.
    const publicURL = `${surfaces.public}/menu?restaurantID=restaurant_1`;
    const finalURL = await page
      .goto(publicURL, { waitUntil: "networkidle" })
      .then((r) => r?.url() ?? "");

    // Should have landed on the customer surface (redirected by server).
    expect(finalURL).toContain(surfaces.customer.replace("http://", ""));

    await context.close();
  });
});

// ─── authenticated surface cache isolation ───────────────────────────────────

test.describe("Authenticated surface cache isolation", () => {
  test("SW never caches merchant dashboard responses", async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();

    // Load merchant surface with SW active from prior tests (or fresh here).
    await page.goto(`${surfaces.merchant}/dashboard`);
    await page.waitForLoadState("networkidle");

    // Check SW cache — merchant paths must not appear.
    const cachedMerchantURLs: string[] = await page.evaluate(async () => {
      const keys = await caches.keys();
      const merchantURLs: string[] = [];
      for (const key of keys) {
        const cache = await caches.open(key);
        const requests = await cache.keys();
        for (const req of requests) {
          if (new URL(req.url).pathname.startsWith("/merchant") ||
              new URL(req.url).pathname.startsWith("/dashboard")) {
            merchantURLs.push(req.url);
          }
        }
      }
      return merchantURLs;
    });

    expect(cachedMerchantURLs).toHaveLength(0);
    await context.close();
  });
});
