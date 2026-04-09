package generator

import "fmt"

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Base62Encoder кодирует числа в короткие строки
type Base62Encoder struct{}

// NewBase62Encoder создаёт новый encoder
func NewBase62Encoder() *Base62Encoder {
	return &Base62Encoder{}
}

// Encode преобразует uint64 в строку base62
// Для числа 2^64-1 максимальная длина строки — 11 символов
func (e *Base62Encoder) Encode(n uint64) string {
	if n == 0 {
		return string(base62Alphabet[0])
	}

	var buf [11]byte
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = base62Alphabet[n%62]
		n /= 62
	}

	return string(buf[i:])
}

// Decode преобразует строку base62 обратно в uint64
func (e *Base62Encoder) Decode(s string) (uint64, error) {
	var n uint64
	for idx := 0; idx < len(s); idx++ {
		c := s[idx]
		var val uint64

		switch {
		case c >= '0' && c <= '9':
			val = uint64(c - '0')
		case c >= 'A' && c <= 'Z':
			val = uint64(c-'A') + 10
		case c >= 'a' && c <= 'z':
			val = uint64(c-'a') + 36
		default:
			return 0, fmt.Errorf("invalid character in base62 string: %c", c)
		}

		n = n*62 + val
	}
	return n, nil
}
