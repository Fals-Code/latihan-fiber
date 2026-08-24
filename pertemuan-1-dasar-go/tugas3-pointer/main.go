package main

import "fmt"

func swap(a, b *int) {
    temp := *a
    *a = *b
    *b = temp
}

func updateSlice(s *[]string, newItem string) {
    *s = append(*s, newItem)
}

func changeValue(n int) {
    n = 100
}

func changePointer(n *int) {
    *n = 100
}

func main() {
    a := 10
    b := 20

    swap(&a, &b)
    fmt.Println("Hasil swap:", a, b)

    items := []string{"Go", "Git"}
    updateSlice(&items, "Fiber")
    fmt.Println("Slice:", items)

    number := 10
    changeValue(number)
    fmt.Println("Pass by value:", number)

    changePointer(&number)
    fmt.Println("Dengan pointer:", number)
}
