import { expect, test } from "@playwright/test";
import { CustomerActor } from "../actors/CustomerActor";
import { MerchantActor } from "../actors/MerchantActor";
import { CookiesForHost } from "../questions/CookiesForHost";
import { OpenRoute } from "../tasks/OpenRoute";

test.describe("Session cookie isolation by surface", () => {
  test("customer and merchant session cookies are separated by host", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);
    const merchant = await MerchantActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));
    const customerCookies = await customer.asks(CookiesForHost("customer"));
    expect(customerCookies).toContain("bitmerchant_customer_session");
    expect(customerCookies).not.toContain("bitmerchant_merchant_session");

    await merchant.attemptsTo(OpenRoute("merchant", "/dashboard"));
    const merchantCookies = await merchant.asks(CookiesForHost("merchant"));
    expect(merchantCookies).toContain("bitmerchant_merchant_session");
    expect(merchantCookies).not.toContain("bitmerchant_customer_session");

    const customerCookiesAfterMerchantVisit = await customer.asks(CookiesForHost("customer"));
    expect(customerCookiesAfterMerchantVisit).toContain("bitmerchant_customer_session");
    expect(customerCookiesAfterMerchantVisit).not.toContain("bitmerchant_merchant_session");

    await context.close();
  });
});
