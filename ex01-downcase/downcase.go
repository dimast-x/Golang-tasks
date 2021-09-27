package downcase

func Downcase(str string) (string, error) {
	var res string
	for _, letter := range str {
		if letter >= 'A' && letter <= 'Z' {
			letter += 32
		}
		res += string(letter)
	}
	return res, nil
}
