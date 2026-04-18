import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e/playwright/specs",
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  fullyParallel: true,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: process.env.CI ? [["github"], ["html", { open: "never" }]] : [["list"], ["html", { open: "never" }]],
  use: {
    serviceWorkers: "allow",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "mobile-chrome-android",
      use: {
        ...devices["Pixel 7"],
        browserName: "chromium",
      },
    },
    {
      name: "mobile-chrome-iphone",
      use: {
        ...devices["iPhone 13"],
        browserName: "chromium",
      },
    },
  ],
  webServer: {
    command: "go run ./cmd/server",
    port: 8080,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      PORT: "8080",
      PUBLIC_BASE_URL: "http://localhost:8080",
      CUSTOMER_BASE_URL: "http://order.localhost:8080",
      MERCHANT_BASE_URL: "http://merchant.localhost:8080",
      DISABLE_RATE_LIMIT: "true",
    },
  },
});
