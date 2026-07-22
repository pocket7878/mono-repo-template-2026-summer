package entity

type Todo struct{}

func NewTodo() *Todo { return &Todo{} }

func ReconstructTodo() *Todo { return &Todo{} }

func NormalizeTitle() string { return "normalized" }
