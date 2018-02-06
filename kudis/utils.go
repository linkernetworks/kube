package kudis

func reduce(frames []int) int {
	sum := 0
	for _, f := range frames {
		sum += f
	}
	return sum
}
