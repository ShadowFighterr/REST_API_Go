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

func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var students []models.Student
	page, limit := utils.GetPaginationParams(r)

	students, totalStudents, err := sqlconnect.GetStudentsDbHandler(students, r, limit, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello GET method in students routes\n"))
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status   string           `json:"status"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
		Count    int              `json:"count"`
		Data     []models.Student `json:"data"`
	}{
		Status:   "success",
		Page:     page,
		PageSize: limit,
		Count:    totalStudents,
		Data:     students,
	}
	json.NewEncoder(w).Encode(response)
}

func GetOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	student, err := sqlconnect.GetStudentByID(w, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello GET method in students routes\n"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func PostStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var newStudents []models.Student
	var rawStudents []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawStudents)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	val := reflect.TypeOf(models.Student{})
	allowedFields := make(map[string]struct{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
		allowedFields[fieldToAdd] = struct{}{}
	}

	for _, student := range rawStudents {
		for key := range student {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, fmt.Sprintf("Invalid field: %s", key), http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newStudents)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	for _, student := range newStudents {
		val := reflect.ValueOf(student)
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.Kind() == reflect.String && field.String() == "" {
				http.Error(w, "All fields must be filled", http.StatusBadRequest)
				return
			}
		}
	}

	addedStudents, err := sqlconnect.AddStudentsDBHandler(newStudents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello POST method in students routes\n"))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(response)
}

func PutStudentsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Student ID is required for update", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	student, err := sqlconnect.UpdateStudentsDBHandler(r, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

// Patch /students/
func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello PATCH method in students routes\n"))
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	rowsAffected, err := sqlconnect.PatchStudentDBHandler(updates)
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
		Message:      "Students updated successfully",
		RowsAffected: rowsAffected,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Patch /students/{id}
func PatchOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Student ID is required for update", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedStudent, err := sqlconnect.PatchOneStudent(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello 1 PATCH method in students routes\n"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudent)

}

func DeleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Hello DELETE method in students routes\n"))
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Student ID is required for deletion", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteStudent(id)
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
		Message: fmt.Sprintf("Student with ID %d deleted successfully", id),
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Hello DELETE method in students routes\n"))

	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteStudents(ids)
	if err != nil {
		http.Error(w, "Failed to delete students", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: fmt.Sprintf("%d students deleted successfully", len(ids)),
	}
	json.NewEncoder(w).Encode(response)
}
