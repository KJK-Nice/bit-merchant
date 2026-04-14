import type { Question } from "../actors/Actor";

export const ExtractOrderNumberFromURL = (): Question<string> => {
  return async (actor) => {
    const path = new URL(actor.page.url()).pathname;
    const match = path.match(/^\/order\/([^/]+)$/);
    if (!match) {
      throw new Error(`current path is not an order status page: ${path}`);
    }
    return match[1];
  };
};
