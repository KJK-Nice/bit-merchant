import { expect, test } from "@playwright/test";
import { urlFor } from "../abilities/BrowseTheWeb";
import { CustomerActor } from "../actors/CustomerActor";
import { MerchantActor } from "../actors/MerchantActor";
import { OpenRoute } from "../tasks/OpenRoute";
import { RegisterOwnerWithPasskey } from "../tasks/RegisterOwnerWithPasskey";

const csrfTokenForMerchant = async (
  actor: Awaited<ReturnType<typeof MerchantActor>>,
): Promise<string> => {
  const cookies = await actor.context.cookies(urlFor("merchant", "/"));
  const csrfCookie = cookies.find((c) => c.name === "csrf");
  return csrfCookie?.value ?? "";
};

const selectedRestaurantID = async (
  actor: Awaited<ReturnType<typeof MerchantActor>>,
): Promise<string> => {
  const hidden = actor.page.locator(
    "button#restaurantID input[data-tui-selectbox-hidden-input]",
  );
  await expect(hidden).toHaveAttribute("value", /.+/);
  return (await hidden.getAttribute("value")) ?? "";
};

// seedCategoryAndItem creates one category + one item via the admin API and
// returns their IDs so the test can navigate to the editor.
const seedCategoryAndItem = async (
  actor: Awaited<ReturnType<typeof MerchantActor>>,
  csrf: string,
  suffix: string,
): Promise<{ categoryID: string; itemID: string }> => {
  return await actor.page.evaluate(
    async ({ csrfToken, tokenSuffix }) => {
      const headers = {
        "Content-Type": "application/x-www-form-urlencoded",
        "X-CSRF-Token": csrfToken,
      };
      const catBody = new URLSearchParams({
        csrf: csrfToken,
        name: `Bao ${tokenSuffix}`,
        displayOrder: "1",
      });
      const catResp = await fetch("/admin/category", {
        method: "POST",
        headers,
        body: catBody.toString(),
      });
      if (!catResp.ok) throw new Error(`create category: ${catResp.status}`);

      const dashboard = await fetch("/admin/dashboard");
      const html = await dashboard.text();
      const catMatch = html.match(/data-category-id=\"([^\"]+)\"/);
      if (!catMatch?.[1]) throw new Error("category id not found");
      const categoryID = catMatch[1];

      const itemBody = new URLSearchParams({
        csrf: csrfToken,
        categoryID,
        name: `Pork Belly Bao ${tokenSuffix}`,
        price: "6.50",
        description: "Slow-braised pork belly, hoisin glaze.",
        available: "on",
      });
      const itemResp = await fetch("/admin/item", {
        method: "POST",
        headers,
        body: itemBody.toString(),
      });
      if (!itemResp.ok) throw new Error(`create item: ${itemResp.status}`);

      const reloaded = await fetch("/admin/dashboard");
      const reloadedHtml = await reloaded.text();
      const itemMatch = reloadedHtml.match(/\/admin\/items\/([^/]+)\/edit/);
      if (!itemMatch?.[1]) throw new Error("item edit link not found");
      return { categoryID, itemID: itemMatch[1] };
    },
    { csrfToken: csrf, tokenSuffix: suffix },
  );
};

// publishWithModifiers posts the editor form populating badges, allergens,
// spice level, and a required "Sauce" modifier group with two options.
const publishWithModifiers = async (
  actor: Awaited<ReturnType<typeof MerchantActor>>,
  csrf: string,
  itemID: string,
  categoryID: string,
): Promise<void> => {
  const optionGroupsJSON = JSON.stringify([
    {
      id: "g_sauce",
      name: "Choose a sauce",
      required: true,
      min_selections: 1,
      max_selections: 1,
      default_option_id: "o_hoisin",
      options: [
        { id: "o_hoisin", name: "Original hoisin", price_delta: 0 },
        { id: "o_sriracha", name: "Sriracha mayo", price_delta: 0 },
      ],
    },
  ]);

  await actor.page.evaluate(
    async ({ csrfToken, itemID, categoryID, ogJSON }) => {
      const body = new URLSearchParams();
      body.set("csrf", csrfToken);
      body.set("name", "Pork Belly Bao");
      body.set("price", "6.50");
      body.set("description", "Slow-braised pork belly.");
      body.set("categoryID", categoryID);
      body.set("available", "on");
      body.set("spice_level", "MILD");
      body.set("schedule", "ALL_DAY");
      body.set("badges_csv", "Popular");
      body.set("allow_special_instructions", "on");
      body.set("option_groups_json", ogJSON);
      body.append("allergens", "Gluten");
      body.append("allergens", "Soy");

      const resp = await fetch(`/admin/items/${itemID}/edit`, {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          "X-CSRF-Token": csrfToken,
        },
        body: body.toString(),
        redirect: "manual",
      });
      // The server returns a 302 with flash=item_saved on success.
      if (resp.status !== 302 && resp.status !== 200) {
        throw new Error(`edit POST failed: ${resp.status}`);
      }
    },
    { csrfToken: csrf, itemID, categoryID, ogJSON: optionGroupsJSON },
  );
};

