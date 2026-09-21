package helper

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

func Negotiate(c *fiber.Ctx, status int, message string, data any, meta *model.Meta) error {
	response := model.WebResponse{Success: true, Message: message, Data: data, Meta: meta}
	accept := strings.ToLower(strings.TrimSpace(c.Get("Accept")))
	if accept == "" || strings.Contains(accept, "application/json") {
		return c.Status(status).JSON(response)
	}
	if strings.Contains(accept, "text/csv") {
		c.Type("csv")
		return c.Status(status).SendString(toCSV(data))
	}
	return NewAppError(fiber.StatusNotAcceptable, "format response tidak didukung", map[string]string{"accept": "hanya application/json atau text/csv yang didukung"})
}

func toCSV(data any) string {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return fmt.Sprint(data)
	}
	var b strings.Builder
	w := csv.NewWriter(&b)
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Pointer {
			item = item.Elem()
		}
		if item.Kind() != reflect.Struct {
			_ = w.Write([]string{fmt.Sprint(item.Interface())})
			continue
		}
		if i == 0 {
			headers := make([]string, item.NumField())
			for j := range headers {
				headers[j] = item.Type().Field(j).Name
			}
			_ = w.Write(headers)
		}
		row := make([]string, item.NumField())
		for j := range row {
			row[j] = fmt.Sprint(item.Field(j).Interface())
		}
		_ = w.Write(row)
	}
	w.Flush()
	return b.String()
}
