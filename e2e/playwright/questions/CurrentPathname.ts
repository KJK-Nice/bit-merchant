import type { Question } from "../actors/Actor";

export const CurrentPathname = (): Question<string> => {
  return async (actor) => new URL(actor.page.url()).pathname;
};
