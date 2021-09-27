package stack

type Stack []int

func (stack *Stack) Push(num int) {
	*stack = append(*stack, num)
}

func (stack *Stack) Pop() int {
	len := len(*stack)
	if len != 0 {
		index := len - 1
		element := (*stack)[index]
		*stack = (*stack)[:index]
		return element
	}
	return 0
}
