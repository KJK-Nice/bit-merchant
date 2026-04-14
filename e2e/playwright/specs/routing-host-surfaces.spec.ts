import { expect, test } from "@playwright/test";
import { CustomerActor } from "../actors/CustomerActor";
import { MerchantActor } from "../actors/MerchantActor";
import { PublicVisitor } from "../actors/PublicVisitor";
import { CurrentURL } from "../questions/CurrentURL";
import { OpenRoute } from "../tasks/OpenRoute";

test.describe("Host-based routing", () => {
  test("customer route on merchant host redirects to customer host and preserves query", async ({ browser }) => {
    const context = await browser.newContext();
    const merchant = await MerchantActor(context);

    await merchant.attemptsTo(OpenRoute("merchant", "/menu?restaurantID=restaurant_1&table=3"));
    const current = await merchant.asks(CurrentURL());

    expect(current.host).toBe("order.localhost:8080");
    expect(current.pathname).toBe("/menu");
    expect(current.searchParams.get("restaurantID")).toBe("restaurant_1");
    expect(current.searchParams.get("table")).toBe("3");

    await context.close();
  });

  test("merchant route on customer host redirects to merchant host and ends on auth login", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/dashboard"));
    const current = await customer.asks(CurrentURL());

    expect(current.host).toBe("merchant.localhost:8080");
    expect(current.pathname).toBe("/auth/login");

    await context.close();
  });

  test("public host serves entry page", async ({ browser }) => {
    const context = await browser.newContext();
    const visitor = await PublicVisitor(context);

    await visitor.attemptsTo(OpenRoute("public", "/"));
    await expect(visitor.page.getByText("Scan a table QR to open a menu")).toBeVisible();

    const current = await visitor.asks(CurrentURL());
    expect(current.host).toBe("localhost:8080");
    expect(current.pathname).toBe("/");

    await context.close();
  });

  test("same-surface customer route stays on customer host (no cross-host redirect)", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));
    const current = await customer.asks(CurrentURL());

    expect(current.host).toBe("order.localhost:8080");
    expect(current.pathname).toBe("/menu");

    await context.close();
  });
});
