package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-mail/mail/v2"
)

func GetExecutivesDbHandler(executives []models.Executive, r *http.Request) ([]models.Executive, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, position, email FROM executives WHERE 1=1"
	var args []interface{}

	query, args = utils.AddFilters(r, query, args)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.HandleError(err, "Error retrieving data")
	}

	defer rows.Close()

	for rows.Next() {
		var executive models.Executive
		err := rows.Scan(&executive.ID, &executive.FirstName, &executive.LastName, &executive.Position, &executive.Email)
		if err != nil {
			return nil, utils.HandleError(err, "Error retrieving data")
		}
		executives = append(executives, executive)
	}
	return executives, nil
}

func GetExecutiveByID(w http.ResponseWriter, id int) (models.Executive, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	var executive models.Executive
	err = db.QueryRow("SELECT id, first_name, last_name, position, email FROM executives WHERE id = ?", id).Scan(&executive.ID, &executive.FirstName, &executive.LastName, &executive.Position, &executive.Email)
	if err == sql.ErrNoRows {
		return models.Executive{}, utils.HandleError(err, "Error retrieving data")
	} else if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error retrieving data")
	}
	return executive, nil
}

func AddExecutivesDBHandler(newExecutives []models.Executive) ([]models.Executive, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.HandleError(err, "Error adding data")
	}
	defer db.Close()

	stmt, err := db.Prepare(utils.GenerateInsertQuery("executives", models.Executive{}))
	if err != nil {
		return nil, utils.HandleError(err, "Error adding data")
	}
	defer stmt.Close()

	addedExecutives := make([]models.Executive, len(newExecutives))
	for i, newExecutive := range newExecutives {
		newExecutive.Password, err = utils.HashPassword(newExecutive.Password)
		if err != nil {
			return nil, err
		}

		result, err := stmt.Exec(utils.GetStructValues(newExecutive)...)
		if err != nil {
			return nil, utils.HandleError(err, "Error adding data")
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, utils.HandleError(err, "Error adding data")
		}

		newExecutive.ID = int(id)
		addedExecutives[i] = newExecutive
	}
	return addedExecutives, nil
}

func UpdateExecutivesDBHandler(r *http.Request, id int) (models.Executive, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	var existingID int
	err = db.QueryRow("SELECT id FROM executives WHERE id = ?", id).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Executive{}, utils.HandleError(fmt.Errorf("executive with ID %d not found", id), "Error updating data")
		}
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}

	var updatedExecutive models.Executive
	err = json.NewDecoder(r.Body).Decode(&updatedExecutive)
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}

	result, err := db.Exec("UPDATE executives SET first_name = ?, last_name = ?, position = ?, email = ? WHERE id = ?", updatedExecutive.FirstName, updatedExecutive.LastName, updatedExecutive.Position, updatedExecutive.Email, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Executive{}, utils.HandleError(err, "Error updating data")
		}
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}
	if rowsAffected == 0 {
		return models.Executive{}, utils.HandleError(fmt.Errorf("executive with ID %d not found", id), "Error updating data")
	}

	var executive models.Executive
	err = db.QueryRow("SELECT id, first_name, last_name, position, email FROM executives WHERE id = ?", id).Scan(&executive.ID, &executive.FirstName, &executive.LastName, &executive.Position, &executive.Email)
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}
	return executive, nil
}

