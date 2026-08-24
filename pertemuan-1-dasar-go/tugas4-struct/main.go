package main

import "fmt"

type Student struct {
    ID       string
    Name     string
    Grade    float64
    IsActive bool
}

func (s Student) GetInfo() string {
    return fmt.Sprintf(
        "ID: %s | Nama: %s | Nilai: %.2f | Aktif: %t",
        s.ID, s.Name, s.Grade, s.IsActive,
    )
}

func (s *Student) UpdateGrade(grade float64) {
    s.Grade = grade
}

func (s *Student) Activate() {
    s.IsActive = true
}

func (s *Student) Deactivate() {
    s.IsActive = false
}

func main() {
    student := Student{
        ID:       "M001",
        Name:     "Falah",
        Grade:    80,
        IsActive: false,
    }

    fmt.Println(student.GetInfo())

    student.UpdateGrade(90)
    student.Activate()

    fmt.Println(student.GetInfo())

    student.Deactivate()
    fmt.Println(student.GetInfo())
}
