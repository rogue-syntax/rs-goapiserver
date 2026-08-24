package user

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/database"
)

type UserInternal struct {
	User_id            int
	Email_id           int
	Email_value        string
	Email_verified     int
	User_first_name    string
	User_last_name     string
	User_pw            string
	User_date_of_birth *string
	Kyc_aml_status     int
	Kyc_aml_date       int
	Kyc_aml_id         string
	User_phone         string
	User_role_id       *int
	Role_name          string
}

type UserExternal struct {
	User_id            int
	Email_id           int
	Email_value        string
	Email_verified     int
	User_first_name    string
	User_last_name     string
	User_date_of_birth *string
	Kyc_aml_status     int
	Kyc_aml_date       int
	Kyc_aml_id         string
	User_phone         string
	User_role_id       *int
	Role_name          string
}

func ScanUserInternal(rows *sql.Rows) (*UserInternal, error) {
	var usr UserInternal
	if rows.Next() {
		err := rows.Scan(&usr.User_id,
			&usr.User_first_name,
			&usr.User_last_name,
			&usr.User_date_of_birth,
			&usr.Email_id,
			&usr.Email_value,
			&usr.Email_verified,
			&usr.User_pw,
			&usr.Kyc_aml_status,
			&usr.Kyc_aml_date,
			&usr.Kyc_aml_id,
			&usr.User_phone,
			&usr.User_role_id,
			&usr.Role_name)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &usr, nil
	} else {
		return nil, nil
	}
}

func ScanUserExternal(rows *sql.Rows) (*UserExternal, error) {
	var usr UserExternal
	if rows.Next() {
		err := rows.Scan(&usr.User_id,
			&usr.User_first_name,
			&usr.User_last_name,
			&usr.User_date_of_birth,
			&usr.Email_id,
			&usr.Email_value,
			&usr.Email_verified,
			&usr.Kyc_aml_status,
			&usr.Kyc_aml_date,
			&usr.Kyc_aml_id,
			&usr.User_phone,
			&usr.User_role_id,
			&usr.Role_name)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &usr, nil
	} else {
		return nil, nil
	}
}

func FindUserInternalByEmail(email_value string) (*UserInternal, error) {
	rows, err := database.DB.Query("SELECT * FROM UserInternal WHERE email_value = ?", email_value)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	usr, err := ScanUserInternal(rows)
	return usr, errors.WithStack(err)
}

func FindUserInternalByUser_id(user_id int) (*UserInternal, error) {
	rows, err := database.DB.Query("SELECT * FROM UserInternal WHERE user_id = ?", user_id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	usr, err := ScanUserInternal(rows)
	return usr, errors.WithStack(err)
}

func FindUserExternalByUser_id(user_id int) (*UserExternal, error) {
	rows, err := database.DB.Query("SELECT * FROM UserExternal WHERE user_id = ?", user_id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	usr, err := ScanUserExternal(rows)
	return usr, errors.WithStack(err)
}

func FindApiKeyByUser_id(user_id int) (string, error) {
	var err error
	var apiKeyHash string
	err = database.DB.Get(&apiKeyHash, "SELECT user_api_tok FROM user_auth WHERE user_id = ?", user_id)
	return apiKeyHash, err
}

// UserINternal to UserExternal
//   - Utility function to copy UserInternal Values to UserExternal Vaules
func UserInternalExternal(ui *UserInternal, ux *UserExternal) *UserExternal {
	(*ux).User_id = (*ui).User_id
	(*ux).Email_id = (*ui).Email_id
	(*ux).Email_value = (*ui).Email_value
	(*ux).User_first_name = (*ui).User_first_name
	(*ux).User_last_name = (*ui).User_last_name
	(*ux).Email_verified = (*ui).Email_verified
	(*ux).User_date_of_birth = (*ui).User_date_of_birth
	(*ux).Kyc_aml_status = (*ui).Kyc_aml_status
	(*ux).Kyc_aml_date = (*ui).Kyc_aml_date
	(*ux).User_phone = (*ui).User_phone
	(*ux).User_role_id = (*ui).User_role_id
	(*ux).Role_name = (*ui).Role_name
	return ux
}
