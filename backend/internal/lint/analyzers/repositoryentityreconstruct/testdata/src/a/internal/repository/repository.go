package repository

import "a/internal/entity"

func allowed() {
	_ = entity.ReconstructTodo()
	_ = entity.NormalizeTitle()
}

func forbidden() {
	_ = entity.NewTodo() // want "repository layer must use entity\\.Reconstruct\\* for entity construction: entity\\.NewTodo is forbidden"
	_ = entity.Todo{}    // want "repository layer must not instantiate entity types directly: entity composite literal is forbidden; use entity\\.Reconstruct\\*"
	_ = &entity.Todo{}   // want "repository layer must not instantiate entity types directly: entity composite literal is forbidden; use entity\\.Reconstruct\\*"
	_ = new(entity.Todo) // want "repository layer must not instantiate entity types directly: new\\(entity\\.\\*\\) is forbidden; use entity\\.Reconstruct\\*"
}
