package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"restapi/internal/models"
	"restapi/internal/repository/sqlconnect"
	"strconv"
	"strings"
)

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Teachers endpoint accessed")
	switch r.Method {
	case http.MethodGet:
		getTeachersHandler(w, r)
	case http.MethodPost:
		postTeachersHandler(w, r)
	case http.MethodPut:
		putTeachersHandler(w, r)
	case http.MethodPatch:
		patchTeachersHandler(w, r)
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE method in teachers routes\n"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	fmt.Fprint(w, "Teachers Endpoint")
}

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

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GET method in teachers routes\n"))
	db, err := sqlconnect.ConnectDB("school_management")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.Trim(path, "/")
	if idStr == "" {
		query := "SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE 1=1"
		var args []interface{}

		query, args = addFilters(r, query, args)

		query = addSorting(r, query)

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Failed to retrieve teachers", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var teachers []models.Teacher
		for rows.Next() {
			var teacher models.Teacher
			err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
			if err != nil {
				http.Error(w, "Failed to scan teacher", http.StatusInternalServerError)
				return
			}
			teachers = append(teachers, teacher)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(teachers)
		return
	}

	//Handle path parameters
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve teacher", http.StatusInternalServerError)
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

func postTeachersHandler(w http.ResponseWriter, r *http.Request) {
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

func putTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello PUT method in teachers routes\n"))
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
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

func patchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello PATCH method in teachers routes\n"))
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
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

	for field, value := range updates {
		switch field {
		case "first_name":
			existingTeacher.FirstName, _ = value.(string)
		case "last_name":
			existingTeacher.LastName, _ = value.(string)
		case "class":
			existingTeacher.Class, _ = value.(string)
		case "subject":
			existingTeacher.Subject, _ = value.(string)
		case "email":
			existingTeacher.Email, _ = value.(string)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}
