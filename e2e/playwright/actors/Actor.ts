import type { BrowserContext, Page } from "@playwright/test";
import type { SurfaceName } from "../support/surfaces";

export type Task = (actor: Actor) => Promise<void>;
export type Question<T> = (actor: Actor) => Promise<T>;

export class Actor {
  constructor(
    public readonly name: string,
    public readonly page: Page,
    public readonly context: BrowserContext,
    public readonly defaultSurface: SurfaceName,
  ) {}

  async attemptsTo(...tasks: Task[]): Promise<void> {
    for (const task of tasks) {
      await task(this);
    }
  }

  async asks<T>(question: Question<T>): Promise<T> {
    return question(this);
  }
}
