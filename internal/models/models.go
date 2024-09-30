package models

type UserRequest struct {
	Login    string
	Password string
}

type WIthdrawRequest struct {
	Order string
	Sum   float64
}

type BalanceGet struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type AccrualRequest struct {
	Order   string
	Status  string
	Accrual float64 `json:"accrual,omitempty"`
}

type OrderGet struct {
	Order       string  `json:"number"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual,omitempty"`
	ProcessedAt string  `json:"uploaded_at"`
}

type WithdrawalGet struct {
	Order       string  `json:"order"`
	Withdrawal  float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
