import { expect, test } from "@playwright/test";
import { MerchantActor } from "../actors/MerchantActor";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

test.describe("Merchant multi-restaurant selection", () => {
  test("user with multiple memberships can switch active restaurant from selection screen", async ({
    browser,
  }) => {
    const context = await browser.newContext();
    const merchant = await MerchantActor(context);

    const suffix = Date.now().toString();
    const firstRestaurantName = `First Cafe ${suffix}`;
    const secondRestaurantName = `Second Cafe ${suffix}`;

    await merchant.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, firstRestaurantName));
    await expect(merchant.page).toHaveURL(/\/dashboard$/);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/restaurants/new"));
    await expect(merchant.page.getByRole("heading", { name: "New restaurant" })).toBeVisible();
    await merchant.page.getByLabel("Restaurant name").fill(secondRestaurantName);
    await Promise.all([
      merchant.page.waitForURL("**/dashboard"),
      merchant.page.getByRole("button", { name: "Create restaurant" }).click(),
    ]);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    await expect(merchant.page.getByRole("heading", { name: "Choose active restaurant" })).toBeVisible();

    const allOptions = merchant.page.locator("#restaurantID option");
    await expect(allOptions).toHaveCount(2);
    const firstOptionText = await allOptions.nth(0).textContent();
    const secondOptionText = await allOptions.nth(1).textContent();
    expect(`${firstOptionText}${secondOptionText}`).toContain(firstRestaurantName);
    expect(`${firstOptionText}${secondOptionText}`).toContain(secondRestaurantName);

    const optionCount = await allOptions.count();
    let selectedValue: string | null = null;
    for (let i = 0; i < optionCount; i += 1) {
      const option = allOptions.nth(i);
      const text = (await option.textContent()) ?? "";
      if (text.includes(secondRestaurantName)) {
        selectedValue = await option.getAttribute("value");
        break;
      }
    }
    expect(selectedValue).toBeTruthy();

    await merchant.page.selectOption("#restaurantID", selectedValue ?? "");
    await Promise.all([
      merchant.page.waitForURL("**/dashboard"),
      merchant.page.getByRole("button", { name: "Use this restaurant" }).click(),
    ]);
    await expect(merchant.page).toHaveURL(/\/dashboard$/);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    const selectedOption = merchant.page.locator("#restaurantID option[selected]");
    await expect(selectedOption).toContainText(secondRestaurantName);

    await context.close();
  });
});
