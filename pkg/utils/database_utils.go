package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func normalizeDBTag(tag string) string {
	return strings.TrimSpace(strings.Split(tag, ",")[0])
}

func GenerateInsertQuery(tableName string, model interface{}) string {
	modelType := reflect.TypeOf(model)
	columns := make([]string, 0, modelType.NumField())
	placeholders := make([]string, 0, modelType.NumField())

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := normalizeDBTag(field.Tag.Get("db"))
		if dbTag != "" && dbTag != "id" {
			columns = append(columns, dbTag)
			placeholders = append(placeholders, "?")
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	return query
}

func GetStructValues(model interface{}) []interface{} {
	modelValue := reflect.ValueOf(model)
	var values []interface{}

	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Type().Field(i)
		dbTag := normalizeDBTag(field.Tag.Get("db"))
		if dbTag != "" && dbTag != "id" {
			values = append(values, modelValue.Field(i).Interface())
		}
	}
	return values
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"class":      true,
		"position":   true,
		"subject":    true,
		"email":      true,
	}
	return validFields[field]
}

func AddSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["SortBy"]
	if len(sortParams) > 0 {
		query += " ORDER BY "
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 || !isValidSortField(parts[0]) || !isValidSortOrder(parts[1]) {
				continue
			}
			field := parts[0]
			order := parts[1]
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("%s %s", field, strings.ToUpper(order))
		}
	}
	return query
}

func AddFilters(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"class":      "class",
		"subject":    "subject",
		"email":      "email",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += fmt.Sprintf(" AND %s = ?", dbField)
			args = append(args, value)
		}
	}
	return query, args
}

func GetPaginationParams(r *http.Request) (int, int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	return page, limit
}
