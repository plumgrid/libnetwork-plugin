package rest

import (
	"fmt"
)

// Login to platform. Cookies are taken care of from the underlay
func AttemptLogin(handle *PgRestHandle) error {
	login_path := getLoginPath(handle)
	login_data := getLoginData(handle)

	err, status_code, _ := RestPost(handle, login_path, login_data)
	if err != nil {
		return fmt.Errorf("Error while logging in: %v", err)
	} else if status_code != 200 {
		return fmt.Errorf("Failed to login: %d", status_code)
	}

	// Set-cookie header should have been captured already by the cookiejar
	// Client will remain open with the cookie set (no keep-alive header though)

	return nil
}

// Logout from platform
func AttemptLogout(handle *PgRestHandle) error {
	logout_path := getLogoutPath(handle)
	logout_data := getLogoutData(handle)

	err, status_code, _ := RestPost(handle, logout_path, logout_data)
	if err != nil {
		return fmt.Errorf("Error while logging out: %v", err)
	} else if status_code != 200 {
		return fmt.Errorf("Failed to logout: %d", status_code)
	}

	return nil
}

// PG Rest Path
func GetRestPath(handle *PgRestHandle) string {
	if handle.Port != 0 {
		return fmt.Sprintf("https://%s:%d", handle.Ip_or_host, handle.Port)
	} else {
		return fmt.Sprintf("https://%s", handle.Ip_or_host)
	}
}

// Login Path
func getLoginPath(handle *PgRestHandle) string {
	return GetRestPath(handle) + "/0/login"
}

// Get login data
func getLoginData(handle *PgRestHandle) string {
	return fmt.Sprintf("{'userName': '%s', 'password': '%s'}", handle.User, handle.Password)
}

// Logout Path
func getLogoutPath(handle *PgRestHandle) string {
	return GetRestPath(handle) + "/0/logout"
}

// Get logout data
func getLogoutData(handle *PgRestHandle) string {
	return fmt.Sprintf("{}")
}
