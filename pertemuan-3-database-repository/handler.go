package main

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

	var students []Student
	var nextID = 1

	// Mencari index mahasiswa berdasarkan ID
	func findStudentIndex(id int) int {
		for i := range students {
			if students[i].ID == id {
				return i
			}
		}
		return -1
	}

	// Pencarian nama tanpa membedakan huruf besar dan kecil
	func cocokPencarian(s Student, kata string) bool {
		kata = strings.ToLower(kata)
		return strings.Contains(strings.ToLower(s.Name), kata)
	}

	// Membaca ID dari URL
	func paramID(c *fiber.Ctx) (int, bool) {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil || id < 1 {
			return 0, false
		}
		return id, true
	}

	// Mengambil daftar mahasiswa
	func listStudents(c *fiber.Ctx) error {
		q := parseListQuery(c)

		hasil := []Student{}

		// Filter dan pencarian
		for _, s := range students {
			if q.IsActive != nil && s.IsActive != *q.IsActive {
				continue
			}

			if q.Search != "" && !cocokPencarian(s, q.Search) {
				continue
			}

			hasil = append(hasil, s)
		}

	// Sorting
	sort.SliceStable(hasil, func(i, j int) bool {
		switch q.Sort {
		case "nim":
			if q.Order == "desc" {
				return hasil[i].NIM > hasil[j].NIM
			}
			return hasil[i].NIM < hasil[j].NIM

		case "name":
			if q.Order == "desc" {
				return hasil[i].Name > hasil[j].Name
			}
			return hasil[i].Name < hasil[j].Name

		case "grade":
			if q.Order == "desc" {
				return hasil[i].Grade > hasil[j].Grade
			}
			return hasil[i].Grade < hasil[j].Grade

		case "is_active":
			if q.Order == "desc" {
				return hasil[i].IsActive && !hasil[j].IsActive
			}
			return !hasil[i].IsActive && hasil[j].IsActive

		default:
			if q.Order == "desc" {
				return hasil[i].ID > hasil[j].ID
			}
			return hasil[i].ID < hasil[j].ID
		}
	})

		// Pagination
		total := len(hasil)
		totalPages := (total + q.Limit - 1) / q.Limit

		mulai := (q.Page - 1) * q.Limit
		if mulai > total {
			mulai = total
		}

		akhir := mulai + q.Limit
		if akhir > total {
			akhir = total
		}

		return okList(
			c,
			"daftar mahasiswa berhasil diambil",
			hasil[mulai:akhir],
			&Meta{
				Page:       q.Page,
				Limit:      q.Limit,
				Total:      total,
				TotalPages: totalPages,
			},
		)
	}

	// Mengambil satu mahasiswa berdasarkan ID
	func getStudent(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
		}

		i := findStudentIndex(id)

		if i == -1 {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		return ok(c, "mahasiswa ditemukan", students[i])
	}

	// Menambahkan mahasiswa baru
	func createStudent(c *fiber.Ctx) error {
		var req CreateStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
		}

		var body map[string]json.RawMessage
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
		}

		errs := map[string]string{}

		if _, ada := body["nim"]; !ada {
			errs["nim"] = "wajib diisi"
		} else if req.NIM <= 0 {
			errs["nim"] = "harus berupa angka positif"
		}

		if _, ada := body["name"]; !ada {
			errs["name"] = "wajib diisi"
		} else {
			req.Name = strings.TrimSpace(req.Name)
			if req.Name == "" {
				errs["name"] = "wajib diisi"
			}
		}

		if _, ada := body["grade"]; !ada {
			errs["grade"] = "wajib diisi"
		} else if req.Grade < 0 || req.Grade > 100 {
			errs["grade"] = "harus berada antara 0 sampai 100"
		}

		if len(errs) > 0 {
			return failValidation(c, errs)
		}

		// NIM tidak boleh sama
		for _, s := range students {
			if s.NIM == req.NIM {
				return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
			}
		}

		student := Student{
			ID:       nextID,
			NIM:      req.NIM,
			Name:     req.Name,
			Grade:    req.Grade,
			IsActive: true,
		}

		students = append(students, student)
		nextID++

		return created(
			c,
			"mahasiswa berhasil ditambahkan",
			student,
			"/api/v1/students/"+strconv.Itoa(student.ID),
		)
	}

	// Mengganti seluruh data mahasiswa
	func replaceStudent(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
		}

		i := findStudentIndex(id)
		if i == -1 {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		var req ReplaceStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
		}

		var body map[string]json.RawMessage
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
		}

		errs := map[string]string{}

		if _, ada := body["nim"]; !ada {
			errs["nim"] = "wajib diisi pada PUT"
		} else if req.NIM <= 0 {
			errs["nim"] = "harus berupa angka positif"
		}

		if _, ada := body["name"]; !ada {
			errs["name"] = "wajib diisi pada PUT"
		} else {
			req.Name = strings.TrimSpace(req.Name)
			if req.Name == "" {
				errs["name"] = "tidak boleh kosong"
			}
		}

		if _, ada := body["grade"]; !ada {
			errs["grade"] = "wajib diisi pada PUT"
		} else if req.Grade < 0 || req.Grade > 100 {
			errs["grade"] = "harus berada antara 0 sampai 100"
		}

		if _, ada := body["is_active"]; !ada {
			errs["is_active"] = "wajib diisi pada PUT"
		}

		if len(errs) > 0 {
			return failValidation(c, errs)
		}

		// NIM tidak boleh dipakai mahasiswa lain
		for index, s := range students {
			if index != i && s.NIM == req.NIM {
				return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
			}
		}

		students[i].NIM = req.NIM
		students[i].Name = req.Name
		students[i].Grade = req.Grade
		students[i].IsActive = req.IsActive

		return ok(c, "data mahasiswa berhasil diganti", students[i])
	}

	// Mengubah sebagian data mahasiswa
	func patchStudent(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
		}

		i := findStudentIndex(id)
		if i == -1 {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		var req PatchStudentRequest

		if err := c.BodyParser(&req); err != nil {
			return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
		}

		if req.NIM == nil &&
			req.Name == nil &&
			req.Grade == nil &&
			req.IsActive == nil {
			return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
		}

		errs := map[string]string{}

		// Validasi dulu
		if req.NIM != nil && *req.NIM <= 0 {
			errs["nim"] = "harus berupa angka positif"
		}

		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)

			if name == "" {
				errs["name"] = "tidak boleh kosong"
			} else {
				*req.Name = name
			}
		}

		if req.Grade != nil {
			if *req.Grade < 0 || *req.Grade > 100 {
				errs["grade"] = "harus berada antara 0 sampai 100"
			}
		}

		if len(errs) > 0 {
			return failValidation(c, errs)
		}

		// Cek NIM duplikat
		if req.NIM != nil {
			for index, s := range students {
				if index != i && s.NIM == *req.NIM {
					return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
				}
			}
		}

		// Baru ubah data setelah semua valid
		if req.NIM != nil {
			students[i].NIM = *req.NIM
		}

		if req.Name != nil {
			students[i].Name = *req.Name
		}

		if req.Grade != nil {
			students[i].Grade = *req.Grade
		}

		if req.IsActive != nil {
			students[i].IsActive = *req.IsActive
		}

		return ok(c, "data mahasiswa berhasil diperbarui", students[i])
	}

	// Menghapus mahasiswa
	func deleteStudent(c *fiber.Ctx) error {
		id, valid := paramID(c)

		if !valid {
			return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
		}

		i := findStudentIndex(id)
		if i == -1 {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		students = append(students[:i], students[i+1:]...)

		return noContent(c)
	}