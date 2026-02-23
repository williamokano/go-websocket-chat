import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ConnectionStatus } from "./index";

describe("ConnectionStatus", () => {
  it("shows green dot and 'Connected' when status is connected", () => {
    render(<ConnectionStatus status="connected" onlineCount={3} />);
    expect(screen.getByText("Connected")).toBeInTheDocument();
    expect(screen.getByText("3 online", { exact: false })).toBeInTheDocument();
  });

  it("shows 'Connecting...' when status is connecting", () => {
    render(<ConnectionStatus status="connecting" onlineCount={0} />);
    expect(screen.getByText("Connecting...")).toBeInTheDocument();
  });

  it("shows 'Disconnected' when status is disconnected", () => {
    render(<ConnectionStatus status="disconnected" onlineCount={0} />);
    expect(screen.getByText("Disconnected")).toBeInTheDocument();
  });

  it("does not show online count when disconnected", () => {
    render(<ConnectionStatus status="disconnected" onlineCount={5} />);
    expect(screen.queryByText("online", { exact: false })).not.toBeInTheDocument();
  });
});
