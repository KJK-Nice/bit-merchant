import { expect, test, type BrowserContext } from "@playwright/test";
import { MerchantActor } from "../actors/MerchantActor";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

const disableWebAuthn = async (context: BrowserContext) => {
  await context.addInitScript(() => {
    (window as any).__passkeyUnsupportedE2E = true;
    (window as any).__BITMERCHANT_FORCE_PASSKEY_UNSUPPORTED = true;

    try {
      // Attempt to remove WebAuthn entrypoint.
      (window as any).PublicKeyCredential = undefined;
    } catch (_) {
      // no-op
    }

    try {
      // Attempt to remove credentials methods directly.
      if (navigator.credentials) {
        (navigator.credentials as any).create = undefined;
        (navigator.credentials as any).get = undefined;
      }
    } catch (_) {
      // no-op
    }

    try {
      // Fallback: override navigator.credentials getter when configurable.
      Object.defineProperty(Navigator.prototype, "credentials", {
        configurable: true,
        get() {
          return undefined;
        },
      });
    } catch (_) {
      // no-op
    }
  });
};

const createKitchenInviteURL = async (merchant: Awaited<ReturnType<typeof MerchantActor>>): Promise<string> => {
  return merchant.page.evaluate(async () => {
    const response = await fetch("/dashboard/invite", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ role: "kitchen_staff" }),
    });

    if (!response.ok) {
      throw new Error(`invite request failed with status ${response.status}`);
    }
    const payload = (await response.json()) as { inviteURL?: string };
    if (!payload.inviteURL) {
      throw new Error("inviteURL missing in response");
    }
    return payload.inviteURL;
  });
};

test.describe("Passkey unsupported runtime guard", () => {
  test.beforeEach(async ({}, testInfo) => {
    test.skip(testInfo.project.name !== "chromium", "This compatibility simulation runs on desktop chromium only.");
  });

  test("signup shows inline error and disables submit when WebAuthn is unavailable", async ({ browser }) => {
    const context = await browser.newContext();
    await disableWebAuthn(context);
    const page = await context.newPage();

    await page.goto("http://merchant.localhost:8080/auth/signup", { waitUntil: "domcontentloaded" });
    await expect(await page.evaluate(() => (window as any).__passkeyUnsupportedE2E === true)).toBeTruthy();

    const submit = page.getByRole("button", { name: "Create with passkey" });
    await expect(submit).toBeDisabled();
    await expect(page.getByRole("alert")).toContainText("Passkeys are not available in this browser/context");

    await context.close();
  });

  test("login shows inline error and disables submit when WebAuthn is unavailable", async ({ browser }) => {
    const context = await browser.newContext();
    await disableWebAuthn(context);
    const page = await context.newPage();

    await page.goto("http://merchant.localhost:8080/auth/login", { waitUntil: "domcontentloaded" });
    await expect(await page.evaluate(() => (window as any).__passkeyUnsupportedE2E === true)).toBeTruthy();

    const submit = page.getByRole("button", { name: "Sign in" });
    await expect(submit).toBeDisabled();
    await expect(page.getByRole("alert")).toContainText("Passkeys are not available in this browser/context");

    await context.close();
  });

  test("invite signup shows inline error and disables submit when WebAuthn is unavailable", async ({ browser }) => {
    const ownerCtx = await browser.newContext();
    const owner = await MerchantActor(ownerCtx);

    const suffix = Date.now().toString();
    await owner.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `Unsupported Invite Cafe ${suffix}`));
    await owner.attemptsTo(OpenRoute("merchant", "/dashboard"));

    const inviteURL = await createKitchenInviteURL(owner);

    await ownerCtx.close();

    const inviteCtx = await browser.newContext();
    await disableWebAuthn(inviteCtx);
    const invitePage = await inviteCtx.newPage();
    await invitePage.goto(`http://merchant.localhost:8080${inviteURL}`, { waitUntil: "domcontentloaded" });
    await expect(await invitePage.evaluate(() => (window as any).__passkeyUnsupportedE2E === true)).toBeTruthy();

    const submit = invitePage.getByRole("button", { name: "Accept and register passkey" });
    await expect(submit).toBeDisabled();
    await expect(invitePage.getByRole("alert")).toContainText("Passkeys are not available in this browser/context");

    await inviteCtx.close();
  });
});
