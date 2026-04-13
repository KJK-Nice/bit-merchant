import type { Task, Question } from "../actors/Actor";

export const Remember = <T>(key: string, value: T): Task => {
  return async (actor) => {
    actor.remember(key, value);
  };
};

export const Recall = <T>(key: string): Question<T> => {
  return async (actor) => actor.recall<T>(key);
};
