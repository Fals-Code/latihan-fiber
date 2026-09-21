# BUG_REPORT

| # | File/line | Category | Symptom | Root cause | Fix | Evidence |
|---|---|---|---|---|---|---|
| 1 | config/app.go | error handling | returned service errors were not centrally formatted | handlers wrote responses directly | AppError and Fiber ErrorHandler | go build/test/vet pass |
| 2 | middleware/middleware.go | observability | request status could be observed before final error handling | status was logged from response only | logger returns downstream error after recording final status | focused code review |
| 3 | app/model/student.go | validation | DTO validation was manual and inconsistent | request structs had no validator tags | validator tags added | build pass |
| 4 | helper/errors.go | pagination | cursors could not be represented safely | no cursor codec | EncodeCursor/DecodeCursor | build pass |
| 5 | app/repository/student_repository.go | pagination | list API only supported offsets | repository lacked keyset query | FindAfterCursor added | build pass |
| 6 | migrations/001_create_students.sql | database | cursor ordering lacked supporting index | no composite index | created_at DESC,id DESC index | migration review |
| 7 | migrations/002_create_achievements.sql | database | achievement cursor ordering lacked supporting index | no composite index | created_at DESC,id DESC index | migration review |
| 8 | services and middleware | error propagation | helper.Fail responses bypassed centralized handling | response helpers were called in business logic | handlers now return AppError | grep/build review |
| 9 | config/app.go | error payload | validation errors lost their field map | generic handler did not preserve structured errors | AppError.Errors included in response | build/test/vet pass |
