import { expect, test } from "@playwright/test";
import { MerchantActor } from "../actors/MerchantActor";
import { LoginWithPasskey } from "../tasks/LoginWithPasskey";
import { Logout } from "../tasks/Logout";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

const escapeRegExp = (value: string): string => value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

test.describe("Merchant multi-restaurant selection", () => {
  test.skip("multi-membership user is prompted to select restaurant after login", async ({ browser }) => {
    // TODO: Depends on passkey re-login, currently stalls at WebAuthn assertion in Playwright.
    const context = await browser.newContext();
    const merchant = await MerchantActor(context);

    const suffix = Date.now().toString();
    const firstRestaurantName = `First Cafe ${suffix}`;
    const secondRestaurantName = `Second Cafe ${suffix}`;

    await merchant.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, firstRestaurantName));
    await expect(merchant.page).toHaveURL(/\/dashboard$/);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/restaurants/new"));
    await merchant.page.getByLabel("Restaurant name").fill(secondRestaurantName);
    await Promise.all([
      merchant.page.waitForURL("**/dashboard", { waitUntil: "domcontentloaded" }),
      merchant.page.getByRole("button", { name: "Create restaurant" }).click(),
    ]);

    await merchant.attemptsTo(Logout(), LoginWithPasskey());
    await expect(merchant.page).toHaveURL(/\/auth\/select-restaurant$/);
    await expect(merchant.page.getByRole("heading", { name: "Choose active restaurant" })).toBeVisible();

    await context.close();
  });

  test("owner creating a new restaurant switches active context to the new restaurant", async ({ browser }) => {
    const context = await browser.newContext();
    const merchant = await MerchantActor(context);

    const suffix = `${Date.now()}-active`;
    const firstRestaurantName = `First Cafe ${suffix}`;
    const secondRestaurantName = `Second Cafe ${suffix}`;

    await merchant.attemptsTo(RegisterOwnerWithPasskey(`Owner ${suffix}`, firstRestaurantName));
    await expect(merchant.page).toHaveURL(/\/dashboard$/);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/restaurants/new"));
    await merchant.page.getByLabel("Restaurant name").fill(secondRestaurantName);
    await Promise.all([
      merchant.page.waitForURL("**/dashboard", { waitUntil: "domcontentloaded" }),
      merchant.page.getByRole("button", { name: "Create restaurant" }).click(),
    ]);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    const selectedLabel = merchant.page.locator("button#restaurantID .select-value");
    await expect(selectedLabel).toContainText(secondRestaurantName);

    await context.close();
  });

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
      merchant.page.waitForURL("**/dashboard", { waitUntil: "domcontentloaded" }),
      merchant.page.getByRole("button", { name: "Create restaurant" }).click(),
    ]);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    await expect(merchant.page.getByRole("heading", { name: "Choose active restaurant" })).toBeVisible();

    const listbox = merchant.page.locator("[data-tui-selectbox-content='true']");
    const allOptions = listbox.locator("[data-tui-selectbox-value]");
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
      if (new RegExp(escapeRegExp(secondRestaurantName), "i").test(text)) {
        selectedValue = await option.getAttribute("data-tui-selectbox-value");
        break;
      }
    }
    expect(selectedValue).toBeTruthy();

    const hiddenInput = merchant.page.locator("button#restaurantID input[data-tui-selectbox-hidden-input]");
    await hiddenInput.evaluate(
      (node, value) => {
        const input = node as HTMLInputElement;
        input.value = value ?? "";
      },
      selectedValue,
    );
    await Promise.all([
      merchant.page.waitForURL("**/dashboard", { waitUntil: "domcontentloaded" }),
      merchant.page.getByRole("button", { name: "Use this restaurant" }).click(),
    ]);
    await expect(merchant.page).toHaveURL(/\/dashboard$/);

    await merchant.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    const selectedLabel = merchant.page.locator("button#restaurantID .select-value");
    await expect(selectedLabel).toContainText(secondRestaurantName);

    await context.close();
  });
});
