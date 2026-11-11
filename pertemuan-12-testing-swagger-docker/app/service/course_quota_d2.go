//go:build d2

package service

import (
	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/repository"
	"tugas1-go/pertemuan-12-testing-swagger-docker/helper"
)

// Two writes must commit together: the optimistic quota update and its event.
func (s *CourseService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.UpdateQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	var updated model.Course
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		var err error
		updated, err = s.courses.WithTx(tx).UpdateQuota(ctx, id, req.Quota, req.Version)
		if err != nil {
			return err
		}
		// The course UPDATE holds its row lock; enrollment writers must pass that
		// same lock. This SELECT snapshots the participant set before commit.
		rows, err := tx.Query(ctx, "SELECT user_id FROM enrollments WHERE course_id=$1 ORDER BY user_id", id)
		if err != nil {
			return err
		}
		recipients := []int{}
		for rows.Next() {
			var userID int
			if err := rows.Scan(&userID); err != nil {
				rows.Close()
				return err
			}
			recipients = append(recipients, userID)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		return s.outbox.WithTx(tx).Insert(ctx, model.EventCourseQuotaChanged, model.QuotaChangedEvent{
			CourseID: id, CourseCode: updated.Code, CourseName: updated.Name, Quota: updated.Quota, UserIDs: recipients,
		})
	})
	if err != nil {
		return translateCourseError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota berhasil diubah", updated)
}
