package service

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
	"tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	"tugas1-go/pertemuan-7-advanced-api-design/helper"
)

type AchievementService struct {
	repo repository.AchievementRepository
}

func NewAchievementService(repo repository.AchievementRepository) *AchievementService {
	return &AchievementService{repo: repo}
}

func translateAchievementError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NewAppError(fiber.StatusNotFound, "achievement tidak ditemukan")
	case errors.Is(err, repository.ErrStudentNotFound):
		return helper.NewAppError(fiber.StatusUnprocessableEntity, "student tidak ditemukan")
	default:
		return helper.NewAppError(fiber.StatusInternalServerError, "terjadi kesalahan pada database")
	}
}

func (s *AchievementService) List(c *fiber.Ctx) error {
	q := helper.ParseAchievementListQuery(c)
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	var items []model.Achievement
	var total int
	var err error
	meta := &model.Meta{Page: q.Page, Limit: q.Limit}
	if q.Cursor != "" {
		createdAt, id, decodeErr := helper.DecodeCursor(q.Cursor)
		if decodeErr != nil {
			return helper.NewAppError(fiber.StatusBadRequest, "cursor tidak valid")
		}
		items, err = s.repo.FindAfterCursor(ctx, q, createdAt, id)
	} else {
		items, total, err = s.repo.FindAll(ctx, q)
		meta.Total, meta.TotalPages = total, CountTotalPages(total, q.Limit)
	}
	if err != nil {
		return translateAchievementError(c, err)
	}
	if len(items) == q.Limit {
		last := items[len(items)-1]
		cursor, encodeErr := helper.EncodeCursor(last.CreatedAt, last.ID)
		if encodeErr != nil {
			return helper.NewAppError(fiber.StatusInternalServerError, "gagal membuat cursor")
		}
		meta.NextCursor = cursor
	}
	return helper.Negotiate(c, fiber.StatusOK, "daftar achievement berhasil diambil", items, meta)
}

func (s *AchievementService) Get(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.NewAppError(400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateAchievementError(c, err)
	}
	return helper.Success(c, 200, "achievement ditemukan", item)
}

func (s *AchievementService) Create(c *fiber.Ctx) error {
	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(400, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateRequest(req); len(errs) > 0 {
		return helper.NewValidationError(errs)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.NewAppError(400, "body harus berupa JSON yang valid")
	}
	req, errs := ValidateAchievementCreate(req, map[string]bool{"name": body["name"] != nil, "student_id": body["student_id"] != nil, "rank": body["rank"] != nil})
	if len(errs) > 0 {
		return helper.NewValidationError(errs)
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	item, err := s.repo.Create(ctx, model.Achievement{Name: req.Name, StudentID: req.StudentID, Rank: req.Rank})
	if err != nil {
		return translateAchievementError(c, err)
	}
	return helper.Created(c, "achievement berhasil ditambahkan", item, "/api/v1/achievements/"+strconv.Itoa(item.ID))
}

func (s *AchievementService) Replace(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.NewAppError(400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateAchievementError(c, err)
	}
	var req model.ReplaceAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(400, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateRequest(req); len(errs) > 0 {
		return helper.NewValidationError(errs)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.NewAppError(400, "body harus berupa JSON yang valid")
	}
	req, errs := ValidateAchievementReplace(req, map[string]bool{"name": body["name"] != nil, "student_id": body["student_id"] != nil, "rank": body["rank"] != nil})
	if len(errs) > 0 {
		return helper.NewValidationError(errs)
	}
	item.Name, item.StudentID, item.Rank = req.Name, req.StudentID, req.Rank
	item, err = s.repo.Update(ctx, item)
	if err != nil {
		return translateAchievementError(c, err)
	}
	return helper.Success(c, 200, "data achievement berhasil diganti", item)
}

func (s *AchievementService) Patch(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.NewAppError(400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateAchievementError(c, err)
	}
	var req model.PatchAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(400, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateRequest(req); len(errs) > 0 {
		return helper.NewValidationError(errs)
	}
	if IsEmptyAchievementPatch(req) {
		return helper.NewAppError(400, "tidak ada field yang diubah")
	}
	updated, errs := ApplyAchievementPatch(item, req)
	if len(errs) > 0 {
		return helper.NewValidationError(errs)
	}
	item, err = s.repo.Update(ctx, updated)
	if err != nil {
		return translateAchievementError(c, err)
	}
	return helper.Success(c, 200, "data achievement berhasil diperbarui", item)
}

func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.NewAppError(400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	if err := s.repo.Delete(ctx, id); err != nil {
		return translateAchievementError(c, err)
	}
	return helper.NoContent(c)
}