test.describe("Item editor — modifier flow", () => {
  test("owner sets modifiers; customer orders with note; kitchen ticket shows both", async ({
    browser,
  }) => {
    const ownerCtx = await browser.newContext();
    const owner = await MerchantActor(ownerCtx);

    const suffix = Date.now().toString();
    await owner.attemptsTo(
      RegisterOwnerWithPasskey(`Owner ${suffix}`, `Modifier Cafe ${suffix}`),
    );
    await expect(owner.page).toHaveURL(/\/dashboard$/);

    await owner.attemptsTo(OpenRoute("merchant", "/auth/select-restaurant"));
    const restaurantID = await selectedRestaurantID(owner);
    expect(restaurantID).toBeTruthy();

    const csrf = await csrfTokenForMerchant(owner);
    expect(csrf).not.toBe("");

    const { categoryID, itemID } = await seedCategoryAndItem(owner, csrf, suffix);
    await publishWithModifiers(owner, csrf, itemID, categoryID);

    // Customer opens the item-detail page and sees the new bits.
    const customerCtx = await browser.newContext();
    const customer = await CustomerActor(customerCtx);
    await customer.attemptsTo(
      OpenRoute(
        "customer",
        `/menu/item/${itemID}?restaurantID=${restaurantID}`,
      ),
    );

    await expect(customer.page.getByText("Popular", { exact: false })).toBeVisible();
    await expect(
      customer.page.getByText(/Contains: gluten, soy/i),
    ).toBeVisible();
    await expect(customer.page.getByText(/Mild/)).toBeVisible();

    // Pick a sauce (radio) and add a note.
    await customer.page.getByLabel("Sriracha mayo").check();
    await customer.page
      .getByPlaceholder(/no cilantro, please/i)
      .fill("no cilantro, please");

    await Promise.all([
      customer.page.waitForURL("**/menu**"),
      customer.page.getByRole("button", { name: /Add to cart/i }).click(),
    ]);

    // Proceed to checkout from the menu page.
    const checkoutLink = customer.page.getByRole("link", {
      name: /view cart|checkout/i,
    }).first();
    await checkoutLink.waitFor({ state: "visible" });
    await Promise.all([
      customer.page.waitForURL("**/order/confirm"),
      checkoutLink.click(),
    ]);

    // Confirm page shows the modifier + note.
    await expect(customer.page.getByText("Sriracha mayo")).toBeVisible();
    await expect(customer.page.getByText("no cilantro, please")).toBeVisible();

    await customer.page.getByLabel("Name for pickup").fill("E2E Diner");
    await Promise.all([
      customer.page.waitForURL(/\/order\/[^/]+$/),
      customer.page.getByRole("button", { name: /Send to kitchen/ }).click(),
    ]);

    // Owner switches to the kitchen board and sees the modifier + note on the
    // ticket. The kitchen card is the canonical place modifiers must appear.
    await owner.attemptsTo(OpenRoute("merchant", "/kitchen"));
    const ticket = owner.page.locator("div[id^='order-']").first();
    await expect(ticket).toBeVisible();
    await expect(ticket).toContainText("Sriracha mayo");
    await expect(ticket).toContainText("no cilantro, please");

    await ownerCtx.close();
    await customerCtx.close();
  });
});
