import type { Task } from "../actors/Actor";

export const PlaceCashOrder = (customerName = "Test"): Task => {
  return async (actor) => {
    await actor.page.getByLabel("Name for pickup").fill(customerName);
    await actor.page.getByRole("button", { name: /Send to kitchen/ }).click();
  };
};