func PatchExecutivesDBHandler(updates []map[string]interface{}) (int64, error) {
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
			return 0, utils.HandleError(err, "Invalid update payload: missing executive ID")
		}

		id, err := strconv.Atoi(fmt.Sprintf("%v", idValue))
		if err != nil {
			tx.Rollback()
			return 0, utils.HandleError(err, "Invalid executive ID")
		}

		var existingExecutive models.Executive
		err = tx.QueryRow("SELECT id, first_name, last_name, position, email FROM executives WHERE id = ?", id).Scan(&existingExecutive.ID, &existingExecutive.FirstName, &existingExecutive.LastName, &existingExecutive.Position, &existingExecutive.Email)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return 0, utils.HandleError(err, "Executive not found")
			}
			return 0, utils.HandleError(err, "Error updating data")
		}

		// Apply updates using reflection
		executiveValue := reflect.ValueOf(&existingExecutive).Elem()

		for field, value := range update {
			if field == "id" {
				continue // Skip ID field
			}
			for i := 0; i < executiveValue.NumField(); i++ {
				structField := executiveValue.Type().Field(i)
				jsonTag := structField.Tag.Get("json")
				jsonField := strings.Split(jsonTag, ",")[0]
				if jsonField == field {
					fieldValue := executiveValue.Field(i)
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

		result, err := tx.Exec("UPDATE executives SET first_name = ?, last_name = ?, position = ?, email = ? WHERE id = ?", existingExecutive.FirstName, existingExecutive.LastName, existingExecutive.Position, existingExecutive.Email, id)
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

func PatchOneExecutive(id int, updates map[string]interface{}) (models.Executive, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}
	defer db.Close()

	var existingExecutive models.Executive
	err = db.QueryRow("SELECT id, first_name, last_name, position, email FROM executives WHERE id = ?", id).Scan(&existingExecutive.ID, &existingExecutive.FirstName, &existingExecutive.LastName, &existingExecutive.Position, &existingExecutive.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Executive{}, utils.HandleError(err, "Error updating data")
		}
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}

	executiveValue := reflect.ValueOf(&existingExecutive).Elem()

	for field, value := range updates {
		for i := 0; i < executiveValue.NumField(); i++ {
			structField := executiveValue.Type().Field(i)
			jsonTag := structField.Tag.Get("json")
			jsonField := strings.Split(jsonTag, ",")[0]
			if jsonField == field {
				fieldValue := executiveValue.Field(i)
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

	result, err := db.Exec("UPDATE executives SET first_name = ?, last_name = ?, position = ?, email = ? WHERE id = ?", existingExecutive.FirstName, existingExecutive.LastName, existingExecutive.Position, existingExecutive.Email, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Executive{}, utils.HandleError(err, "Error updating data")
		}
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}
	if rowsAffected == 0 {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}

	var executive models.Executive
	err = db.QueryRow("SELECT id, first_name, last_name, position, email FROM executives WHERE id = ?", id).Scan(&executive.ID, &executive.FirstName, &executive.LastName, &executive.Position, &executive.Email)
	if err != nil {
		return models.Executive{}, utils.HandleError(err, "Error updating data")
	}
	return executive, nil
}

func DeleteExecutive(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM executives WHERE id = ?", id)
	if err != nil {
		return utils.HandleError(err, "Error deleting executive")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.HandleError(err, "Error deleting executive")
	}
	if rowsAffected == 0 {
		return utils.HandleError(fmt.Errorf("Executive with ID %d not found", id), "Error deleting executive")
	}
	return nil
}

func GetExecutiveByUsername(req models.Executive) (*models.Executive, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	user := &models.Executive{}
	err = db.QueryRow(`SELECT id, first_name, last_name, position, email, username, password, inactive_status FROM executives WHERE username = ?`,
		req.Username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Position, &user.Email, &user.Username, &user.Password, &user.InactiveStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.HandleError(err, "Invalid username or password")
		}
		return nil, utils.HandleError(err, "Error fetching user data")
	}
	return user, nil
}

func UpdatePasswordInDb(id int, currentPassword, newPassword string) (string, error) {
	db, err := ConnectDB()
	if err != nil {
		return "", utils.HandleError(err, "Error retrieving data")
	}
	defer db.Close()

	var username string
	var userPassword string
	var position string
	err = db.QueryRow("SELECT username, password, position FROM executives WHERE id = ?", id).Scan(&username, &userPassword, &position)
	if err != nil {
		return "", utils.HandleError(err, "user not found")
	}

	err = utils.VerifyPassword(currentPassword, userPassword)
	if err != nil {
		return "", utils.HandleError(err, "Current password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return "", utils.HandleError(err, "Error hashing new password")
	}

	_, err = db.Exec("UPDATE executives SET password = ?, password_changed_at = ? WHERE id = ?", hashedPassword, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return "", utils.HandleError(err, "Error updating password")
	}

	tokenString, err := utils.SignToken(id, username, position)
	if err != nil {
		return "", utils.HandleError(err, "Failed to create token after password update")
	}

	return tokenString, nil
}

func ForgotPasswordDbHandler(emailId string) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.HandleError(err, "Database connection error")
	}
	defer db.Close()

	var executive models.Executive
	err = db.QueryRow("SELECT id FROM executives WHERE email = ?", emailId).Scan(&executive.ID)
	if err != nil {
		return utils.HandleError(err, "User not found")
	}

	duration, err := strconv.Atoi(os.Getenv("RESET_PASSWORD_TOKEN_EXPIRES_IN"))
	if err != nil {
		return utils.HandleError(err, "Invalid token expiration duration")
	}

	mins := time.Duration(duration) * time.Minute

	expiry := time.Now().Add(mins).Format(time.RFC3339)

	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return utils.HandleError(err, "Failed to generate reset token")
	}

	token := hex.EncodeToString(tokenBytes)
	hashedToken := sha256.Sum256(tokenBytes)
	hashedTokenStr := hex.EncodeToString(hashedToken[:])

	_, err = db.Exec("UPDATE executives SET password_reset_token = ?, password_token_expires = ? WHERE id = ?", hashedTokenStr, expiry, executive.ID)
	if err != nil {
		return utils.HandleError(err, "Failed to save reset token")
	}

	resetURL := fmt.Sprintf("https://localhost:3000/execs/resetpassword/reset/%s", token)
	message := fmt.Sprintf("You requested a password reset. Click the link to reset your password: %s. Duration is %d minutes.", resetURL, duration)

	m := mail.NewMessage()
	m.SetHeader("From", "schooladmin@school.com")
	m.SetHeader("To", emailId)
	m.SetHeader("Subject", "Your password reset link")
	m.SetBody("text/plain", message)

	d := mail.NewDialer("localhost", 1025, "", "")
	if err := d.DialAndSend(m); err != nil {
		return utils.HandleError(err, "Failed to send reset email")
	}
	return nil
}

func ResetPasswordDbHandler(resetCode, newPassword string) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.HandleError(err, "Database connection error")
	}
	defer db.Close()

	tokenBytes, err := hex.DecodeString(resetCode)
	if err != nil {
		return utils.HandleError(err, "Invalid reset token")
	}
	hashedToken := sha256.Sum256(tokenBytes)
	hashedTokenStr := hex.EncodeToString(hashedToken[:])

	var executive models.Executive
	err = db.QueryRow("SELECT id FROM executives WHERE password_reset_token = ? AND password_token_expires > ?", hashedTokenStr, time.Now().Format(time.RFC3339)).Scan(&executive.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.HandleError(fmt.Errorf("Invalid or expired reset token"), "Invalid or expired reset token")
		}
		return utils.HandleError(err, "Error fetching user for password reset")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return utils.HandleError(err, "Error hashing new password")
	}

	_, err = db.Exec("UPDATE executives SET password = ?, password_changed_at = ?, password_reset_token = NULL, password_token_expires = NULL WHERE id = ?", hashedPassword, time.Now().Format(time.RFC3339), executive.ID)
	if err != nil {
		return utils.HandleError(err, "Error resetting password")
	}
	return nil
}
