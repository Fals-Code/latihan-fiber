package service

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
	"tugas1-go/pertemuan-5-authentication-security/app/repository"
	"tugas1-go/pertemuan-5-authentication-security/helper"
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
		return helper.Fail(c, fiber.StatusNotFound, "achievement tidak ditemukan")
	case errors.Is(err, repository.ErrStudentNotFound):
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "student tidak ditemukan")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada database")
	}
}

func (s *AchievementService) List(c *fiber.Ctx) error {
	q := helper.ParseAchievementListQuery(c)
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateAchievementError(c, err)
	}
	return helper.SuccessList(c, "daftar achievement berhasil diambil", items, &model.Meta{Page: q.Page, Limit: q.Limit, Total: total, TotalPages: CountTotalPages(total, q.Limit)})
}

func (s *AchievementService) Get(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, 400, "id harus berupa angka positif")
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
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	req, errs := ValidateAchievementCreate(req, map[string]bool{"name": body["name"] != nil, "student_id": body["student_id"] != nil, "rank": body["rank"] != nil})
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
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
		return helper.Fail(c, 400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateAchievementError(c, err)
	}
	var req model.ReplaceAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	req, errs := ValidateAchievementReplace(req, map[string]bool{"name": body["name"] != nil, "student_id": body["student_id"] != nil, "rank": body["rank"] != nil})
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
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
		return helper.Fail(c, 400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateAchievementError(c, err)
	}
	var req model.PatchAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	if IsEmptyAchievementPatch(req) {
		return helper.Fail(c, 400, "tidak ada field yang diubah")
	}
	updated, errs := ApplyAchievementPatch(item, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
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
		return helper.Fail(c, 400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	if err := s.repo.Delete(ctx, id); err != nil {
		return translateAchievementError(c, err)
	}
	return helper.NoContent(c)
}
