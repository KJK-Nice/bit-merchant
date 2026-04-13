import type { BrowserContext } from "@playwright/test";
import { Actor } from "./Actor";

export const CustomerActor = async (context: BrowserContext): Promise<Actor> => {
  const page = await context.newPage();
  return new Actor("Customer", page, context, "customer");
};
