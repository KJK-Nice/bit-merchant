import type { Question } from "../actors/Actor";

export const AssertPath = (expectedPath: string): Question<boolean> => {
  return async (actor) => new URL(actor.page.url()).pathname === expectedPath;
};
