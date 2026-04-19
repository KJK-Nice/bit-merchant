import { expect, test } from "@playwright/test";
import { CustomerActor } from "../actors/CustomerActor";
import { OpenRoute } from "../tasks/OpenRoute";

test.describe("Cart CTA reset", () => {
  test("Add to Cart reappears after decrementing qty to zero", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));

    // Grab the first available item card's CTA area.
    const firstCard = customer.page.locator('[data-show]').filter({ hasText: /add to cart/i }).first();
    const addBtn = firstCard.getByRole("button", { name: /add to cart/i });
    const decrementBtn = customer.page.getByRole("button", { name: /decrease quantity/i }).first();

    // Confirm Add to Cart is visible before interaction.
    await expect(addBtn).toBeVisible();

    // Click Add to Cart → stepper should appear, Add to Cart should hide.
    await addBtn.click();
    await expect(decrementBtn).toBeVisible({ timeout: 5_000 });
    await expect(addBtn).toBeHidden({ timeout: 5_000 });

    // Decrement back to zero → Add to Cart should reappear.
    await decrementBtn.click();
    await expect(addBtn).toBeVisible({ timeout: 5_000 });
    await expect(decrementBtn).toBeHidden({ timeout: 5_000 });

    await context.close();
  });

  test("Add to Cart reappears after incrementing then decrementing to zero", async ({ browser }) => {
    const context = await browser.newContext();
    const customer = await CustomerActor(context);

    await customer.attemptsTo(OpenRoute("customer", "/menu?restaurantID=restaurant_1"));

    const addBtn = customer.page.getByRole("button", { name: /add to cart/i }).first();
    const incrementBtn = customer.page.getByRole("button", { name: /increase quantity/i }).first();
    const decrementBtn = customer.page.getByRole("button", { name: /decrease quantity/i }).first();

    // Add item twice.
    await addBtn.click();
    await expect(incrementBtn).toBeVisible({ timeout: 5_000 });
    await incrementBtn.click();

    // Qty counter should show 2.
    const counter = customer.page.locator('[data-text*="cartItemQty"]').first();
    await expect(counter).toHaveText("2", { timeout: 5_000 });

    // Decrement twice → CTA resets.
    await decrementBtn.click();
    await expect(counter).toHaveText("1", { timeout: 5_000 });
    await decrementBtn.click();
    await expect(addBtn).toBeVisible({ timeout: 5_000 });
    await expect(decrementBtn).toBeHidden({ timeout: 5_000 });

    await context.close();
  });
});
