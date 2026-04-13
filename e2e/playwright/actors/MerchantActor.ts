import type { BrowserContext } from "@playwright/test";
import { Actor } from "./Actor";

export const MerchantActor = async (context: BrowserContext): Promise<Actor> => {
  const page = await context.newPage();
  return new Actor("Merchant", page, context, "merchant");
};
