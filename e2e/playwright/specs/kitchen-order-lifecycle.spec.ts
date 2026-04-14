import { expect, test } from "@playwright/test";
import { urlFor } from "../abilities/BrowseTheWeb";
import { CustomerActor } from "../actors/CustomerActor";
import { MerchantActor } from "../actors/MerchantActor";
import { ExtractOrderNumberFromURL } from "../questions/ExtractOrderNumberFromURL";
import { AddMenuItemToCart } from "../tasks/AddMenuItemToCart";
import { OpenRoute } from "../tasks/OpenRoute";
import { PlaceCashOrder } from "../tasks/PlaceCashOrder";
import { ProceedToCheckout } from "../tasks/ProceedToCheckout";
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

const createCategoryAndItem = async (
  actor: Awaited<ReturnType<typeof MerchantActor>>,
  csrf: string,
  suffix: string,
): Promise<void> => {
  await actor.page.evaluate(
    async ({ csrfToken, tokenSuffix }) => {
      const categoryBody = new URLSearchParams({
        csrf: csrfToken,
        name: `Category ${tokenSuffix}`,
        displayOrder: "1",
      });
      const categoryResp = await fetch("/admin/category", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          "X-CSRF-Token": csrfToken,
        },
        body: categoryBody.toString(),
      });
      if (!categoryResp.ok) {
        throw new Error(`create category failed: ${categoryResp.status}`);
      }

      const dashboardResp = await fetch("/admin/dashboard", { method: "GET" });
      const html = await dashboardResp.text();
      const categoryMatch = html.match(/data-category-id=\"([^\"]+)\"/);
      if (!categoryMatch?.[1]) {
        throw new Error("categoryID not found after category creation");
      }
      const categoryID = categoryMatch[1];

      const itemBody = new URLSearchParams({
        csrf: csrfToken,
        categoryID,
        name: `Item ${tokenSuffix}`,
        price: "9.50",
        description: "E2E menu item",
        available: "on",
      });
      const itemResp = await fetch("/admin/item", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          "X-CSRF-Token": csrfToken,
        },
        body: itemBody.toString(),
      });
      if (!itemResp.ok) {
        throw new Error(`create item failed: ${itemResp.status}`);
      }
    },
    { csrfToken: csrf, tokenSuffix: suffix },
  );
};

test.describe("Kitchen order lifecycle", () => {
  test("kitchen staff can move an order paid -> preparing -> ready and customer observes status updates", async ({
    browser,
  }) => {
    const ownerContext = await browser.newContext();
    const owner = await MerchantActor(ownerContext);

    const suffix = Date.now().toString();
    await owner.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `Kitchen Flow Cafe ${suffix}`));
    await expect(owner.page).toHaveURL(/\/dashboard$/);

    await owner.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    const restaurantID = await owner.page.locator("#restaurantID option[selected]").getAttribute("value");
    expect(restaurantID).toBeTruthy();

    const csrf = await csrfTokenForMerchantHost(owner);
    expect(csrf).not.toBe("");
    await createCategoryAndItem(owner, csrf, suffix);

    const inviteURL = await createKitchenInviteURL(owner, csrf);
    expect(inviteURL).toMatch(/^\/auth\/invite\/.+/);

    const kitchenContext = await browser.newContext();
    const kitchenStaff = await MerchantActor(kitchenContext);
    await kitchenStaff.attemptsTo(OpenRoute("merchant", inviteURL));
    await kitchenStaff.page.getByLabel("Your name").fill(`Kitchen ${suffix}`);
    await Promise.all([
      kitchenStaff.page.waitForURL("**/kitchen", { waitUntil: "domcontentloaded" }),
      kitchenStaff.page.getByRole("button", { name: "Accept and register passkey" }).click(),
    ]);
    await expect(kitchenStaff.page).toHaveURL(/\/kitchen$/);

    const customerContext = await browser.newContext();
    const customer = await CustomerActor(customerContext);
    await customer.attemptsTo(OpenRoute("customer", `/menu?restaurantID=${restaurantID}`));
    await customer.attemptsTo(AddMenuItemToCart(), ProceedToCheckout(), PlaceCashOrder());
    await expect(customer.page).toHaveURL(/\/order\/[^/]+$/);

    const orderNumber = await customer.asks(ExtractOrderNumberFromURL());
    expect(orderNumber).not.toBe("");
    const customerStatus = customer.page.locator("#status-display");
    await expect(customerStatus).toContainText("UNPAID");
    await expect(customerStatus).toContainText("paid");

    await kitchenStaff.attemptsTo(OpenRoute("merchant", "/kitchen"));
    const kitchenOrderCard = kitchenStaff.page.locator("div[id^='order-']").filter({ hasText: `Order #${orderNumber}` });
    await expect(kitchenOrderCard).toBeVisible();

    await kitchenOrderCard.getByRole("button", { name: "Mark Paid" }).click();
    await expect(kitchenOrderCard).toContainText("PAID");
    await expect(customerStatus).toContainText("PAID");

    await kitchenOrderCard.getByRole("button", { name: "Start Preparing" }).click();
    await expect(customerStatus).toContainText("preparing");

    await kitchenOrderCard.getByRole("button", { name: "Mark Ready" }).click();
    await expect(customerStatus).toContainText("ready");

    await ownerContext.close();
    await kitchenContext.close();
    await customerContext.close();
  });
});
