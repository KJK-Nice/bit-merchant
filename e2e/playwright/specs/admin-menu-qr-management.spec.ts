import { expect, test } from "@playwright/test";
import { urlFor } from "../abilities/BrowseTheWeb";
import { CustomerActor } from "../actors/CustomerActor";
import { MerchantActor } from "../actors/MerchantActor";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

const csrfTokenForMerchantHost = async (actor: Awaited<ReturnType<typeof MerchantActor>>): Promise<string> => {
  const cookies = await actor.context.cookies(urlFor("merchant", "/"));
  const csrfCookie = cookies.find((cookie) => cookie.name === "csrf");
  return csrfCookie?.value ?? "";
};

const createCategoryAndItem = async (
  actor: Awaited<ReturnType<typeof MerchantActor>>,
  csrf: string,
  suffix: string,
): Promise<{ itemName: string }> => {
  const itemName = `Item ${suffix}`;
  await actor.page.evaluate(
    async ({ csrfToken, tokenSuffix, createdItemName }) => {
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
        name: createdItemName,
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
    { csrfToken: csrf, tokenSuffix: suffix, createdItemName: itemName },
  );

  return { itemName };
};

test.describe("Admin menu and QR management", () => {
  test.describe.configure({ mode: "serial", retries: 1 });

  test("owner toggles item availability and customer menu reflects out-of-stock state", async ({ browser }) => {
    test.setTimeout(90_000);

    const merchantContext = await browser.newContext();
    const owner = await MerchantActor(merchantContext);
    const suffix = Date.now().toString();

    await owner.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `Availability Cafe ${suffix}`));
    await expect(owner.page).toHaveURL(/\/dashboard$/);

    await owner.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    const restaurantID = await owner.page.locator("#restaurantID option[selected]").getAttribute("value");
    expect(restaurantID).toBeTruthy();

    const csrf = await csrfTokenForMerchantHost(owner);
    expect(csrf).not.toBe("");
    const { itemName } = await createCategoryAndItem(owner, csrf, suffix);

    await owner.attemptsTo(OpenRoute("merchant", "/admin/dashboard"));
    const row = owner.page.locator("tr").filter({ hasText: itemName }).first();
    await expect(row).toBeVisible();
    await row.getByRole("button", { name: "Mark unavailable" }).click();
    await expect(row.getByText("Unavailable")).toBeVisible();

    const customerContext = await browser.newContext();
    const customer = await CustomerActor(customerContext);
    await customer.attemptsTo(OpenRoute("customer", `/menu?restaurantID=${restaurantID}`));

    // Customer-facing menu may either hide unavailable items or show them as out-of-stock.
    // Assert the stable behavior: the toggled item cannot be ordered.
    await expect(customer.page.getByText(itemName)).toHaveCount(0);
    await expect(customer.page.getByRole("button", { name: "Add to Cart" })).toHaveCount(0);

    await merchantContext.close();
    await customerContext.close();
  });

  test("owner updates QR table settings and customer menu link with table parameter remains valid", async ({
    browser,
  }) => {
    test.setTimeout(90_000);

    const merchantContext = await browser.newContext();
    const owner = await MerchantActor(merchantContext);
    const suffix = `${Date.now()}-qr`;

    await owner.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, `QR Cafe ${suffix}`));
    await expect(owner.page).toHaveURL(/\/dashboard$/);

    await owner.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    const restaurantID = await owner.page.locator("#restaurantID option[selected]").getAttribute("value");
    expect(restaurantID).toBeTruthy();

    await owner.attemptsTo(OpenRoute("merchant", "/admin/qr"));
    await expect(owner.page.getByRole("heading", { name: "Table QR codes" })).toBeVisible();

    await owner.page.getByLabel("Tables").fill("3");
    await Promise.all([
      owner.page.waitForURL("**/admin/qr?saved=1", { waitUntil: "domcontentloaded" }),
      owner.page.getByRole("button", { name: "Save" }).click(),
    ]);
    await expect(owner.page.getByText("Settings saved.")).toBeVisible();
    await expect(owner.page.getByText(/Table 3/)).toBeVisible();

    const customerContext = await browser.newContext();
    const customer = await CustomerActor(customerContext);
    await customer.attemptsTo(OpenRoute("customer", `/menu?restaurantID=${restaurantID}&table=3`));
    await expect(customer.page.getByText("Table 3")).toBeVisible();

    await merchantContext.close();
    await customerContext.close();
  });
});
