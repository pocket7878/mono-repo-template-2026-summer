package usecase

import "a/internal/entity"

func applicationCodeMayUseValidatedConstructors() {
	_ = entity.NewTodo()
}
