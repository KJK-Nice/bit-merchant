import type { Task } from "../actors/Actor";

export const PlaceCashOrder = (): Task => {
  return async (actor) => {
    await actor.page.getByRole("button", { name: "Place Order" }).click();
  };
};
