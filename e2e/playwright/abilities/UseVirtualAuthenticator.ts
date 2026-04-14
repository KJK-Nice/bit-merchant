import type { Page } from "@playwright/test";

type CDPSession = {
  send<T = unknown>(method: string, params?: Record<string, unknown>): Promise<T>;
};

type CDPContext = {
  newCDPSession(page: Page): Promise<CDPSession>;
};

export const UseVirtualAuthenticator = async (page: Page): Promise<void> => {
  const cdpContext = page.context() as unknown as CDPContext;
  const session = await cdpContext.newCDPSession(page);

  await session.send("WebAuthn.enable");
  await session.send("WebAuthn.addVirtualAuthenticator", {
    options: {
      protocol: "ctap2",
      transport: "internal",
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
      automaticPresenceSimulation: true,
    },
  });
};
