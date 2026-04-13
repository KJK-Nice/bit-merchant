import type { Task } from "../actors/Actor";

export const ProceedToCheckout = (): Task => {
  return async (actor) => {
    const checkoutLink = actor.page.getByRole("link", { name: /view cart|checkout/i }).first();
    await checkoutLink.waitFor({ state: "visible" });
    await Promise.all([actor.page.waitForURL("**/order/confirm"), checkoutLink.click()]);
  };
};
