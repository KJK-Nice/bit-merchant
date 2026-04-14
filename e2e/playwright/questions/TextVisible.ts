import type { Question } from "../actors/Actor";

export const TextVisible = (text: string): Question<boolean> => {
  return async (actor) => actor.page.getByText(text, { exact: false }).first().isVisible();
};
