import type { Task } from "../actors/Actor";
import { urlFor } from "../abilities/BrowseTheWeb";
import type { SurfaceName } from "../support/surfaces";

export const OpenRoute = (surface: SurfaceName, path: string): Task => {
  return async (actor) => {
    const targetURL = urlFor(surface, path);
    const retryableError = (msg: string): boolean =>
      /net::ERR_(FAILED|ABORTED|CONNECTION_RESET|CONNECTION_CLOSED|CONNECTION_REFUSED)|timeout/i.test(msg);

    let lastError: unknown;
    for (let attempt = 0; attempt < 4; attempt += 1) {
      try {
        await actor.page.goto(targetURL, { waitUntil: "commit", timeout: 15_000 });
        await actor.page.waitForLoadState("domcontentloaded", { timeout: 15_000 });
        return;
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        if (!retryableError(msg)) {
          throw err;
        }
        lastError = err;

        // Occasionally Playwright reports an aborted navigation even when the target host has loaded.
        const currentURL = actor.page.url();
        if (currentURL && currentURL !== "about:blank") {
          try {
            if (new URL(currentURL).host === new URL(targetURL).host) {
              return;
            }
          } catch {
            // Keep retrying on parse errors.
          }
        }

        await actor.page.waitForTimeout(150 * (attempt + 1));
      }
    }

    throw lastError;
  };
};
