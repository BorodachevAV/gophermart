package customerrors

type AccrualRateLimitError struct {
}

func (r *AccrualRateLimitError) Error() string {
	return "Accrual rate limit error"
}
