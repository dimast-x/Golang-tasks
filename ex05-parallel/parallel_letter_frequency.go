package letter

type Letters map[rune]int

func Frequency(text string) Letters {
	frequency := Letters{}
	for _, letter := range text {
		frequency[letter]++
	}
	return frequency
}

func ConcurrentFrequency(inputs []string) Letters {
	res := make(chan Letters)
	for _, s := range inputs {
		go func(s string) {
			res <- Frequency(s)
		}(s)
	}
	output := <-res
	for i := 1; i < len(inputs); i++ {
		for index, freq := range <-res {
			output[index] += freq
		}
	}
	close(res)
	return output
}
