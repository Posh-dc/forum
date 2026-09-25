package backEnd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	tpl = template.Must(template.ParseGlob("./static/*html"))
)

type UserRegInfo struct {
	DisplayName     string `json:"displayName"`
	UserName        string `json:"userName"`
	Email           string `json:"emailAddress"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	PhoneNumber     string `json:"phoneNumber"`
}

var (
	displayName = "displayName"
	userName    = "userName"
	email       = "email"
	password    = "password"
	phoneNumber = "phoneNumber"
)

type DBstruct struct {
	DB *pgxpool.Pool
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tpl.ExecuteTemplate(w, "index.html", nil)

}

func OnboardingHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "onboarding.html", nil)
}

// CHECKING USERNAME AVAILABILITY HANDLER STARTS HERE
func (db *DBstruct) UsernameAvailabilityHandler(w http.ResponseWriter, r *http.Request) {

	data, _ := io.ReadAll(r.Body)

	ctx := r.Context()

	username, err := ValidateUsername(string(data))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userNameExist, err := CheckAlreadyExist("username", username, db.DB, ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userNameExist {
		http.Error(w, "Username Already Exist", http.StatusConflict)
		return
	}

	w.WriteHeader(200)
	fmt.Fprint(w, "Valid Username")

}

// USER ACCOUNT REGISTRATION HANDLER STARTS FROM HERE
func (db *DBstruct) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {

	var user UserRegInfo

	ctx := r.Context()

	decodingErr := json.NewDecoder(r.Body).Decode(&user)

	if decodingErr != nil {
		fmt.Println(decodingErr)

		RegistrationResponse(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	//                        checking user data starts here

	var err error


	// Validate displayname
	user.DisplayName, err = ValidateDisplayName(user.DisplayName)

	if err != nil {
		fmt.Println(err)

		RegistrationResponse(w, http.StatusBadRequest, err.Error(),displayName)
		return
	}

	// Validate username
	user.UserName, err = ValidateUsername(user.UserName)

	if err != nil {
		fmt.Println(err)

		RegistrationResponse(w, http.StatusBadRequest, err.Error(), userName)
		return
	}

	// Validate email
	user.Email, err = ValidateEmail(user.Email)

	if err != nil {
		fmt.Println(err)
		RegistrationResponse(w, http.StatusBadRequest, err.Error(), email)
		return
	}

	// Validate password
	user.Password, err = ValidatePassword(user.Password)

	if err != nil {
		fmt.Println(err)
		RegistrationResponse(w, http.StatusBadRequest, err.Error(), password)
		return
	}

	// Validate phonenumber
	user.PhoneNumber, err = ValidatePhoneNumber(user.PhoneNumber)

	if err != nil {
		fmt.Println(err)
		RegistrationResponse(w, http.StatusBadRequest, err.Error(), phoneNumber)
		return
	}

	//                         checking user data ends here

	//      hashing user password
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		fmt.Println(err)

		// http.Error(w, "something went wrong - password can't be hashed", http.StatusInternalServerError)
		RegistrationResponse(w, http.StatusInternalServerError, err.Error(), "")
		return
	}

	sessionId, err := GenerateSessionId()
	if err != nil {
		fmt.Println(err)

		// fmt.Fprint(w, "something went wrong - sessionId can't be generated")
		RegistrationResponse(w, http.StatusInternalServerError, err.Error(), "")
		return
	}

	//   inserting user details + sesession details into the database
	insertingError := InsertUser(user, sessionId, db.DB, ctx)

	if insertingError != nil {
		errMessage, code, ele := CheckUniqueConstraint(insertingError)

		if code == http.StatusConflict {

			// http.Error(w, errMessage.Error(), code)
			RegistrationResponse(w, code, errMessage.Error(), ele)
			return
		}

		fmt.Println(errMessage)

		// http.Error(w, "Database error", code)
		RegistrationResponse(w, code, errMessage.Error(), "")
		return
	}

	cookieExpTime := time.Now().Add(time.Hour)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  cookieExpTime,
	}

	http.SetCookie(w, cookie)

	RegistrationResponse(w, http.StatusOK, "Account Created", "")

}
