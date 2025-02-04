package apathy

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetAwd(t *testing.T) {
	pwd, err := os.Getwd()
	assert.Nil(t, err)

	awd, err := GetAwd()
	assert.Nil(t, err, "got an error calling GetAwd()")

	posixPwd := strings.ReplaceAll(pwd, "\\", "/")
	assert.Equal(t, posixPwd, awd.String())
	assert.Equal(t, pwd, awd.ToNative())
}
