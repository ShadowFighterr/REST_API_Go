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
	"time"
)

func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello GET method in executives routes\n"))

	var execs []models.Executive
	execs, err := sqlconnect.GetExecutivesDbHandler(execs, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execs)
}

func GetOneExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid executive ID", http.StatusBadRequest)
		return
	}

	executive, err := sqlconnect.GetExecutiveByID(w, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello GET method in executives routes\n"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executive)
}

func PostExecsHandler(w http.ResponseWriter, r *http.Request) {
	var newExecs []models.Executive
	var rawExecs []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawExecs)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	val := reflect.TypeOf(models.Executive{})
	allowedFields := make(map[string]struct{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
		allowedFields[fieldToAdd] = struct{}{}
	}

	for _, executive := range rawExecs {
		for key := range executive {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, fmt.Sprintf("Invalid field: %s", key), http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newExecs)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	for _, executive := range newExecs {
		val := reflect.ValueOf(executive)
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.Kind() == reflect.String && field.String() == "" {
				http.Error(w, "All fields must be filled", http.StatusBadRequest)
				return
			}
		}
	}

	addedExecs, err := sqlconnect.AddExecutivesDBHandler(newExecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string             `json:"status"`
		Count  int                `json:"count"`
		Data   []models.Executive `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(response)
}

// Patch /executives/
func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello PATCH method in executives routes\n"))
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	rowsAffected, err := sqlconnect.PatchExecutivesDBHandler(updates)
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
		Message:      "Executives updated successfully",
		RowsAffected: rowsAffected,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Patch /executives/{id}
func PatchOneExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Executive ID is required for update", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid executive ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedExecutive, err := sqlconnect.PatchOneExecutive(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Hello 1 PATCH method in executives routes\n"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedExecutive)

}

func DeleteExecHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Hello DELETE method in executives routes\n"))
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Executive ID is required for deletion", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid executive ID", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteExecutive(id)
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
		Message: fmt.Sprintf("Executive with ID %d deleted successfully", id),
	}
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Executive
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := sqlconnect.GetExecutiveByUsername(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if user.InactiveStatus {
		http.Error(w, "Account is inactive. Please contact support.", http.StatusForbidden)
		return
	}

	err = utils.VerifyPassword(req.Password, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tokenString, err := utils.SignToken(user.ID, user.Username, user.Position)
	if err != nil {
		utils.HandleError(err, "Error generating token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	}{
		Status: "success",
		Token:  tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "Logged out successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Executive ID is required for password update", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid executive ID", http.StatusBadRequest)
		return
	}

	var req models.UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Current password and new password are required", http.StatusBadRequest)
		return
	}

	tokenString, err := sqlconnect.UpdatePasswordInDb(id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour)})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Token   string `json:"token"`
	}{
		Status:  "success",
		Message: "Password updated successfully",
		Token:   tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	err = sqlconnect.ForgotPasswordDbHandler(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "If an account with that email exists, a password reset link has been sent.",
	}
	json.NewEncoder(w).Encode(response)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	resetCode := r.PathValue("resetcode")
	if resetCode == "" {
		http.Error(w, "Reset code is required", http.StatusBadRequest)
		return
	}

	var req struct {
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	if req.NewPassword == "" || req.ConfirmPassword == "" {
		http.Error(w, "New password and confirm password are required", http.StatusBadRequest)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		http.Error(w, "New password and confirm password do not match", http.StatusBadRequest)
		return
	}

	err = sqlconnect.ResetPasswordDbHandler(resetCode, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "Password has been reset successfully",
	}
	json.NewEncoder(w).Encode(response)
}
