package repository

import "a/internal/entity"

func testsMayBuildFixtures() {
	_ = entity.NewTodo()
	_ = entity.Todo{}
}
