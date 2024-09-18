package models

type UserJSONRequest struct {
	Login    string
	Password string
}

type WIthdrawJSONRequest struct {
	Order string
	Sum   float64
}

type BalanceGetJSON struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type AccrualJSONRequest struct {
	Order   string
	Status  string
	Accrual float64 `json:"accrual,omitempty"`
}

type OrderGetJSON struct {
	Order       string  `json:"number"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual,omitempty"`
	ProcessedAt string  `json:"uploaded_at"`
}

type WithdrawalGetJSON struct {
	Order       string  `json:"order"`
	Withdrawal  float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
