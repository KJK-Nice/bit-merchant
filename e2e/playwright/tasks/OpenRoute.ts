import type { Task } from "../actors/Actor";
import { urlFor } from "../abilities/BrowseTheWeb";
import type { SurfaceName } from "../support/surfaces";

export const OpenRoute = (surface: SurfaceName, path: string): Task => {
  return async (actor) => {
    const targetURL = urlFor(surface, path);
    try {
      await actor.page.goto(targetURL, { waitUntil: "domcontentloaded" });
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      if (!msg.includes("net::ERR_FAILED")) {
        throw err;
      }
      await actor.page.waitForTimeout(150);
      await actor.page.goto(targetURL, { waitUntil: "domcontentloaded" });
    }
  };
};
