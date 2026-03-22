package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/internal/repository/sqlconnect"
	"restapi/pkg/utils"
	"strconv"
	"strings"
)

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GET method in teachers routes\n"))

	var teachers []models.Teacher
	teachers, err := sqlconnect.GetTeachersDbHandler(teachers, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teachers)
}

func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	teacher, err := sqlconnect.GetTeacherByID(w, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello GET method in teachers routes\n"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func PostTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var newTeachers []models.Teacher
	var rawTeachers []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawTeachers)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	val := reflect.TypeOf(models.Teacher{})
	allowedFields := make(map[string]struct{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
		allowedFields[fieldToAdd] = struct{}{}
	}

	for _, teacher := range rawTeachers {
		for key := range teacher {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, fmt.Sprintf("Invalid field: %s", key), http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newTeachers)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	for _, teacher := range newTeachers {
		val := reflect.ValueOf(teacher)
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.Kind() == reflect.String && field.String() == "" {
				http.Error(w, "All fields must be filled", http.StatusBadRequest)
				return
			}
		}
	}

	addedTeachers, err := sqlconnect.AddTeachersDBHandler(newTeachers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello POST method in teachers routes\n"))
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

	teacher, err := sqlconnect.UpdateTeachersDBHandler(r, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

// Patch /teachers/
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello PATCH method in teachers routes\n"))
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	rowsAffected, err := sqlconnect.PatchTeacherDBHandler(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	updatedTeacher, err := sqlconnect.PatchOneTeacher(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello 1 PATCH method in teachers routes\n"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)

}

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Hello DELETE method in teachers routes\n"))
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

	err = sqlconnect.DeleteTeacher(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	// w.Write([]byte("Hello DELETE method in teachers routes\n"))

	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteTeachers(ids)
	if err != nil {
		http.Error(w, "Failed to delete teachers", http.StatusInternalServerError)
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

func GetStudentsByTeacherId(w http.ResponseWriter, r *http.Request) {
	//admin, manager, executive
	_, err := utils.AuthorizeUser([]string{"admin", "manager", "executive"}, r.Context().Value(utils.ContextKey("role")).(string))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	teacherId := r.PathValue("id")

	var students []models.Student

	students, err = sqlconnect.GetStudentsByTeacherIdFromDb(teacherId, students)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func GetStudentCountByTeacherId(w http.ResponseWriter, r *http.Request) {
	teacherId := r.PathValue("id")

	var studentCount int

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "SELECT COUNT(*) FROM students WHERE class = (SELECT class FROM teachers WHERE id = ?)"
	err = db.QueryRow(query, teacherId).Scan(&studentCount)
	if err != nil {
		http.Error(w, "Failed to retrieve student count", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}{
		Status: "success",
		Count:  studentCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
