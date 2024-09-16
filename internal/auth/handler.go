package auth

import (
	"assessment/dto"
	"assessment/internal/database"
	"assessment/models"
	"assessment/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
)

var validate = validator.New()

func HandleLogin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginUser dto.UserRequest
	err := json.NewDecoder(r.Body).Decode(&loginUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	result := database.DB.Where("email = ?", loginUser.Email).First(&user)
	if result.Error != nil {
		http.Error(w, "Invalid email", http.StatusUnauthorized)
		return
	}

	if !utils.ComparePasswords(user.Password, []byte(loginUser.Password)) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// todo : add doc comments
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//request mapping
	var UserRequest dto.UserRequest
	err := json.NewDecoder(r.Body).Decode(&UserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// request validation
	err = validate.Struct(UserRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	var existingUser models.User
	res := database.DB.Where("email = ?", UserRequest.Email).First(&existingUser)
	if res.Error == nil {
		http.Error(w, "Email already exists", http.StatusBadRequest)
		return
	}

	// password hashing
	hashedPassword, err := utils.HashPassword(UserRequest.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// persistance
	newUser := models.User{
		Email:    UserRequest.Email,
		Password: hashedPassword,
		Role:     "user", // todo : use defined const enums for this
	}

	result := database.DB.Create(&newUser)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser) //case
}
