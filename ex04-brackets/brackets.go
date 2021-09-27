package brackets

type Stack []string

func (stack *Stack) Push(str string) {
	*stack = append(*stack, str)
}

func (stack *Stack) Pop() string {
	len := len(*stack)
	if len != 0 {
		index := len - 1
		element := (*stack)[index]
		*stack = (*stack)[:index]
		return element
	}
	return ""
}

func (stack *Stack) Peek() string {
	len := len(*stack)
	if len != 0 {
		index := len - 1
		element := (*stack)[index]
		return element
	}
	return ""
}

func Bracket(str string) (bool, error) {
	var stack *Stack = new(Stack)
	state := true
	for i := 0; i < len(str); i++ {
		el := string(str[i])
		if el == "{" || el == "(" || el == "[" {
			stack.Push(el)
			state = false
		}
		if (el == "}" || el == ")" || el == "]") && !state {
			state = true
			last := stack.Peek()
			if (last == "{" && el == "}") || (last == "(" && el == ")") || (last == "[" && el == "]") {
				stack.Pop()
			} else {
				return false, nil
			}
		}
	}
	if state {
		return true, nil
	}
	return false, nil
}
