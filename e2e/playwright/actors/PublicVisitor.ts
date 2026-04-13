import type { BrowserContext } from "@playwright/test";
import { Actor } from "./Actor";

export const PublicVisitor = async (context: BrowserContext): Promise<Actor> => {
  const page = await context.newPage();
  return new Actor("PublicVisitor", page, context, "public");
};
