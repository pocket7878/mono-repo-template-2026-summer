import { Navigate, Route, Routes } from "react-router-dom";

import { TodoPage } from "./features/todos/TodoPage";

export function App() {
  return (
    <Routes>
      <Route element={<TodoPage />} path="/todos" />
      <Route element={<Navigate replace to="/todos" />} path="*" />
    </Routes>
  );
}
