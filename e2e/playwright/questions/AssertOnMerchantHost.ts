import type { Question } from "../actors/Actor";

export const AssertOnMerchantHost = (): Question<boolean> => {
  return async (actor) => new URL(actor.page.url()).host === "merchant.localhost:8080";
};
