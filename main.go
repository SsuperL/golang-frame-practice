package main

import "fmt"

func main() {
	a := []int{1, 2}
	b := a
	a[0] = 100
	fmt.Println(a)
	fmt.Println(b)
}
