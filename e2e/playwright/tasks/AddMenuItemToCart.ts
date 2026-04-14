import type { Task } from "../actors/Actor";

export const AddMenuItemToCart = (): Task => {
  return async (actor) => {
    await actor.page.getByRole("button", { name: "Add to Cart" }).first().click();
  };
};
