package main

import (
	"bufio"
	"fmt"
	"os"
)

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func appendUnique(a []int, x int) []int {
	for _, y := range a {
		if x == y {
			return a
		}
	}
	return append(a, x)
}

func max(a []int) (int, int) {
	idx, max := 0, 0
	for i, val := range a {
		if val > max {
			max, idx = val, i
		}
	}
	return idx, max
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mapEq(a, b map[int]int) bool {
	for k, v := range b {
		if a[k] != v {
			return false
		}
	}
	return true
}

func avgMapInt(a, b map[int]int) map[int]int {
	//j := make(map[int]int)
	j := joinMapIntInt(a, b)
	for k, v := range j {
		j[k] = v / 2
	}
	return j
}

func median(x, weights []float64) float64 {
	size := 0.0
	for _, v := range weights {
		size += v
	}

	tmp := 0.0
	for k, v := range weights {
		tmp += v
		if (size / 2) < tmp {
			return x[k]
		}
	}
	return 0

}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func pauseit() {
	reader := bufio.NewReader(os.Stdin)
	// fmt.Print("Enter.")
	text, _ := reader.ReadString('\n')
	fmt.Println(text)
	// fmt.Println("Enter texttext: ")
	// text2 := ""
	// fmt.Scanln(text2)
	// fmt.Println(text2)
}
