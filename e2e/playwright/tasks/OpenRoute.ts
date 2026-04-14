import type { Task } from "../actors/Actor";
import { urlFor } from "../abilities/BrowseTheWeb";
import type { SurfaceName } from "../support/surfaces";

export const OpenRoute = (surface: SurfaceName, path: string): Task => {
  return async (actor) => {
    await actor.page.goto(urlFor(surface, path), { waitUntil: "domcontentloaded" });
  };
};
