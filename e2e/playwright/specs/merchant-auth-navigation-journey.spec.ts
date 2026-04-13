import { expect, test } from "@playwright/test";
import { MerchantActor } from "../actors/MerchantActor";
import { AssertOnMerchantHost } from "../questions/AssertOnMerchantHost";
import { AssertPath } from "../questions/AssertPath";
import { Logout } from "../tasks/Logout";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

test.describe("Merchant full journey (passkey + core nav)", () => {
  test.describe.configure({ retries: 1 });

  test("signup with passkey, access merchant surfaces, logout to public entry", async ({ browser }) => {
    test.setTimeout(60_000);

    const context = await browser.newContext();
    const merchant = await MerchantActor(context);

    const suffix = Date.now().toString();
    await merchant.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `E2E Cafe ${suffix}`));
    await expect(merchant.page).toHaveURL(/\/dashboard$/);
    expect(await merchant.asks(AssertOnMerchantHost())).toBeTruthy();
    expect(await merchant.asks(AssertPath("/dashboard"))).toBeTruthy();

    await merchant.attemptsTo(OpenRoute("merchant", "/admin/dashboard"));
    await expect(merchant.page).toHaveURL(/\/admin\/dashboard$/);
    await expect(merchant.page.getByRole("heading", { name: "Menu Management", exact: true })).toBeVisible();

    await merchant.attemptsTo(OpenRoute("merchant", "/admin/qr"));
    await expect(merchant.page).toHaveURL(/\/admin\/qr$/);
    await expect(merchant.page.getByRole("heading", { name: "Table QR codes" })).toBeVisible();

    await merchant.attemptsTo(OpenRoute("merchant", "/kitchen"));
    await expect(merchant.page).toHaveURL(/\/kitchen$/);
    await expect(merchant.page).toHaveTitle(/Kitchen Display/);

    await merchant.attemptsTo(OpenRoute("merchant", "/dashboard"), Logout());
    await expect(merchant.page).toHaveURL(/\/$/);

    await context.close();
  });
});
