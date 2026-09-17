package backEnd

import (
	"encoding/json"
	"fmt"
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

type DBstruct struct {
	DB *pgxpool.Pool
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tpl.ExecuteTemplate(w, "index.html", nil)

}

func OnboardingHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "onboarding.html", nil)
}

// USER ACCOUNT REGISTRATION HANDLER STARTS FROM HERE
func (db *DBstruct) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {

	var user UserRegInfo

	ctx := r.Context()

	decodingErr := json.NewDecoder(r.Body).Decode(&user)

	if decodingErr != nil {
		fmt.Println(decodingErr)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//                        checking user data starts here

	var err error

	fmt.Println(user.DisplayName, user.UserName, user.Email, user.Password, user.PhoneNumber)

	// Validate displayname
	user.DisplayName, err = ValidateDisplayName(user.DisplayName)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate username
	user.UserName, err = ValidateUsername(user.UserName)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate email
	user.Email, err = ValidateEmail(user.Email)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate password
	user.Password, err = ValidatePassword(user.Password)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate phonenumber
	user.PhoneNumber, err = ValidatePhoneNumber(user.PhoneNumber)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//                         checking user data ends here

	//      hashing user password
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong - password can't be hashed", http.StatusInternalServerError)
		return
	}

	sessionId, err := GenerateSessionId()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "something went wrong - sessionId can't be generated")
		return
	}

	//   inserting user details + sesession details into the database
	insertingError := InsertUser(user, sessionId, db.DB, ctx)

	if insertingError != nil {
		errMessage, code := CheckUniqueConstraint(insertingError)

		if code == http.StatusConflict {
			http.Error(w, errMessage.Error(), code)
			return
		}

		fmt.Println(errMessage)
		http.Error(w, "Database error", code)
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
	w.WriteHeader(200)
	// tpl.ExecuteTemplate(w, "index.html", nil)
	fmt.Fprint(w, "account created")

}
