import type { components } from "../../lib/backend-api/openapi/types";

export type Todo = components["schemas"]["Todo"];
export type Problem = components["schemas"]["Problem"];

export class APIProblemError extends Error {
  constructor(
    public readonly status: number,
    public readonly problem: Problem,
  ) {
    super(problem.detail ?? problem.title);
    this.name = "APIProblemError";
  }
}

export type TodoAPI = {
  listTodos: () => Promise<Todo[]>;
  createTodo: (title: string) => Promise<Todo>;
  updateTodo: (todo: Todo, completed: boolean) => Promise<Todo>;
  deleteTodo: (id: string) => Promise<void>;
};

const apiBaseURL = import.meta.env.VITE_BACKEND_API_BASE_URL ?? "http://127.0.0.1:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseURL}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init,
  });
  if (!response.ok) {
    const contentType = response.headers.get("Content-Type")?.toLowerCase() ?? "";
    if (contentType.includes("application/problem+json")) {
      const problem = (await response.json()) as Problem;
      throw new APIProblemError(response.status, problem);
    }
    throw new Error(`API request failed with ${response.status}`);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export const todoAPI: TodoAPI = {
  listTodos: () => request<Todo[]>("/todos"),
  createTodo: (title) => request<Todo>("/todos", { method: "POST", body: JSON.stringify({ title }) }),
  updateTodo: (todo, completed) =>
    request<Todo>(`/todos/${todo.id}`, {
      method: "PATCH",
      body: JSON.stringify({ completed, version: todo.version }),
    }),
  deleteTodo: (id) => request<void>(`/todos/${id}`, { method: "DELETE" }),
};
