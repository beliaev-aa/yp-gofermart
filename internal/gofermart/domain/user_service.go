package domain

// UserBalance - представляет баланс пользователя, включая текущий баланс и сумму снятий
type UserBalance struct {
	Current   float64
	Withdrawn float64
}
