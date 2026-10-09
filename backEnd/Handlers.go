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
	"github.com/jackc/pgx/v5/pgconn"
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

	emailDetails, err := GetUserEmailVerificationDetails(r.Context(), db.DB, "user_id", userDetails.UserID)

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
    DELETE FROM email_verification_codes
    WHERE id = $1
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

// RESEND VERIFICATION CODE EMAIL  STARTS HERE
func (db *DBstruct) ResendEmailHandler(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, "Email Already Verified", http.StatusConflict)
		return
	}

	emailDetails, err := GetUserEmailVerificationDetails(r.Context(), db.DB, "user_id", userDetails.UserID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if time.Since(emailDetails.CreatedAT) < time.Minute {
		http.Error(w, "Please wait before requesting another code", http.StatusTooManyRequests)
		return
	}

	tx, err := db.DB.Begin(r.Context())

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(r.Context())

	// generating and saving verification code

	code, err := GenerateVerificationCode()

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	codeHashed := GeneralHashFunction(code)

	now := time.Now()

	// inserting verification code

	tag, err := tx.Exec(r.Context(), `UPDATE email_verification_codes
	    SET code_hash = $1,
          attempts = 0,
          created_at = NOW(),
          expires_at = $2
		WHERE id = $3 AND created_at < NOW() - INTERVAL '1 minute'
		`, codeHashed, now.Add(10*time.Minute), emailDetails.ID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if tag.RowsAffected() != 1 {
		http.Error(w, "Please wait before requesting another code", http.StatusTooManyRequests)
		return
	}

	// email_outbox starts here

	_, err = tx.Exec(r.Context(), `INSERT INTO email_outbox(recipient_email, subject, body) VALUES($1,$2,$3)`,
		userDetails.Email,
		EmailVerificationSubject,
		CreateVerificationEmailBody(code),
	)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.Commit(r.Context())

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Code has been resend")
}

// CHANGING USER EMAIL HANDLER STARTS HERE
func (db *DBstruct) ChangeEmailHandler(w http.ResponseWriter, r *http.Request) {

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
		fmt.Println(err)
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
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if userDetails.EmailVerified {
		http.Error(w, "Email Already Verified", http.StatusConflict)
		return
	}

	//   validating the new email
	newEmail := r.FormValue("emailAddress")
	fmt.Println(newEmail)

	if newEmail == userDetails.Email {
		http.Error(w, "This is already your email address", http.StatusConflict)
		return
	}

	newEmail, err = ValidateEmail(newEmail)

	if err != nil {
		fmt.Println("Validating email err : ", err)
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	//   validating the new email
	err = ChangeUserEmail(r.Context(), db.DB, userDetails.UserID, newEmail)

	if err != nil {

		var pgErr *pgconn.PgError

		if err.Error() == "user not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "users_email_key" {
				http.Error(w, "Email Already Exist", http.StatusConflict)
				return
			}
		}
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Email changed")
}

// FORGERT PASSWORD HANDLER STARTS HERE
func (db *DBstruct) ForgetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	newEmail := r.FormValue("emailAddress")
	fmt.Println(newEmail)

	newEmail, err := ValidateEmail(newEmail)
	if err != nil {
		fmt.Println()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userInfo, err := GetUserDetails(r.Context(), db.DB, "email", newEmail)

	if err != nil {
		fmt.Println(err)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "If an account exists for this email, a password reset link has been sent.", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	code, err := GenerateRandomToken()

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token := GeneralHashFunction(code)
	now := time.Now()
	ctx := r.Context()
	tx, err := db.DB.Begin(ctx)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(ctx)

	//  invalidating previous password reset link by deleting it from the table
	_, err = tx.Exec(ctx, `DELETE FROM email_verification_codes WHERE user_id = $1 AND purpose = $2`,
		userInfo.UserID,
		"password_reset",
	)

	if err != nil {
		fmt.Println("could not insert into email verification code/link", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// inserting verification code
	_, err = tx.Exec(ctx, `INSERT INTO email_verification_codes(user_id, code_hash, expires_at, purpose) VALUES($1,$2,$3,$4)`,
		userInfo.UserID,
		token,
		now.Add(10*time.Minute),
		"password_reset",
	)

	if err != nil {
		fmt.Println("could not insert into email verification code/link", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// email_outbox starts here

	_, err = tx.Exec(ctx, `INSERT INTO email_outbox(recipient_email, subject, body) VALUES($1,$2,$3)`,
		userInfo.Email,
		"Rest your account password\r\n",
		CreateResetVericationToken(userInfo.UserName, "http://localhost:9090/change-password?token="+code),
	)

	if err != nil {
		fmt.Println("could not insert into outbox", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tx.Commit(ctx)

	if err != nil {
		fmt.Println("could not comomit transaction", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	email := MaskEmail(userInfo.Email)
	fmt.Fprint(w, email)

}

// CHANGE PASSWORD HANGLER STARTS HERE
func (db *DBstruct) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {

	password := r.FormValue("password")

	fmt.Println(string(password))
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

	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	hashedSession := GeneralHashFunction(sessionCookie.Value)

	sessionDetails, err := GetUserSessionDetails(r.Context(), db.DB, hashedSession)
	if err != nil {
		fmt.Println("Getting session details error: ", err)
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	if time.Now().After(sessionDetails.ExpiresAt) {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	userDetails, err := GetUserDetails(r.Context(), db.DB, "id", sessionDetails.UserID)

	if err != nil {
		fmt.Println("Getting user details error: ", err)
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	if userDetails.EmailVerified {
		fmt.Println("verified?")
		http.Redirect(w, r, "/", 303)

		return
	}

	emailDetails, err := GetUserEmailVerificationDetails(r.Context(), db.DB, "user_id", userDetails.UserID)

	if err != nil {
		fmt.Println("Getting email verification details error: ", err)
		http.Redirect(w, r, "/onboarding", 303)
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

// page for changing email page
func (db *DBstruct) ChangeEmailPageHandler(w http.ResponseWriter, r *http.Request) {

	type VerifyEmailPageData struct {
		Email string
		// ExpiresAt time.Time
	}

	var data VerifyEmailPageData

	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	hashedSession := GeneralHashFunction(sessionCookie.Value)

	sessionDetails, err := GetUserSessionDetails(r.Context(), db.DB, hashedSession)
	if err != nil {
		fmt.Println("Getting session details error: ", err)
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	if time.Now().After(sessionDetails.ExpiresAt) {
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	userDetails, err := GetUserDetails(r.Context(), db.DB, "id", sessionDetails.UserID)

	if err != nil {
		fmt.Println("Getting user details error: ", err)
		http.Redirect(w, r, "/onboarding", 303)
		return
	}

	if userDetails.EmailVerified {
		http.Redirect(w, r, "/", 303)
		return
	}

	data.Email = MaskEmail(userDetails.Email)

	err = tpl.ExecuteTemplate(w, "change-email.html", data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// forget password page
func ForgetPasswordPageHandler(w http.ResponseWriter, r *http.Request) {

	err := tpl.ExecuteTemplate(w, "forget-password.html", nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// forget password page
func (db *DBstruct) ChangePasswordPageHandler(w http.ResponseWriter, r *http.Request) {

	type OutputData struct {
		ExpiresAt time.Time
	}

	var data OutputData
	code := r.URL.Query().Get("token")

	codeHash := GeneralHashFunction(code)

	emailVerificationDetails, err := GetUserEmailVerificationDetails(r.Context(), db.DB, "code_hash", codeHash)
	if err != nil {
		fmt.Println(err)
		tpl.ExecuteTemplate(w, "change-error.html", nil)
		return
	}

	if time.Since(emailVerificationDetails.ExpiresAt) > 10*time.Minute {
		tpl.ExecuteTemplate(w, "change-error.html", nil)
		return
	}

	data.ExpiresAt = emailVerificationDetails.ExpiresAt

	tpl.ExecuteTemplate(w, "change-password.html", data)

}
