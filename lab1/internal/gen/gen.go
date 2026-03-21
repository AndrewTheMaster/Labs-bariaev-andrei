package gen

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	mathrand "math/rand"
)

// NewDeterministicSource создаёт детерминированный источник случайности по заданному seed.
func NewDeterministicSource(seed int64) *mathrand.Rand {
	return mathrand.New(mathrand.NewSource(seed))
}

// RandomBytes генерирует слайс случайных байт заданной длины, используя криптографический rand.
// Подходит для генерации ключей/значений в тестах.
func RandomBytes(n int) []byte {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		// Падаем максимально громко — это вспомогательная функция для тестов.
		panic(err)
	}
	return buf
}

// RandomUint64 генерирует случайное uint64 с использованием crypto/rand.
func RandomUint64() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint64(b[:])
}

// RandomFloat64 возвращает случайное число в [0,1) на основе math/rand.
func RandomFloat64(r *mathrand.Rand) float64 {
	return r.Float64() * (1.0 - math.SmallestNonzeroFloat64)
}

