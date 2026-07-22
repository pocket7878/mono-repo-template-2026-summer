import { afterEach, describe, expect, it, vi } from "vitest";

import { todoAPI } from "./api";

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

  it("throws for an unsuccessful response", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response(null, { status: 409 }));

    await expect(todoAPI.listTodos()).rejects.toThrow("409");
  });
});
