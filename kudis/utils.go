package kudis

import "strings"

func reduce(frames []int) int {
	sum := 0
	for _, f := range frames {
		sum += f
	}
	return sum
}

func podNameToJobName(podName string) string {
	p := strings.Split(podName, "-")
	// pop the last element from slice
	return strings.Join(p[:len(p)-1], "-")
}
