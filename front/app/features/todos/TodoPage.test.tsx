import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { TodoPage } from "./TodoPage";
import { APIProblemError } from "./api";

describe("TodoPage", () => {
  it("creates and displays a todo", async () => {
    const user = userEvent.setup();
    const createTodo = vi.fn().mockResolvedValue({
      id: "todo-1",
      title: "Buy milk",
      completed: false,
      version: 1,
      createdAt: "2026-07-22T12:00:00Z",
      updatedAt: "2026-07-22T12:00:00Z",
    });

    render(
      <TodoPage
        api={{ listTodos: vi.fn().mockResolvedValue([]), createTodo, updateTodo: vi.fn(), deleteTodo: vi.fn() }}
      />,
    );

    await user.type(screen.getByLabelText("Title"), "Buy milk");
    await user.click(screen.getByRole("button", { name: "Add todo" }));

    expect(await screen.findByText("Buy milk")).toBeVisible();
    expect(createTodo).toHaveBeenCalledWith("Buy milk");
  });

  it("rejects a blank title before calling the API", async () => {
    const user = userEvent.setup();
    const createTodo = vi.fn();

    render(
      <TodoPage
        api={{ listTodos: vi.fn().mockResolvedValue([]), createTodo, updateTodo: vi.fn(), deleteTodo: vi.fn() }}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Add todo" }));

    expect(screen.getByRole("alert")).toHaveTextContent("Title is required.");
    expect(createTodo).not.toHaveBeenCalled();
  });

  it("shows a conflict recovery message when completion update fails", async () => {
    const user = userEvent.setup();
    const todo = {
      id: "todo-1",
      title: "Buy milk",
      completed: false,
      version: 1,
      createdAt: "2026-07-22T12:00:00Z",
      updatedAt: "2026-07-22T12:00:00Z",
    };

    render(
      <TodoPage
        api={{
          listTodos: vi.fn().mockResolvedValue([todo]),
          createTodo: vi.fn(),
          updateTodo: vi.fn().mockRejectedValue(
            new APIProblemError(409, {
              type: "https://reference-todo.dev/problems/version-conflict",
              title: "Todo was modified",
              status: 409,
            }),
          ),
          deleteTodo: vi.fn(),
        }}
      />,
    );

    await user.click(await screen.findByRole("checkbox", { name: "Buy milk" }));

    expect(screen.getByRole("alert")).toHaveTextContent("This todo changed elsewhere.");
  });

  it("does not describe an unexpected update failure as a conflict", async () => {
    const user = userEvent.setup();
    const todo = {
      id: "todo-1",
      title: "Buy milk",
      completed: false,
      version: 1,
      createdAt: "2026-07-22T12:00:00Z",
      updatedAt: "2026-07-22T12:00:00Z",
    };

    render(
      <TodoPage
        api={{
          listTodos: vi.fn().mockResolvedValue([todo]),
          createTodo: vi.fn(),
          updateTodo: vi.fn().mockRejectedValue(new Error("network failure")),
          deleteTodo: vi.fn(),
        }}
      />,
    );

    await user.click(await screen.findByRole("checkbox", { name: "Buy milk" }));

    expect(screen.getByRole("alert")).toHaveTextContent("Could not update todo.");
  });

  it("deletes a todo", async () => {
    const user = userEvent.setup();
    const todo = {
      id: "todo-1",
      title: "Buy milk",
      completed: false,
      version: 1,
      createdAt: "2026-07-22T12:00:00Z",
      updatedAt: "2026-07-22T12:00:00Z",
    };
    const deleteTodo = vi.fn().mockResolvedValue(undefined);

    render(
      <TodoPage
        api={{ listTodos: vi.fn().mockResolvedValue([todo]), createTodo: vi.fn(), updateTodo: vi.fn(), deleteTodo }}
      />,
    );

    await user.click(await screen.findByRole("button", { name: "Delete" }));

    expect(deleteTodo).toHaveBeenCalledWith("todo-1");
    expect(screen.queryByText("Buy milk")).not.toBeInTheDocument();
  });
});
