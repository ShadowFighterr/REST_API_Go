package sqlconnect

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"

	// "restapi/internal/repository/sqlconnect"
	"restapi/pkg/utils"
	"strconv"
	"strings"
)

func GetTeachersDbHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE 1=1"
	var args []interface{}

	query, args = utils.AddFilters(r, query, args)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleError(err, "Error retrieving data")
	}

	defer rows.Close()

	// teachersList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
		if err != nil {
			return nil, utils.HandleError(err, "Error retrieving data")
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

func GetTeacherByID(w http.ResponseWriter, id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.HandleError(err, "Error retrieving data")
	} else if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error retrieving data")
	}
	return teacher, nil
}

func AddTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.HandleError(err, "Error adding data")
	}
	defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, class, subject, email) VALUES (?, ?, ?, ?, ?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("teachers", models.Teacher{}))
	if err != nil {
		return nil, utils.HandleError(err, "Error adding data")
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		result, err := stmt.Exec(utils.GetStructValues(newTeacher)...)
		// result, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Class, newTeacher.Subject, newTeacher.Email)
		if err != nil {
			return nil, utils.HandleError(err, "Error adding data")
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, utils.HandleError(err, "Error adding data")
		}
		newTeacher.ID = int(id)
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func UpdateTeachersDBHandler(r *http.Request, id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}

	result, err := db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, class = ?, subject = ?, email = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.Email, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Teacher{}, utils.HandleError(err, "Error updating data")
		}
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}
	if rowsAffected == 0 {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}
	return teacher, utils.HandleError(err, "Error updating data")
}

func PatchTeacherDBHandler(updates []map[string]interface{}) (int64, error) {
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
			return 0, utils.HandleError(err, "Invalid update payload: missing teacher ID")
		}

		id, err := strconv.Atoi(fmt.Sprintf("%v", idValue))
		if err != nil {
			tx.Rollback()
			return 0, utils.HandleError(err, "Invalid teacher ID")
		}

		var existingTeacher models.Teacher
		err = tx.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Class, &existingTeacher.Subject, &existingTeacher.Email)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return 0, utils.HandleError(err, "Teacher not found")
			}
			return 0, utils.HandleError(err, "Error updating data")
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
							return 0, utils.HandleError(err, "Error updating data")
						}
					}
					break
				}
			}
		}

		result, err := tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, class = ?, subject = ?, email = ? WHERE id = ?", existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Class, existingTeacher.Subject, existingTeacher.Email, id)
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

func PatchOneTeacher(id int, updates map[string]interface{}) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Class, &existingTeacher.Subject, &existingTeacher.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Teacher{}, utils.HandleError(err, "Error updating data")
		}
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}

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
			return models.Teacher{}, utils.HandleError(err, "Error updating data")
		}
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}
	if rowsAffected == 0 {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
	if err != nil {
		return models.Teacher{}, utils.HandleError(err, "Error updating data")
	}
	return teacher, nil
}

func DeleteTeacher(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		return utils.HandleError(err, "Error deleting teacher")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleError(err, "Error deleting teacher")
	}
	if rowsAffected == 0 {
		return utils.HandleError(fmt.Errorf("Teacher with ID %d not found", id), "Error deleting teacher")
	}
	return nil
}

func DeleteTeachers(ids []int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.HandleError(err, "Error deleting data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.HandleError(err, "Error deleting data")
	}
	defer tx.Rollback()

	for _, id := range ids {
		result, err := tx.Exec("DELETE FROM teachers WHERE id = ?", id)
		if err != nil {
			return utils.HandleError(err, "Error deleting data")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return utils.HandleError(err, "Error deleting data")
		}
		if rowsAffected == 0 {
			return utils.HandleError(fmt.Errorf("Teacher with ID %d not found", id), "Error deleting data")
		}
	}

	err = tx.Commit()
	if err != nil {
		return utils.HandleError(err, "Error updating data")
	}
	return nil
}

func GetStudentsByTeacherIdFromDb(teacherId string, students []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		// log.Println(utils.HandleError())
		log.Println(err)
		return nil, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()
	query := `SELECT s.id, s.first_name, s.last_name, s.class, s.email
	FROM students s
	JOIN teachers ts ON s.class = ts.class
	WHERE ts.id = ?`
	rows, err := db.Query(query, teacherId)
	if err != nil {
		log.Println(err)
		return nil, utils.HandleError(err, "Error retrieving data")
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Class, &student.Email)
		if err != nil {
			log.Println(err)
			return nil, utils.HandleError(err, "Error retrieving data")
		}
		students = append(students, student)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return students, nil
}
