import { render, screen } from "@testing-library/react";
import React from "react";
import { describe, expect, it } from "vitest";
import { EmptyState, ErrorState, LoadingState } from "./ui";

describe("dashboard UI states", () => {
  it("renders loading state", () => {
    render(React.createElement(LoadingState, { label: "Loading repositories" }));
    expect(screen.getByText("Loading repositories...")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    render(React.createElement(EmptyState, { title: "No repositories found.", body: "Run a scan." }));
    expect(screen.getByText("No repositories found.")).toBeInTheDocument();
    expect(screen.getByText("Run a scan.")).toBeInTheDocument();
  });

  it("renders error state", () => {
    render(React.createElement(ErrorState, { message: "Request failed" }));
    expect(screen.getByText("Request failed")).toBeInTheDocument();
  });
});
