import { expect, test } from "@playwright/test";
import { Recall, Remember } from "../abilities/Remember";
import { CustomerActor } from "../actors/CustomerActor";
import { CurrentPathname } from "../questions/CurrentPathname";
import { ExtractOrderNumberFromURL } from "../questions/ExtractOrderNumberFromURL";
import { TextVisible } from "../questions/TextVisible";
import { AddMenuItemToCart } from "../tasks/AddMenuItemToCart";
import { OpenRoute } from "../tasks/OpenRoute";
import { PlaceCashOrder } from "../tasks/PlaceCashOrder";
import { ProceedToCheckout } from "../tasks/ProceedToCheckout";

test.describe("Customer full journey", () => {
  test("menu to cart to checkout to order status", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));
    await expect(customer.page.getByRole("heading", { name: "BitMerchant Cafe" })).toBeVisible();

    await customer.attemptsTo(AddMenuItemToCart(), ProceedToCheckout());
    await expect(customer.page).toHaveURL(/\/order\/confirm$/);
    await expect(customer.page.getByRole("heading", { name: "Confirm Order" })).toBeVisible();

    await customer.attemptsTo(PlaceCashOrder());
    const path = await customer.asks(CurrentPathname());
    expect(path).toMatch(/^\/order\/[^/]+$/);

    const orderNumber = await customer.asks(ExtractOrderNumberFromURL());
    await customer.attemptsTo(Remember("orderNumber", orderNumber));

    const storedOrderNumber = await customer.asks(Recall<string>("orderNumber"));
    expect(storedOrderNumber).toBe(orderNumber);

    const statusHeadingVisible = await customer.asks(TextVisible(`Order #${orderNumber}`));
    expect(statusHeadingVisible).toBeTruthy();
    await expect(customer.page.getByText("Total:")).toBeVisible();
    await expect(customer.page.getByText("Items")).toBeVisible();

    await context.close();
  });
});
