import type { Task } from "../actors/Actor";
import { OpenRoute } from "./OpenRoute";

export const RegisterOwnerWithPasskey = (displayName: string, restaurantName: string): Task => {
  return async (actor) => {
    await OpenRoute("merchant", "/auth/signup")(actor);
    await actor.page.getByLabel("Your name").fill(displayName);
    await actor.page.getByLabel("Restaurant name").fill(restaurantName);
    const submit = actor.page.getByRole("button", { name: "Create with passkey" });
    await Promise.all([actor.page.waitForURL("**/dashboard", { waitUntil: "domcontentloaded" }), submit.click()]);
  };
};
