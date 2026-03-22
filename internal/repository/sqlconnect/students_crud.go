package sqlconnect

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
)

func GetStudentsDbHandler(students []models.Student, r *http.Request, limit, page int) ([]models.Student, int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, 0, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, class, email FROM students WHERE 1=1"
	var args []interface{}

	query, args = utils.AddFilters(r, query, args)

	offset := (page - 1) * limit
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, utils.HandleError(err, "Error retrieving data")
	}

	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Class, &student.Email)
		if err != nil {
			return nil, 0, utils.HandleError(err, "Error retrieving data")
		}
		students = append(students, student)
	}

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM students"
	err = db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		utils.HandleError(err, "Error retrieving data")
		totalCount = 0
	}

	return students, totalCount, nil
}

func GetStudentByID(w http.ResponseWriter, id int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	var student models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, class, email FROM students WHERE id = ?", id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Class, &student.Email)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.HandleError(err, "Error retrieving data")
	} else if err != nil {
		return models.Student{}, utils.HandleError(err, "Error retrieving data")
	}
	return student, nil
}

func AddStudentsDBHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.HandleError(err, "Error adding data")
	}
	defer db.Close()

	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.HandleError(err, "Error adding data")
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		result, err := stmt.Exec(utils.GetStructValues(newStudent)...)
		if err != nil {
			return nil, utils.HandleError(err, "Error adding data")
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, utils.HandleError(err, "Error adding data")
		}
		newStudent.ID = int(id)
		addedStudents[i] = newStudent
	}
	return addedStudents, nil
}

func UpdateStudentsDBHandler(r *http.Request, id int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	var updatedStudent models.Student
	err = json.NewDecoder(r.Body).Decode(&updatedStudent)
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}

	result, err := db.Exec("UPDATE students SET first_name = ?, last_name = ?, class = ?, email = ? WHERE id = ?", updatedStudent.FirstName, updatedStudent.LastName, updatedStudent.Class, updatedStudent.Email, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.HandleError(err, "Error updating data")
		}
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}
	if rowsAffected == 0 {
		return models.Student{}, utils.HandleError(fmt.Errorf("student with ID %d not found", id), "Error updating data")
	}

	var student models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, class, email FROM students WHERE id = ?", id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Class, &student.Email)
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}
	return student, nil
}

func PatchStudentDBHandler(updates []map[string]interface{}) (int64, error) {
	db, err := ConnectDB()
	if err != nil {
		return 0, utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return 0, utils.HandleError(err, "Error updating data")
	}
	defer tx.Rollback()

	rowsAffected := int64(0)
	for _, update := range updates {
		idValue, ok := update["id"]
		if !ok {
			tx.Rollback()
			return 0, utils.HandleError(err, "Invalid update payload: missing student ID")
		}

		id, err := strconv.Atoi(fmt.Sprintf("%v", idValue))
		if err != nil {
			tx.Rollback()
			return 0, utils.HandleError(err, "Invalid student ID")
		}

		var existingStudent models.Student
		err = tx.QueryRow("SELECT id, first_name, last_name, class, email FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Class, &existingStudent.Email)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return 0, utils.HandleError(err, "Student not found")
			}
			return 0, utils.HandleError(err, "Error updating data")
		}

		// Apply updates using reflection
		studentValue := reflect.ValueOf(&existingStudent).Elem()

		for field, value := range update {
			if field == "id" {
				continue // Skip ID field
			}
			for i := 0; i < studentValue.NumField(); i++ {
				structField := studentValue.Type().Field(i)
				jsonTag := structField.Tag.Get("json")
				jsonField := strings.Split(jsonTag, ",")[0]
				if jsonField == field {
					fieldValue := studentValue.Field(i)
					if fieldValue.CanSet() {
						val := reflect.ValueOf(value)
						if val.Type().ConvertibleTo(fieldValue.Type()) {
							fieldValue.Set(val.Convert(fieldValue.Type()))
						} else {
							tx.Rollback()
							log.Printf("Invalid type for field %s: expected %s but got %s", field, fieldValue.Type(), val.Type())
							return 0, utils.HandleError(err, "Error updating data")
						}
					}
					break
				}
			}
		}

		result, err := tx.Exec("UPDATE students SET first_name = ?, last_name = ?, class = ?, email = ? WHERE id = ?", existingStudent.FirstName, existingStudent.LastName, existingStudent.Class, existingStudent.Email, id)
		if err != nil {
			return 0, utils.HandleError(err, "Error updating data")
		}

		rowsAffectedThisUpdate, err := result.RowsAffected()
		if err != nil {
			return 0, utils.HandleError(err, "Error updating data")
		}
		rowsAffected += rowsAffectedThisUpdate
	}

	err = tx.Commit()
	if err != nil {
		return 0, utils.HandleError(err, "Error updating data")
	}
	return rowsAffected, nil
}

func PatchOneStudent(id int, updates map[string]interface{}) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, class, email FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Class, &existingStudent.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.HandleError(err, "Error updating data")
		}
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}

	studentValue := reflect.ValueOf(&existingStudent).Elem()

	for field, value := range updates {
		for i := 0; i < studentValue.NumField(); i++ {
			structField := studentValue.Type().Field(i)
			jsonTag := structField.Tag.Get("json")
			jsonField := strings.Split(jsonTag, ",")[0]
			if jsonField == field {
				fieldValue := studentValue.Field(i)
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

	result, err := db.Exec("UPDATE students SET first_name = ?, last_name = ?, class = ?, email = ? WHERE id = ?", existingStudent.FirstName, existingStudent.LastName, existingStudent.Class, existingStudent.Email, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.HandleError(err, "Error updating data")
		}
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}
	if rowsAffected == 0 {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}

	var student models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, class, email FROM students WHERE id = ?", id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Class, &student.Email)
	if err != nil {
		return models.Student{}, utils.HandleError(err, "Error updating data")
	}
	return student, nil
}

func DeleteStudent(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return utils.HandleError(err, "Error deleting student")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleError(err, "Error deleting student")
	}
	if rowsAffected == 0 {
		return utils.HandleError(fmt.Errorf("Student with ID %d not found", id), "Error deleting student")
	}
	return nil
}

func DeleteStudents(ids []int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.HandleError(err, "Error updating data")
	}
	defer tx.Rollback()

	for _, id := range ids {
		result, err := tx.Exec("DELETE FROM students WHERE id = ?", id)
		if err != nil {
			return utils.HandleError(err, "Error updating data")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return utils.HandleError(err, "Error updating data")
		}
		if rowsAffected == 0 {
			return utils.HandleError(fmt.Errorf("Student with ID %d not found", id), "Error updating data")
		}
	}

	err = tx.Commit()
	if err != nil {
		return utils.HandleError(err, "Error updating data")
	}
	return nil
}
