package main

import "fmt"

func main() {
    var name string = "Falah"
    var age int = 21
    var gpa float64 = 3.75
    var isActive bool = true
    var skills []string = []string{"Go", "JavaScript", "SQL"}

    fmt.Println(name, age, gpa, isActive, skills)

    students := map[string]string{
        "Andi": "Teknik Informatika",
        "Budi": "Sistem Informasi",
    }

    students["Citra"] = "Teknik Komputer"

    major, exists := students["Andi"]
    if exists {
        fmt.Println("Andi:", major)
    }

    delete(students, "Budi")

    for name, major := range students {
        fmt.Println(name, "-", major)
    }
}
