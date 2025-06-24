package services

import (
	"math/rand/v2"
	"time"
	"validation-service/internal/dto"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(*dto.Invoice) bool {
	time.Sleep(time.Duration(rand.Int()%5_000) * time.Millisecond) // some validation logic
	const approveProbability float64 = 0.9
	return rand.Float64() < approveProbability
}
