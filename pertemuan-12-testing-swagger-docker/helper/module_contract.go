package helper

func NewError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}
func BadRequest(message string) error            { return NewAppError(400, message) }
func Unauthorized(message string) error          { return NewAppError(401, message) }
func Forbidden(message string) error             { return NewAppError(403, message) }
func NotFound(message string) error              { return NewAppError(404, message) }
func ServiceUnavailable(message string) error    { return NewError(503, "SERVICE_UNAVAILABLE", message) }
func Validation(fields map[string]string) error  { return NewValidationError(fields) }
func ValidateStruct(value any) map[string]string { return ValidateRequest(value) }
