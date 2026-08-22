package auth

import "testing"

func TestCheckPasswordHashWithActualPwdAndHash(t *testing.T) {
	pwd := "thiS$is^mY&$tr0nG_pwD"

	pwdHash, err := HashPassword(pwd)
	if err != nil {
		t.Errorf(`failed to get hash for given password - "%s", error: %v`, pwd, err)
	}

	isValid, err := CheckPasswordHash(pwd, pwdHash)
	if !isValid || err != nil {
		t.Errorf(`CheckPasswordHash("%s", "%s"); expected: true, actual: false`, pwd, pwdHash)
	}

}
func TestCheckPasswordHashWithDifferentPwdAndHash(t *testing.T) {
	pwd := "thiS$is^mY&$tr0nG_pwD"

	pwdHash, err := HashPassword(pwd)
	if err != nil {
		t.Errorf(`failed to get hash for given password - "%s", error: %v`, pwd, err)
	}

	isValid, err := CheckPasswordHash(pwd+"abc", pwdHash)
	if isValid || err != nil {
		t.Errorf(`CheckPasswordHash("%s", "%s"); expected: false, actual: true`, pwd+"abc", pwdHash)
	}
}
