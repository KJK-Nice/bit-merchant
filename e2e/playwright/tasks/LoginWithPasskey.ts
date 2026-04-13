import type { Task } from "../actors/Actor";
import { OpenRoute } from "./OpenRoute";

export const LoginWithPasskey = (): Task => {
  return async (actor) => {
    await OpenRoute("merchant", "/auth/login")(actor);
    const signIn = actor.page.getByRole("button", { name: "Sign in" });
    await Promise.all([actor.page.waitForURL("**/dashboard"), signIn.click()]);
  };
};
