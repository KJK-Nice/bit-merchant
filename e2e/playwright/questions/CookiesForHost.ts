import type { Question } from "../actors/Actor";
import { urlFor } from "../abilities/BrowseTheWeb";
import type { SurfaceName } from "../support/surfaces";

export const CookiesForHost = (surface: SurfaceName): Question<string[]> => {
  return async (actor) => {
    const cookies = await actor.context.cookies([urlFor(surface, "/")]);
    return cookies.map((c) => c.name);
  };
};
