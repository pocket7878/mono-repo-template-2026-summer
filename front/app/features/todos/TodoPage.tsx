import { FormEvent, useEffect, useState } from "react";

import type { Todo, TodoAPI } from "./api";
import { todoAPI } from "./api";

type TodoPageProps = {
  api?: TodoAPI;
};

export function TodoPage({ api = todoAPI }: TodoPageProps) {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [title, setTitle] = useState("");
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    void api
      .listTodos()
      .then(setTodos)
      .catch(() => setMessage("Could not load todos."));
  }, [api]);

  async function createTodo(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      setMessage("Title is required.");
      return;
    }
    try {
      const created = await api.createTodo(trimmedTitle);
      setTodos((current) => [...current, created]);
      setTitle("");
      setMessage(null);
    } catch {
      setMessage("Could not create todo.");
    }
  }

  async function toggleTodo(todo: Todo) {
    try {
      const updated = await api.updateTodo(todo, !todo.completed);
      setTodos((current) => current.map((item) => (item.id === updated.id ? updated : item)));
      setMessage(null);
    } catch {
      setMessage("This todo changed elsewhere. Reload to get the latest version.");
    }
  }

  async function deleteTodo(id: string) {
    try {
      await api.deleteTodo(id);
      setTodos((current) => current.filter((item) => item.id !== id));
    } catch {
      setMessage("Could not delete todo.");
    }
  }

  return (
    <main>
      <h1>Todos</h1>
      <form onSubmit={createTodo}>
        <label htmlFor="todo-title">Title</label>
        <input id="todo-title" value={title} onChange={(event) => setTitle(event.target.value)} />
        <button type="submit">Add todo</button>
      </form>
      {message ? <p role="alert">{message}</p> : null}
      <ul>
        {todos.map((todo) => (
          <li key={todo.id}>
            <label>
              <input
                aria-label={todo.title}
                checked={todo.completed}
                onChange={() => void toggleTodo(todo)}
                type="checkbox"
              />
              <span>{todo.title}</span>
            </label>
            <span>{todo.completed ? "Completed" : "Open"}</span>
            <button onClick={() => void deleteTodo(todo.id)} type="button">
              Delete
            </button>
          </li>
        ))}
      </ul>
    </main>
  );
}
