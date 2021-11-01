package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/stretchr/testify/assert"
)

var secretPwd = []byte("LongSecret")
var johnDoeKey *crypto.Key
var janeDoeKey *crypto.Key
var steveDoeKey *crypto.Key
var johnArmored string
var janeArmored string
var steveArmored string

func init() {
	theKey, err := helper.GenerateKey("John Doe", "john@doe.com", secretPwd, "rsa", 4096)
	if err != nil {
		panic(err)
	}
	johnArmored = theKey
	johnDoeKey, err = crypto.NewKeyFromArmored(johnArmored)
	if err != nil {
		panic(err)
	}

	janeKeyStr, err := helper.GenerateKey("Jane Doe", "jane@doe.com", secretPwd, "rsa", 4096)

	janeArmored = janeKeyStr
	if err != nil {
		panic(err)
	}
	janeDoeKey, err = crypto.NewKeyFromArmored(janeKeyStr)
	if err != nil {
		panic(err)
	}

	steve, err := helper.GenerateKey("Steve Doe", "Steve@doe.com", secretPwd, "rsa", 4096)

	steveArmored = steve
	if err != nil {
		panic(err)
	}
	steveDoeKey, err = crypto.NewKeyFromArmored(steveArmored)
	if err != nil {
		panic(err)
	}
}

func Test_Encrypt_Decrypt_To_Multiple_Keys(t *testing.T) {
	err := os.Setenv("XLRTE_PASSPHRASE", "LongSecret")
	assert.NoError(t, err)
	envName := RandStringBytes()
	err = os.MkdirAll(filepath.Join("testdata", "environments", envName, "secrets"), 0750)
	assert.NoError(t, err)
	err = writePubKey("testdata", envName, johnDoeKey)
	assert.NoError(t, err)
	err = writePubKey("testdata", envName, janeDoeKey)
	assert.NoError(t, err)

	err = WriteSecret("testdata", envName, "FOO", "foobar")
	assert.NoError(t, err)
	err = WriteSecret("testdata", envName, "BAR", "bazqux")
	assert.NoError(t, err)

	secrets, err := getAllSecretsPrivate(johnArmored, "testdata", envName)
	assert.NoError(t, err)
	assert.Len(t, secrets, 2)

	secrets2, err := getAllSecretsPrivate(janeArmored, "testdata", envName)
	assert.NoError(t, err)

	_, err = getAllSecretsPrivate(steveArmored, "testdata", envName)
	assert.Error(t, err)

	assert.Equal(t, secrets, secrets2)

	assert.Equal(t, *secrets[0], Secret{Name: "BAR", Value: "bazqux"})

	assert.Equal(t, *secrets[1], Secret{Name: "FOO", Value: "foobar"})

	err = os.RemoveAll(filepath.Join("testdata", "environments", envName))
	assert.NoError(t, err)
	err = os.Setenv("XLRTE_PASSPHRASE", "")
	assert.NoError(t, err)
}

func Test_Refresh_Add_One_Key_Remove_Another(t *testing.T) {
	err := os.Setenv("XLRTE_PASSPHRASE", "LongSecret")
	assert.NoError(t, err)
	envName := RandStringBytes()
	err = os.MkdirAll(filepath.Join("testdata", "environments", envName, "secrets"), 0750)
	assert.NoError(t, err)
	err = writePubKey("testdata", envName, johnDoeKey)
	assert.NoError(t, err)
	err = writePubKey("testdata", envName, janeDoeKey)
	assert.NoError(t, err)

	err = WriteSecret("testdata", envName, "FOO", "foobar")
	assert.NoError(t, err)
	err = WriteSecret("testdata", envName, "BAR", "bazqux")
	assert.NoError(t, err)

	secrets, err := getAllSecretsPrivate(johnArmored, "testdata", envName)
	assert.NoError(t, err)
	assert.Len(t, secrets, 2)

	_, err = getAllSecretsPrivate(janeArmored, "testdata", envName)
	assert.NoError(t, err)

	_, err = getAllSecretsPrivate(steveArmored, "testdata", envName)
	assert.Error(t, err)

	err = writePubKey("testdata", envName, steveDoeKey)
	assert.NoError(t, err)
	err = os.Remove(filepath.Join("testdata", "environments", envName, "pubkeys", "jane-doe-jane@doe.com.asc"))
	assert.NoError(t, err)
	err = refreshPrivate(johnArmored, "testdata", envName)
	assert.NoError(t, err)
	stevesSecrets, err := getAllSecretsPrivate(steveArmored, "testdata", envName)
	assert.NoError(t, err)
	assert.Equal(t, stevesSecrets, secrets)

	_, err = getAllSecretsPrivate(janeArmored, "testdata", envName)
	assert.Error(t, err)

	assert.Equal(t, *secrets[0], Secret{Name: "BAR", Value: "bazqux"})

	assert.Equal(t, *secrets[1], Secret{Name: "FOO", Value: "foobar"})

	err = os.RemoveAll(filepath.Join("testdata", "environments", envName))
	assert.NoError(t, err)
	err = os.Setenv("XLRTE_PASSPHRASE", "")
	assert.NoError(t, err)
}

