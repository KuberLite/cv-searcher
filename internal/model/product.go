package model

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProductEvent struct {
	Action  string  `json:"action"`
	Payload Product `json:"payload"`
}
