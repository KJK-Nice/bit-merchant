import type { Task } from "../actors/Actor";

export const Logout = (): Task => {
  return async (actor) => {
    await actor.page.evaluate(() => {
      const form = document.getElementById("layout-logout-form") as HTMLFormElement | null;
      if (!form) throw new Error("logout form not found");
      form.submit();
    });
    await actor.page.waitForURL("**/");
  };
};
