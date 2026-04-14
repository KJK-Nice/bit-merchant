import type { Question } from "../actors/Actor";

export const CurrentURL = (): Question<URL> => {
  return async (actor) => new URL(actor.page.url());
};
