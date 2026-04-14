import type { BrowserContext, Page } from "@playwright/test";
import type { SurfaceName } from "../support/surfaces";

export type Task = (actor: Actor) => Promise<void>;
export type Question<T> = (actor: Actor) => Promise<T>;

export class Actor {
  private readonly memory = new Map<string, unknown>();

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

  remember<T>(key: string, value: T): void {
    this.memory.set(key, value);
  }

  recall<T>(key: string): T {
    if (!this.memory.has(key)) {
      throw new Error(`${this.name} has no memory for key: ${key}`);
    }
    return this.memory.get(key) as T;
  }
}
