package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/internal/repository/sqlconnect"
	"strconv"
	"strings"
)

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"class":      true,
		"subject":    true,
		"email":      true,
	}
	return validFields[field]
}

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GET method in teachers routes\n"))

	query := "SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE 1=1"
	var args []interface{}

	query, args = addFilters(r, query, args)

	query = addSorting(r, query)

	teachersList, shouldReturn := sqlconnect.GetTeachersDbHandler(w, query, args)
	if shouldReturn {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teachersList)
}

func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GET method in teachers routes\n"))
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	teacher, shouldReturn := sqlconnect.GetTeacherByID(w, id)
	if shouldReturn {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func addSorting(r *http.Request, query string) string {
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

func addFilters(r *http.Request, query string, args []interface{}) (string, []interface{}) {
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

func PostTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello POST method in teachers routes\n"))
	db, err := sqlconnect.ConnectDB("school_management")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, class, subject, email) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		result, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Class, newTeacher.Subject, newTeacher.Email)
		if err != nil {
			http.Error(w, "Failed to insert teacher", http.StatusInternalServerError)
			return
		}
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to retrieve last insert ID", http.StatusInternalServerError)
			return
		}
		newTeacher.ID = int(id)
		addedTeachers[i] = newTeacher
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}

func PutTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello PUT method in teachers routes\n"))
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Teacher ID is required for update", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB("school_management")
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, class = ?, subject = ?, email = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.Email, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update teacher", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to retrieve affected rows", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
	if err != nil {
		http.Error(w, "Failed to retrieve updated teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

// Patch /teachers/
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDB("school_management")
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var updates []map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	rowsAffected := int64(0)
	for _, update := range updates {
		idValue, ok := update["id"]
		if !ok {
			tx.Rollback()
			http.Error(w, "Teacher ID is required for update", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(fmt.Sprintf("%v", idValue))
		if err != nil {
			tx.Rollback()
			log.Println(err)
			http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
			return
		}

		var existingTeacher models.Teacher
		err = tx.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Class, &existingTeacher.Subject, &existingTeacher.Email)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				http.Error(w, "Teacher not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to retrieve teacher", http.StatusInternalServerError)
			return
		}

		// Apply updates using reflection
		teacherValue := reflect.ValueOf(&existingTeacher).Elem()

		for field, value := range update {
			if field == "id" {
				continue // Skip ID field
			}
			for i := 0; i < teacherValue.NumField(); i++ {
				structField := teacherValue.Type().Field(i)
				jsonTag := structField.Tag.Get("json")
				jsonField := strings.Split(jsonTag, ",")[0]
				if jsonField == field {
					fieldValue := teacherValue.Field(i)
					if fieldValue.CanSet() {
						val := reflect.ValueOf(value)
						if val.Type().ConvertibleTo(fieldValue.Type()) {
							fieldValue.Set(val.Convert(fieldValue.Type()))
						} else {
							tx.Rollback()
							log.Printf("Invalid type for field %s: expected %s but got %s", field, fieldValue.Type(), val.Type())
							return
						}
					}
					break
				}
			}
		}

		result, err := tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, class = ?, subject = ?, email = ? WHERE id = ?", existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Class, existingTeacher.Subject, existingTeacher.Email, id)
		if err != nil {
			http.Error(w, "Failed to update teacher", http.StatusInternalServerError)
			return
		}

		rowsAffectedThisUpdate, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Failed to retrieve affected rows", http.StatusInternalServerError)
			return
		}
		rowsAffected += rowsAffectedThisUpdate
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello PATCH method in teachers routes\n"))

	response := struct {
		Status       string `json:"status"`
		Message      string `json:"message"`
		RowsAffected int64  `json:"rows_affected"`
	}{
		Status:       "success",
		Message:      "Teachers updated successfully",
		RowsAffected: rowsAffected,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Patch /teachers/{id}
func PatchOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Teacher ID is required for update", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB("school_management")
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Class, &existingTeacher.Subject, &existingTeacher.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve teacher", http.StatusInternalServerError)
		return
	}

	//Apply updates using reflection
	techerValue := reflect.ValueOf(&existingTeacher).Elem()

	for field, value := range updates {
		for i := 0; i < techerValue.NumField(); i++ {
			structField := techerValue.Type().Field(i)
			jsonTag := structField.Tag.Get("json")
			jsonField := strings.Split(jsonTag, ",")[0]
			if jsonField == field {
				fieldValue := techerValue.Field(i)
				if fieldValue.CanSet() {
					val := reflect.ValueOf(value)
					if val.Type().ConvertibleTo(fieldValue.Type()) {
						fieldValue.Set(val.Convert(fieldValue.Type()))
					}
				}
				break
			}
		}
	}

	result, err := db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, class = ?, subject = ?, email = ? WHERE id = ?", existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Class, existingTeacher.Subject, existingTeacher.Email, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update teacher", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to retrieve affected rows", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
	if err != nil {
		http.Error(w, "Failed to retrieve updated teacher", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello 1 PATCH method in teachers routes\n"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello DELETE method in teachers routes\n"))
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Teacher ID is required for deletion", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB("school_management")
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete teacher", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to retrieve affected rows", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}

	// w.WriteHeader(http.StatusNoContent)

	//Reponse with success message
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: fmt.Sprintf("Teacher with ID %d deleted successfully", id),
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello DELETE method in teachers routes\n"))
	db, err := sqlconnect.ConnectDB("school_management")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, id := range ids {
		result, err := tx.Exec("DELETE FROM teachers WHERE id = ?", id)
		if err != nil {
			http.Error(w, "Failed to delete teacher", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Failed to retrieve affected rows", http.StatusInternalServerError)
			return
		}
		if rowsAffected == 0 {
			http.Error(w, fmt.Sprintf("Teacher with ID %d not found", id), http.StatusNotFound)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: fmt.Sprintf("%d teachers deleted successfully", len(ids)),
	}
	json.NewEncoder(w).Encode(response)
}
