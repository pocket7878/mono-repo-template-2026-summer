import "@typespec/http";
import "@typespec/openapi";

using Http;
using TypeSpec.OpenAPI;

@service(#{ title: "Reference Todo API" })
namespace ReferenceTodo;

model Todo {
  id: string;
  title: string;
  completed: boolean;
  version: int32;
  dueDate?: utcDateTime;
  createdAt: utcDateTime;
  updatedAt: utcDateTime;
}

model CreateTodoRequest {
  title: string;
  dueDate?: utcDateTime;
}

model UpdateTodoRequest {
  title?: string;
  completed?: boolean;
  dueDate?: utcDateTime;
  version: int32;
}

model Problem {
  type: string;
  title: string;
  status: int32;
  detail?: string;
}

@error
model BadRequestProblem extends Problem {
  @statusCode status: 400;
  @header contentType: "application/problem+json";
}

@error
model NotFoundProblem extends Problem {
  @statusCode status: 404;
  @header contentType: "application/problem+json";
}

@error
model ConflictProblem extends Problem {
  @statusCode status: 409;
  @header contentType: "application/problem+json";
}

@error
model InternalServerErrorProblem extends Problem {
  @statusCode status: 500;
  @header contentType: "application/problem+json";
}

@route("/todos")
@tag("Todos")
interface Todos {
  @get
  @operationId("TodosList")
  list(): Todo[] | InternalServerErrorProblem;

  @post
  @operationId("TodosCreate")
  create(@body body: CreateTodoRequest): Todo | BadRequestProblem | InternalServerErrorProblem;

  @patch
  @route("/{todoId}")
  @operationId("TodosUpdate")
  update(@path todoId: string, @body body: UpdateTodoRequest): Todo | BadRequestProblem | NotFoundProblem | ConflictProblem | InternalServerErrorProblem;

  @delete
  @route("/{todoId}")
  @operationId("TodosDelete")
  delete(@path todoId: string): void | NotFoundProblem | InternalServerErrorProblem;
}
