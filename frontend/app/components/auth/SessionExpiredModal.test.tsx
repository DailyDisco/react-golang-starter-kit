import { act, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { SessionExpiredModal } from "./SessionExpiredModal";

const authState = vi.hoisted(() => ({ hasValidatedSession: false }));

vi.mock("../../stores/auth-store", () => ({
  useAuthStore: { getState: () => authState },
}));

vi.mock("react-i18next", () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

function expire(detail?: { fromApiRequest: boolean }) {
  act(() => {
    window.dispatchEvent(new CustomEvent("session-expired", detail ? { detail } : undefined));
  });
}

describe("SessionExpiredModal", () => {
  beforeEach(() => {
    authState.hasValidatedSession = false;
  });

  it("shows for heartbeat or cross-tab expiry, which log out before dispatching", () => {
    render(<SessionExpiredModal />);

    expire();

    expect(screen.getByText("session.expired")).toBeInTheDocument();
  });

  it("ignores a 401 from an API request when the tab never had a validated session", () => {
    render(<SessionExpiredModal />);

    expire({ fromApiRequest: true });

    expect(screen.queryByText("session.expired")).not.toBeInTheDocument();
  });

  it("shows for a 401 from an API request during a validated session", () => {
    authState.hasValidatedSession = true;
    render(<SessionExpiredModal />);

    expire({ fromApiRequest: true });

    expect(screen.getByText("session.expired")).toBeInTheDocument();
  });
});
