package cipher

import (
	"strings"
	"unicode"
)

type Cipher interface {
	Encode(string) string
	Decode(string) string
}

func NewCaesar() Cipher {
	return NewShift(3)
}

type Shift struct {
	shift int
}

func NewShift(num int) Cipher {
	if num == 0 || num < -25 || num > 25 {
		return nil
	}
	return &Shift{shift: num}
}

func (c Shift) Encode(str string) string {
	var res string
	str = strings.ToLower(str)
	for _, letter := range str {
		if unicode.IsLetter(letter) {
			letter += rune(c.shift)
			if letter > 'z' {
				letter -= 26
			} else if letter < 'a' {
				letter += 26
			}
			res += string(letter)
		}
	}
	return res
}

func (c Shift) Decode(str string) string {
	res := ""
	for _, letter := range str {
		letter -= rune(c.shift)
		if letter > 'z' {
			letter -= 26
		} else if letter < 'a' {
			letter += 26
		}
		res += string(letter)
	}
	return res
}

type Vigenere struct {
	key string
}

func NewVigenere(str string) Cipher {
	check := false
	for _, letter := range str {
		if letter < 'a' || letter > 'z' {
			return nil
		} else if letter > 'a' {
			check = true
		}
	}
	if !check {
		return nil
	}
	return &Vigenere{key: str}
}

func (v Vigenere) Encode(str string) string {
	res, count := "", 0
	str = strings.ToLower(str)
	for _, letter := range str {
		if unicode.IsLetter(letter) {
			res += string('a' + (letter-'a'+rune(v.key[count%len(v.key)])-'a')%26)
			count++
		}
	}
	return res
}

func (v Vigenere) Decode(str string) string {
	res, count := "", 0
	str = strings.ToLower(str)
	for _, letter := range str {
		if unicode.IsLetter(letter) {
			res += string('a' + (letter+26-rune(v.key[count%len(v.key)]))%26)
			count++
		}
	}
	return res
}
