package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
)

var studentNIMPattern = regexp.MustCompile(`^[0-9]{12}$`)
var studentYearPattern = regexp.MustCompile(`^[0-9]{4}$`)

var ErrStudentValidation = errors.New("student validation failed")
var ErrDuplicateStudentNIM = errors.New("student NIM already exists")
var ErrDuplicateStudentEmail = errors.New("student email already exists")
var ErrStudentNotFound = errors.New("student not found")
var ErrStudentForbidden = errors.New("student access forbidden")

type StudentStore interface {
	List(context.Context, model.StudentFilters) ([]model.Student, int64, error)
	Create(context.Context, model.NewStudent, func() (string, error)) (model.Student, error)
	GetDetail(context.Context, int64) (model.StudentDetail, error)
	StudentIDByUserID(context.Context, int64) (int64, error)
	Update(context.Context, int64, model.StudentUpdate) (model.Student, error)
	SoftDelete(context.Context, int64) error
}

type StudentService struct {
	store StudentStore
	now   func() time.Time
}

func NewStudentService(store StudentStore) *StudentService {
	return &StudentService{store: store, now: time.Now}
}

func ValidateStudentQuery(query map[string]string) (model.StudentFilters, map[string][]string) {
	filters := model.StudentFilters{Page: 1, PerPage: 10, Prodi: strings.TrimSpace(query["prodi"]), Search: strings.TrimSpace(query["search"]), Sort: query["sort"]}
	validation := make(map[string][]string)
	if raw := query["page"]; raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			validation["page"] = append(validation["page"], "Page minimal 1")
		} else {
			filters.Page = value
		}
	}
	if raw := query["per_page"]; raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 50 {
			validation["per_page"] = append(validation["per_page"], "Per page harus antara 1 dan 50")
		} else {
			filters.PerPage = value
		}
	}
	if raw := query["angkatan"]; raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1000 || value > 9999 {
			validation["angkatan"] = append(validation["angkatan"], "Angkatan harus 4 digit")
		} else {
			filters.Angkatan = &value
		}
	}
	if uint64(filters.Page-1) > uint64(math.MaxInt64/int64(filters.PerPage)) {
		validation["page"] = append(validation["page"], "Page di luar batas yang didukung")
	}
	if filters.Sort != "" && filters.Sort != "nama" && filters.Sort != "-ipk_terakhir" {
		validation["sort"] = append(validation["sort"], "Sort harus nama atau -ipk_terakhir")
	}
	if len(validation) > 0 {
		return model.StudentFilters{}, validation
	}
	return filters, nil
}

func (s *StudentService) List(ctx context.Context, query map[string]string) ([]model.Student, model.StudentFilters, int64, map[string][]string, error) {
	filters, validation := ValidateStudentQuery(query)
	if validation != nil {
		return nil, model.StudentFilters{}, 0, validation, nil
	}
	students, total, err := s.store.List(ctx, filters)
	return students, filters, total, nil, err
}

func (s *StudentService) Create(ctx context.Context, input model.NewStudent) (model.Student, map[string][]string, error) {
	validation := s.validateCreate(input)
	if validation != nil {
		return model.Student{}, validation, nil
	}
	student, err := s.store.Create(ctx, input, func() (string, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.NIM), bcrypt.DefaultCost)
		return string(hash), err
	})
	if errors.Is(err, repository.ErrDuplicateNIM) {
		return model.Student{}, map[string][]string{"nim": {"NIM sudah terdaftar"}}, nil
	}
	if errors.Is(err, repository.ErrDuplicateEmail) {
		return model.Student{}, map[string][]string{"email": {"Email sudah terdaftar"}}, nil
	}
	return student, nil, err
}

func (s *StudentService) Detail(ctx context.Context, id int64, actor model.CurrentUser) (model.StudentDetail, error) {
	detail, err := s.store.GetDetail(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.StudentDetail{}, ErrStudentNotFound
	}
	if err != nil {
		return model.StudentDetail{}, err
	}
	if actor.Role == "mahasiswa" {
		ownedID, err := s.store.StudentIDByUserID(ctx, actor.ID)
		if errors.Is(err, repository.ErrNotFound) {
			return model.StudentDetail{}, ErrStudentNotFound
		}
		if err != nil {
			return model.StudentDetail{}, err
		}
		if ownedID != id {
			return model.StudentDetail{}, ErrStudentForbidden
		}
	}
	if detail.Courses == nil {
		detail.Courses = []model.StudentCourse{}
	}
	detail.TotalSKS = 0
	for _, course := range detail.Courses {
		detail.TotalSKS += course.SKS
	}
	detail.BatasSKS = SKSLimit(detail.IPKTerakhir)
	return detail, nil
}

