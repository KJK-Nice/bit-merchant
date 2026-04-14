import { expect, test } from "@playwright/test";
import { MerchantActor } from "../actors/MerchantActor";
import { AssertOnMerchantHost } from "../questions/AssertOnMerchantHost";
import { AssertPath } from "../questions/AssertPath";
import { LoginWithPasskey } from "../tasks/LoginWithPasskey";
import { Logout } from "../tasks/Logout";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

test.describe("Merchant auth access journey", () => {
  test.skip("returning owner can sign in with passkey and land on dashboard", async ({ browser }) => {
    // TODO: Passkey re-login currently stalls at WebAuthn assertion in Playwright virtual authenticator.
    const context = await browser.newContext();
    const merchant = await MerchantActor(context);

    const suffix = Date.now().toString();
    await merchant.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `Return Login Cafe ${suffix}`));
    await expect(merchant.page).toHaveURL(/\/dashboard$/);

    await merchant.attemptsTo(Logout(), LoginWithPasskey());
    await expect(merchant.page).toHaveURL(/\/dashboard$/);
    expect(await merchant.asks(AssertOnMerchantHost())).toBeTruthy();
    expect(await merchant.asks(AssertPath("/dashboard"))).toBeTruthy();

    await context.close();
  });

  test("after logout, protected merchant routes redirect to auth login", async ({ browser }) => {
    const context = await browser.newContext();
    const merchant = await MerchantActor(context);

    const suffix = Date.now().toString();
    await merchant.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `Return Login Cafe ${suffix}`));
    await expect(merchant.page).toHaveURL(/\/dashboard$/);
    expect(await merchant.asks(AssertOnMerchantHost())).toBeTruthy();
    expect(await merchant.asks(AssertPath("/dashboard"))).toBeTruthy();

    await merchant.attemptsTo(Logout());
    await expect(merchant.page).toHaveURL(/\/$/);

    await merchant.attemptsTo(OpenRoute("merchant", "/dashboard"));
    await expect(merchant.page).toHaveURL(/\/auth\/login$/);
    expect(await merchant.asks(AssertOnMerchantHost())).toBeTruthy();
    expect(await merchant.asks(AssertPath("/auth/login"))).toBeTruthy();

    await context.close();
  });
});
