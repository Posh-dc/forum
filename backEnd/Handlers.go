package backEnd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5"
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

type LoginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

		RegistrationResponse(w, http.StatusBadRequest, err.Error(), displayName)
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
	//        hash session ID
	hashedSessionId := GeneralHashFunction(sessionId)

	//   inserting user details + sesession details into the database
	insertingError := InsertUser(user, hashedSessionId, db.DB, ctx)

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

	sessionExpTime := time.Now().Add(24 * time.Hour)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  sessionExpTime,
	}

	http.SetCookie(w, cookie)

	RegistrationResponse(w, http.StatusOK, "Account Created", "")

}

// VERIFYING EMAIL VERIFICATION CODE FROM USER HANDLER STARTS HERE
func (db *DBstruct) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {

	code := r.FormValue("1") + r.FormValue("2") + r.FormValue("3") + r.FormValue("4") + r.FormValue("5") + r.FormValue("6")

	if len(code) != 6 {
		http.Error(w, "Invalid verification code", http.StatusBadRequest)
		return
	}

	for _, r := range code {
		if r < '0' || r > '9' {
			http.Error(w, "Invalid verification code", http.StatusBadRequest)
			return
		}
	}

	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hashedSession := GeneralHashFunction(sessionCookie.Value)

	sessionDetails, err := GetUserSessionDetails(r.Context(), db.DB, hashedSession)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Unathorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if time.Now().After(sessionDetails.ExpiresAt) {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	userDetails, err := GetUserDetails(r.Context(), db.DB, "id", sessionDetails.UserID)

	if err != nil {
		fmt.Println(err)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Unathorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if userDetails.EmailVerified {
		http.Error(w, "Email Already Verified", 201)
		return
	}

	emailDetails, err := GetUserEmailVerificationDetails(r.Context(), db.DB, userDetails.UserID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if emailDetails.Attempts >= 5 {
		http.Error(w, "More than five Attempts, Click Resend code", http.StatusTooManyRequests)
		return
	}

	if time.Now().After(emailDetails.ExpiresAt) {
		http.Error(w, "Verification code expired Click Resend code", http.StatusGone)
		return
	}

	if !VerifyGeneralHash(code, emailDetails.CodeHash) {

		_, err := db.DB.Exec(r.Context(), `
        UPDATE email_verification_codes
        SET attempts = attempts + 1
        WHERE id = $1
		AND attempts < 5
    `, emailDetails.ID)

		if err != nil {
			fmt.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, "Invalid verification code", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(r.Context())

	tag, err := tx.Exec(r.Context(), `
    UPDATE email_verification_codes
    SET used_at = NOW()
    WHERE id = $1
      AND used_at IS NULL
      AND expires_at > NOW()
  `, emailDetails.ID)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if tag.RowsAffected() != 1 {
		http.Error(w, "Verification code is no longer valid", http.StatusGone)
		return
	}

	_, err = tx.Exec(r.Context(), `
      UPDATE users
      SET is_email_verified = true
       WHERE id = $1
    `, emailDetails.UserID)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Email verified ")
}

// USER ACCOUNT LOGIN HANDLER STARTS HERE
func (db *DBstruct) LoginAccountHandler(w http.ResponseWriter, r *http.Request) {

	var details LoginInfo

	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&details)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid Credentils", http.StatusBadRequest)
		return
	}

	//        validate email
	details.Email, err = ValidateEmail(details.Email)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid Credentils", http.StatusBadRequest)
		return
	}

	//        get user details
	userDetails, err := GetUserDetails(ctx, db.DB, "email", details.Email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//        password look up
	if !VerifyPassword(details.Password, userDetails.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionId, err := GenerateSessionId()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	//        hash session ID
	hashedSessionID := GeneralHashFunction(sessionId)

	sessionExpTime := time.Now().Add(24 * time.Hour)

	//        saving session into database
	_, err = db.DB.Exec(ctx, `INSERT INTO session(id, user_id, expires_at) VALUES($1,$2,$3);`,
		hashedSessionID, userDetails.UserID, sessionExpTime)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  sessionExpTime,
	}

	http.SetCookie(w, cookie)

	fmt.Fprint(w, "Login successful")

}

/*                            PAGES HANDLERS              */

// home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "index.html", nil)
}

// onboarding page
func OnboardingHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "onboarding.html", nil)
}

// email verification page
func (db *DBstruct) VerifyEmailPageHandler(w http.ResponseWriter, r *http.Request) {

	type VerifyEmailPageData struct {
		Email     string
		ExpiresAt time.Time
	}

	var data VerifyEmailPageData

	fmt.Println("INSIDE HERE")

	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	hashedSession := GeneralHashFunction(sessionCookie.Value)

	sessionDetails, err := GetUserSessionDetails(r.Context(), db.DB, hashedSession)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	if time.Now().After(sessionDetails.ExpiresAt) {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	userDetails, err := GetUserDetails(r.Context(), db.DB, "id", sessionDetails.UserID)

	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	if userDetails.EmailVerified {
		fmt.Println("verified?")
		http.Redirect(w, r, "/", 303)

		return
	}

	emailDetails, err := GetUserEmailVerificationDetails(r.Context(), db.DB, userDetails.UserID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data.ExpiresAt = emailDetails.ExpiresAt
	data.Email = MaskEmail(userDetails.Email)

	err = tpl.ExecuteTemplate(w, "verify-email.html", data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
