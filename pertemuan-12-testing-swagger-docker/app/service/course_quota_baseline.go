//go:build !d2

package service

import (
	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
	"tugas1-go/pertemuan-12-testing-swagger-docker/helper"
)

func (s *CourseService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Satu pernyataan UPDATE saja, sehingga tidak perlu WithTx:
	// setiap pernyataan tunggal di PostgreSQL sudah atomik dengan sendirinya.
	updated, err := s.courses.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateCourseError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota berhasil diubah", updated)
}
