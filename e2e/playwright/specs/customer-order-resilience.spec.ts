import { expect, test } from "@playwright/test";
import { CustomerActor } from "../actors/CustomerActor";
import { ExtractOrderNumberFromURL } from "../questions/ExtractOrderNumberFromURL";
import { AddMenuItemToCart } from "../tasks/AddMenuItemToCart";
import { OpenRoute } from "../tasks/OpenRoute";
import { PlaceCashOrder } from "../tasks/PlaceCashOrder";
import { ProceedToCheckout } from "../tasks/ProceedToCheckout";

test.describe("Customer order resilience", () => {
  test("customer can recover an order from /order/lookup history", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));
    await customer.attemptsTo(AddMenuItemToCart(), ProceedToCheckout(), PlaceCashOrder());
    await expect(customer.page).toHaveURL(/\/order\/[^/]+$/);

    const orderNumber = await customer.asks(ExtractOrderNumberFromURL());
    expect(orderNumber).not.toBe("");

    await customer.attemptsTo(OpenRoute("customer", "/order/lookup"));
    await expect(customer.page.getByRole("heading", { name: "Your Orders" })).toBeVisible();
    await expect(customer.page.getByRole("link", { name: new RegExp(`Order #${orderNumber}`) })).toBeVisible();

    await Promise.all([
      customer.page.waitForURL(`**/order/${orderNumber}`, { waitUntil: "domcontentloaded" }),
      customer.page.getByRole("link", { name: new RegExp(`Order #${orderNumber}`) }).click(),
    ]);
    await expect(customer.page.getByText(`Order #${orderNumber}`)).toBeVisible();

    await context.close();
  });

  test("customer can navigate back to menu from confirmation to edit cart", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));
    await customer.attemptsTo(AddMenuItemToCart(), ProceedToCheckout());
    await expect(customer.page).toHaveURL(/\/order\/confirm(\?.*)?$/);
    await expect(customer.page.getByRole("button", { name: /Send to kitchen/ })).toBeVisible();

    // "Edit cart" returns to the menu with restaurant context preserved so the
    // customer can adjust quantities on the menu's cart drawer/floating button.
    await Promise.all([
      customer.page.waitForURL(/\/menu\?.*restaurantID=restaurant_1.*/),
      customer.page.getByRole("link", { name: "Edit cart" }).click(),
    ]);

    await context.close();
  });

  test("opening /order/confirm with an empty cart redirects to entry with restaurant_required reason", async ({
    browser,
  }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));
    await customer.attemptsTo(OpenRoute("customer", "/order/confirm"));

    await expect(customer.page).toHaveURL(/\/\?reason=restaurant_required$/);
    await expect(customer.page.getByText("Restaurant context is required")).toBeVisible();

    await context.close();
  });
});
