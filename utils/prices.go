package utils

func ToBigInt(f float64) int64 {
	// Multiply by 10^10 and cast to int64
	return int64(f * 1e9)
}

func ToFloat(i int64) float64 {
	// Divide by 10^10 and cast to float64
	return float64(i) / 1e9
}
