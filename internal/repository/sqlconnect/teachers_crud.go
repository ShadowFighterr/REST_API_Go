package sqlconnect

import (
	"database/sql"
	"net/http"
	"restapi/internal/models"
)

func GetTeachersDbHandler(w http.ResponseWriter, query string, args []interface{}) ([]models.Teacher, bool) {
	db, err := ConnectDB("school_management")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return nil, true
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to retrieve teachers", http.StatusInternalServerError)
		return nil, true
	}

	defer rows.Close()

	teachersList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
		if err != nil {
			http.Error(w, "Failed to scan teacher", http.StatusInternalServerError)
			return nil, true
		}
		teachersList = append(teachersList, teacher)
	}
	return teachersList, false
}

func GetTeacherByID(w http.ResponseWriter, id int) (models.Teacher, bool) {
	db, err := ConnectDB("school_management")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return models.Teacher{}, true
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, class, subject, email FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Class, &teacher.Subject, &teacher.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Teacher{}, true
	} else if err != nil {
		http.Error(w, "Failed to retrieve teacher", http.StatusInternalServerError)
		return models.Teacher{}, true
	}
	return teacher, false
}
