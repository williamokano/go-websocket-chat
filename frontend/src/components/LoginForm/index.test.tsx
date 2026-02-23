import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import { LoginForm } from "./index";

vi.mock("../../hooks/useAuth", () => ({
  useAuth: () => ({
    login: vi.fn().mockResolvedValue(undefined),
  }),
}));

describe("LoginForm", () => {
  it("renders username and password fields", () => {
    render(<LoginForm onSwitchToRegister={vi.fn()} />);

    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
  });

  it("renders a sign in button", () => {
    render(<LoginForm onSwitchToRegister={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
  });

  it("shows Register link", () => {
    render(<LoginForm onSwitchToRegister={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Register" })).toBeInTheDocument();
  });

  it("calls onSwitchToRegister when Register is clicked", async () => {
    const user = userEvent.setup();
    const onSwitch = vi.fn();

    render(<LoginForm onSwitchToRegister={onSwitch} />);
    await user.click(screen.getByRole("button", { name: "Register" }));

    expect(onSwitch).toHaveBeenCalledOnce();
  });
});
