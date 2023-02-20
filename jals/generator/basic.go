package generator

import (
	"math/rand"
	"time"
)

// basic generator without collision check, to be replaced
func BasicGenerator(length int, useEmoji bool) string {
	rand.Seed(time.Now().UnixNano())

	const characters = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	generated := make([]byte, length)
	for i := range generated {
		generated[i] = characters[rand.Intn(len(characters))]
	}
	return string(generated)
}
