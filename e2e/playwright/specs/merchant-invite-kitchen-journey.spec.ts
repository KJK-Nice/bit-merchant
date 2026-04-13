import { expect, test } from "@playwright/test";
import { urlFor } from "../abilities/BrowseTheWeb";
import { MerchantActor } from "../actors/MerchantActor";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

const csrfTokenForMerchantHost = async (actor: Awaited<ReturnType<typeof MerchantActor>>): Promise<string> => {
  const cookies = await actor.context.cookies(urlFor("merchant", "/"));
  const csrfCookie = cookies.find((cookie) => cookie.name === "csrf");
  return csrfCookie?.value ?? "";
};

const createKitchenInviteURL = async (
  actor: Awaited<ReturnType<typeof MerchantActor>>,
  csrf: string,
): Promise<string> => {
  const inviteURL = await actor.page.evaluate(async (csrfToken) => {
    const response = await fetch("/dashboard/invite", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrfToken,
      },
    });

    if (!response.ok) {
      throw new Error(`invite request failed with status ${response.status}`);
    }

    const payload = (await response.json()) as { inviteURL?: string };
    if (!payload.inviteURL) {
      throw new Error("inviteURL missing in response");
    }
    return payload.inviteURL;
  }, csrf);

  return inviteURL;
};

test.describe("Merchant invite onboarding journey", () => {
  test("owner can invite kitchen staff and invited user is restricted from owner-only admin", async ({ browser }) => {
    const ownerContext = await browser.newContext();
    const owner = await MerchantActor(ownerContext);

    const suffix = Date.now().toString();
    await owner.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `Invite Cafe ${suffix}`));
    await expect(owner.page).toHaveURL(/\/dashboard$/);

    const csrf = await csrfTokenForMerchantHost(owner);
    expect(csrf).not.toBe("");

    const inviteURL = await createKitchenInviteURL(owner, csrf);
    expect(inviteURL).toMatch(/^\/auth\/invite\/.+/);

    const kitchenContext = await browser.newContext();
    const kitchenUser = await MerchantActor(kitchenContext);

    await kitchenUser.attemptsTo(OpenRoute("merchant", inviteURL));
    await expect(kitchenUser.page.getByRole("heading", { name: "Accept invitation" })).toBeVisible();

    await kitchenUser.page.getByLabel("Your name").fill(`Kitchen ${suffix}`);
    await Promise.all([
      kitchenUser.page.waitForURL("**/kitchen"),
      kitchenUser.page.getByRole("button", { name: "Accept and register passkey" }).click(),
    ]);
    await expect(kitchenUser.page).toHaveURL(/\/kitchen$/);

    const ownerOnlyResponse = await kitchenUser.page.goto(urlFor("merchant", "/admin/dashboard"), {
      waitUntil: "domcontentloaded",
    });
    expect(ownerOnlyResponse?.status()).toBe(403);
    await expect(kitchenUser.page.getByText("forbidden")).toBeVisible();

    await ownerContext.close();
    await kitchenContext.close();
  });

  test("invite link is single-use after kitchen user accepts invitation", async ({ browser }) => {
    const ownerContext = await browser.newContext();
    const owner = await MerchantActor(ownerContext);

    const suffix = `${Date.now()}-reuse`;
    await owner.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `Invite Login Cafe ${suffix}`));
    await expect(owner.page).toHaveURL(/\/dashboard$/);

    const csrf = await csrfTokenForMerchantHost(owner);
    expect(csrf).not.toBe("");

    const inviteURL = await createKitchenInviteURL(owner, csrf);
    expect(inviteURL).toMatch(/^\/auth\/invite\/.+/);

    const kitchenContext = await browser.newContext();
    const kitchenUser = await MerchantActor(kitchenContext);
    await kitchenUser.attemptsTo(OpenRoute("merchant", inviteURL));
    await kitchenUser.page.getByLabel("Your name").fill(`Kitchen ${suffix}`);
    await Promise.all([
      kitchenUser.page.waitForURL("**/kitchen"),
      kitchenUser.page.getByRole("button", { name: "Accept and register passkey" }).click(),
    ]);
    await expect(kitchenUser.page).toHaveURL(/\/kitchen$/);

    const secondAttemptContext = await browser.newContext();
    const secondAttemptUser = await MerchantActor(secondAttemptContext);
    await secondAttemptUser.attemptsTo(OpenRoute("merchant", inviteURL));
    await expect(secondAttemptUser.page.getByText("Invitation is no longer valid")).toBeVisible();

    await ownerContext.close();
    await kitchenContext.close();
    await secondAttemptContext.close();
  });
});