func SKSLimit(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}

func (s *StudentService) Update(ctx context.Context, id int64, request map[string]any) (model.Student, map[string][]string, error) {
	update := model.StudentUpdate{}
	validation := make(map[string][]string)
	for field := range request {
		switch field {
		case "nim":
			validation["nim"] = append(validation["nim"], "NIM tidak dapat diubah")
		case "nama", "prodi", "angkatan", "ipk_terakhir":
		default:
			validation[field] = append(validation[field], "Field tidak dapat diperbarui")
		}
	}
	if value, ok := request["nama"]; ok {
		name, valid := value.(string)
		if !valid || strings.TrimSpace(name) == "" {
			validation["nama"] = append(validation["nama"], "Nama tidak boleh kosong")
		} else {
			update.Nama = &name
		}
	}
	if value, ok := request["prodi"]; ok {
		prodi, valid := value.(string)
		if !valid || strings.TrimSpace(prodi) == "" {
			validation["prodi"] = append(validation["prodi"], "Prodi tidak boleh kosong")
		} else {
			update.Prodi = &prodi
		}
	}
	if value, ok := request["angkatan"]; ok {
		year, valid := numericInt(value)
		if !valid || year < 1000 || year > 9999 {
			validation["angkatan"] = append(validation["angkatan"], "Angkatan harus 4 digit")
		} else if year > s.now().Year() {
			validation["angkatan"] = append(validation["angkatan"], fmt.Sprintf("Angkatan tidak boleh melebihi tahun %d", s.now().Year()))
		} else {
			update.Angkatan = &year
		}
	}
	if value, ok := request["ipk_terakhir"]; ok {
		ipk, valid := numericFloat(value)
		if !valid || ipk < 0 || ipk > 4 {
			validation["ipk_terakhir"] = append(validation["ipk_terakhir"], "IPK harus antara 0.00 dan 4.00")
		} else {
			update.IPKTerakhir = &ipk
		}
	}
	if len(validation) > 0 {
		return model.Student{}, validation, nil
	}
	if update.Nama == nil && update.Prodi == nil && update.Angkatan == nil && update.IPKTerakhir == nil {
		return model.Student{}, map[string][]string{"body": {"Minimal satu field yang dapat diperbarui wajib diisi"}}, nil
	}
	student, err := s.store.Update(ctx, id, update)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Student{}, nil, ErrStudentNotFound
	}
	return student, nil, err
}

func numericInt(value any) (int, bool) {
	switch number := value.(type) {
	case float64:
		if number != math.Trunc(number) || number < math.MinInt || number > math.MaxInt {
			return 0, false
		}
		return int(number), true
	case int:
		return number, true
	default:
		return 0, false
	}
}

func numericFloat(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case int:
		return float64(number), true
	default:
		return 0, false
	}
}

func (s *StudentService) Delete(ctx context.Context, id int64) error {
	err := s.store.SoftDelete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrStudentNotFound
	}
	return err
}

func (s *StudentService) validateCreate(input model.NewStudent) map[string][]string {
	validation := make(map[string][]string)
	if !studentNIMPattern.MatchString(input.NIM) {
		validation["nim"] = append(validation["nim"], "NIM wajib terdiri dari 12 digit angka")
	}
	if strings.TrimSpace(input.Nama) == "" {
		validation["nama"] = append(validation["nama"], "Nama wajib diisi")
	}
	if strings.TrimSpace(input.Prodi) == "" {
		validation["prodi"] = append(validation["prodi"], "Prodi wajib diisi")
	}
	if strings.TrimSpace(input.Email) == "" {
		validation["email"] = append(validation["email"], "Email wajib diisi")
	} else if parsed, err := mail.ParseAddress(input.Email); err != nil || parsed.Address != input.Email {
		validation["email"] = append(validation["email"], "Format email tidak valid")
	}
	yearString := strconv.Itoa(input.Angkatan)
	if !studentYearPattern.MatchString(yearString) {
		validation["angkatan"] = append(validation["angkatan"], "Angkatan harus 4 digit")
	} else if input.Angkatan > s.now().Year() {
		validation["angkatan"] = append(validation["angkatan"], fmt.Sprintf("Angkatan tidak boleh melebihi tahun %d", s.now().Year()))
	}
	if input.IPKTerakhir < 0 || input.IPKTerakhir > 4 {
		validation["ipk_terakhir"] = append(validation["ipk_terakhir"], "IPK harus antara 0.00 dan 4.00")
	}
	if len(validation) == 0 {
		return nil
	}
	return validation
}
