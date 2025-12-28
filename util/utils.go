package util

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strconv"
)

func DeduplicateSlice[T comparable](input []T) []T {
	unique := make([]T, 0)
	occurrenceMap := make(map[T]struct{})

	for _, val := range input {
		if _, ok := occurrenceMap[val]; !ok {
			occurrenceMap[val] = struct{}{}
			unique = append(unique, val)
		}
	}

	return unique
}

func GenerateRandomNumericCode(length int) (string, error) {
	var otpCharset = "0123456789"
	otp := ""
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(otpCharset))))
		if err != nil {
			return "", err
		}
		otp += string(otpCharset[num.Int64()])
	}

	return otp, nil
}

func IsValidEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func ParseUint(val string) (uint64, error) {
	return strconv.ParseUint(val, 10, 64)
}
