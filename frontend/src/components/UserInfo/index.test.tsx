import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import { UserInfo } from "./index";

describe("UserInfo", () => {
  const user = { id: "u1", username: "alice" };

  it("renders the username", () => {
    render(<UserInfo user={user} onLogout={vi.fn()} />);

    expect(screen.getByText("alice")).toBeInTheDocument();
  });

  it("renders a sign out button", () => {
    render(<UserInfo user={user} onLogout={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
  });

  it("calls onLogout when sign out is clicked", async () => {
    const userEvent_ = userEvent.setup();
    const onLogout = vi.fn();

    render(<UserInfo user={user} onLogout={onLogout} />);
    await userEvent_.click(screen.getByRole("button", { name: "Sign out" }));

    expect(onLogout).toHaveBeenCalledOnce();
  });
});
