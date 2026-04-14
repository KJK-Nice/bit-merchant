import { surfaces, type SurfaceName } from "../support/surfaces";

export const urlFor = (surface: SurfaceName, path: string): string => {
  const base = surfaces[surface];
  return new URL(path, base).toString();
};
