import type { BrowserContext } from "@playwright/test";
import { UseVirtualAuthenticator } from "../abilities/UseVirtualAuthenticator";
import { Actor } from "./Actor";

export const MerchantActor = async (context: BrowserContext): Promise<Actor> => {
  const page = await context.newPage();
  await UseVirtualAuthenticator(page);
  return new Actor("Merchant", page, context, "merchant");
};