func Test_Encrypt_No_Keys(t *testing.T) {
	err := os.Setenv("XLRTE_PASSPHRASE", "LongSecret")
	assert.NoError(t, err)
	envName := RandStringBytes()
	err = os.MkdirAll(filepath.Join("testdata", "environments", envName, "secrets"), 0750)
	assert.NoError(t, err)

	err = WriteSecret("testdata", envName, "FOO", "foobar")
	assert.Error(t, err)

	err = os.RemoveAll(filepath.Join("testdata", "environments", envName))
	assert.NoError(t, err)
}

func Test_Init_Env(t *testing.T) {
	envName := RandStringBytes()

	testFilesForEnv(t, envName, func() error {
		return initEnv("testdata", envName, johnDoeKey)
	})

}

func Test_Init_Secrets_With_No_Key_File(t *testing.T) {
	envName := RandStringBytes()

	err := os.RemoveAll(filepath.Join("testdata", ".xlrte"))
	assert.NoError(t, err)
	init, err := initSecretsPrivate("testdata", "testdata", envName, func() (*keySettings, error) {
		return &keySettings{
			"john doe", "john@doe.com", "pass", "pass",
		}, nil
	})
	assert.True(t, init)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join("testdata", ".xlrte", "private-key.asc"))
	assert.NoError(t, err)
	init, err = initSecretsPrivate("testdata", "testdata", envName, func() (*keySettings, error) {
		return nil, fmt.Errorf("should never be called")
	})

	assert.False(t, init)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join("testdata", ".xlrte", "private-key.asc"))
	assert.NoError(t, err)

	err = os.RemoveAll(filepath.Join("testdata", ".xlrte"))
	assert.NoError(t, err)
}

func Test_Init_Secrets_With_Key_Set_In_Env(t *testing.T) {
	envName := RandStringBytes()

	err := os.Setenv("XLRTE_PRIVATE_KEY", johnArmored)
	assert.NoError(t, err)

	testFilesForEnv(t, envName, func() error {
		_, e := InitSecrets("testdata", "testdata", envName)
		return e
	})
	err = os.Setenv("XLRTE_PRIVATE_KEY", "")
	assert.NoError(t, err)
}

func testFilesForEnv(t *testing.T, env string, fn func() error) {
	err := os.RemoveAll(filepath.Join("testdata", "environments", env))
	assert.NoError(t, err)
	err = fn()
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join("testdata", "environments", env, "pubkeys", "john-doe-john@doe.com.asc"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join("testdata", "environments", env, "secrets"))
	assert.NoError(t, err)
	err = os.RemoveAll(filepath.Join("testdata", "environments", env))
	assert.NoError(t, err)
}

func Test_RandString(t *testing.T) {
	pastStrings := []string{}

	for i := 1; i < 10; i++ {
		str := RandStringBytes()
		str = strings.ReplaceAll(str, " ", "")
		assert.GreaterOrEqual(t, len(str), 10)
		fmt.Println(str)
		assert.LessOrEqual(t, len(str), 30)

		for _, past := range pastStrings {
			assert.NotEqual(t, past, str)
		}
		pastStrings = append(pastStrings, str)
	}

}
