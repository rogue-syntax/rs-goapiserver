package user

import (
	"rs-apiserver.com/database"
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
}

func FindUserInternalByEmail(email_value string) (*UserInternal, error) {
	var err error
	var usr UserInternal
	err = database.DB.Get(&usr, "SELECT * FROM UserInternal WHERE email_value = ?", email_value)
	return &usr, err
}

func FindUserInternalByUser_id(user_id int) (*UserInternal, error) {
	var err error
	var usr UserInternal
	err = database.DB.Get(&usr, "SELECT * FROM UserInternal WHERE user_id = ?", user_id)
	return &usr, err
}

func FindUserExternalByUser_id(user_id int) (*UserExternal, error) {
	var err error
	var usr UserExternal
	err = database.DB.Get(&usr, "SELECT * FROM UserExternal WHERE user_id = ?", user_id)
	return &usr, err
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
	return ux
}
