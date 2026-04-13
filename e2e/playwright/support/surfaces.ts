export const surfaces = {
  public: "http://localhost:8080",
  customer: "http://order.localhost:8080",
  merchant: "http://merchant.localhost:8080",
} as const;

export type SurfaceName = keyof typeof surfaces;
