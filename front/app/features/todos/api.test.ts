import { afterEach, describe, expect, it, vi } from "vitest";

import { APIProblemError, todoAPI } from "./api";

describe("todoAPI", () => {
  afterEach(() => vi.restoreAllMocks());

  it("sends a typed create request", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          id: "todo-1",
          title: "Buy milk",
          completed: false,
          version: 1,
          createdAt: "2026-07-22T12:00:00Z",
          updatedAt: "2026-07-22T12:00:00Z",
        }),
        {
          status: 200,
        },
      ),
    );

    const todo = await todoAPI.createTodo("Buy milk");

    expect(todo.title).toBe("Buy milk");
    expect(fetchMock).toHaveBeenCalledWith(
      "http://127.0.0.1:8080/todos",
      expect.objectContaining({ method: "POST", body: JSON.stringify({ title: "Buy milk" }) }),
    );
  });

  it("preserves Problem Details for an unsuccessful response", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          type: "https://reference-todo.dev/problems/version-conflict",
          title: "Todo was modified",
          status: 409,
          detail: "Reload before updating.",
        }),
        { status: 409, headers: { "Content-Type": "application/problem+json" } },
      ),
    );

    const error = await todoAPI.listTodos().catch((reason: unknown) => reason);

    expect(error).toBeInstanceOf(APIProblemError);
    expect(error).toMatchObject({ status: 409, problem: { title: "Todo was modified" } });
  });

  it("falls back to a status-only error for a non-Problem response", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(null, { status: 503 }));

    await expect(todoAPI.listTodos()).rejects.toThrow("503");
  });
});
