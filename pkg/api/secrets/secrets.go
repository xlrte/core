package secrets

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"

	"github.com/AlecAivazis/survey/v2"
)

type keySettings struct {
	Name              string //
	Email             string
	Passphrase        string
	ConfirmPassphrase string
}

type Secret struct {
	Name  string
	Value string
}

const intBytes = "0123456789"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

func RandStringBytes() string {
	len := 0
	for len < 10 || len > 30 {
		len = rand.Intn(30) //nolint
	}
	return randStringBytesOfLength(len)
}

func RandIntBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = intBytes[rand.Intn(len(intBytes))] //nolint
	}
	return string(b)
}

func randStringBytesOfLength(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))] //nolint
	}
	return string(b)
}

func AddSecret(baseDir, env, secretName string) error {
	path := filepath.Join(baseDir, "environments", env, "secrets", fmt.Sprintf("%s.asc", secretName))
	_, err := os.Stat(path)
	if err == nil {
		overwrite := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("The secret %s/%s already exists, do you want to overwrite it?", env, secretName),
		}
		err = survey.AskOne(prompt, &overwrite)
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println("cancelling secret creation")
			return nil
		}
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}
	secretValue := ""
	prompt := &survey.Password{Message: "Please enter the value of the secret:"}
	err = survey.AskOne(prompt, &secretValue)
	if err != nil {
		return err
	}

	return WriteSecret(baseDir, env, secretName, secretValue)

}

func GetAllSecrets(homeDir, baseDir, env string) ([]*Secret, error) {
	_, armoredKey, err := GetPrivateKey(homeDir)
	if err != nil {
		return nil, err
	}
	return getAllSecretsPrivate(armoredKey, baseDir, env)
}

func getAllSecretsPrivate(armoredKey, baseDir, env string) ([]*Secret, error) {
	if os.Getenv("XLRTE_PASSPHRASE") == "" {
		defer func() {
			_ = os.Setenv("XLRTE_PASSPHRASE", "")
		}()
	}
	pubKeyDir := filepath.Join(baseDir, "environments", env, "secrets")
	allSecrets := []*Secret{}
	err := filepath.Walk(pubKeyDir, func(path string, info os.FileInfo, e error) error {
		if !strings.HasSuffix(path, ".asc") {
			return nil
		}
		data, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		pass, err := getPassphrase()
		if err != nil {
			return err
		}
		clearText, err := helper.DecryptMessageArmored(armoredKey, []byte(pass), string(data))
		if err != nil {
			return fmt.Errorf("decryption failed, did you enter the correct passphrase? %w", err)
		}
		_, file := filepath.Split(path)
		allSecrets = append(allSecrets, &Secret{
			Name:  strings.TrimSuffix(file, ".asc"),
			Value: clearText,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allSecrets, nil
}

func Refresh(homeDir, baseDir, env string) error {
	_, armoredKey, err := GetPrivateKey(homeDir)
	if err != nil {
		return err
	}
	return refreshPrivate(armoredKey, baseDir, env)
}

func refreshPrivate(armoredKey, baseDir, env string) error {
	secrets, err := getAllSecretsPrivate(armoredKey, baseDir, env)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		path := filepath.Join(baseDir, "environments", env, "secrets", fmt.Sprintf("%s.asc", secret.Name))
		err = os.Remove(path)
		if err != nil {
			return err
		}
		err = WriteSecret(baseDir, env, secret.Name, secret.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteSecret(baseDir, env, secretName, secretValue string) error {
	path := filepath.Join(baseDir, "environments", env, "secrets", fmt.Sprintf("%s.asc", secretName))

	keys, err := getPublicKeys(baseDir, env)
	if err != nil {
		return err
	}
	var keyRing *crypto.KeyRing
	for _, key := range keys {
		if keyRing == nil {
			keyRing, err = crypto.NewKeyRing(key)
			if err != nil {
				return err
			}
		} else {
			err = keyRing.AddKey(key)
			if err != nil {
				return err
			}
		}
	}
	if keyRing == nil {
		return fmt.Errorf("no public keys for environment %s", env)
	}

	sessionKey, err := crypto.GenerateSessionKey()
	if err != nil {
		return err
	}

	message := crypto.NewPlainMessage([]byte(secretValue))
	dataPacket, err := sessionKey.Encrypt(message)
	if err != nil {
		return err
	}
	keyPacket, err := keyRing.EncryptSessionKey(sessionKey)
	if err != nil {
		return err
	}

	pgpSplitMessage := crypto.NewPGPSplitMessage(keyPacket, dataPacket)
	pgpMessage := pgpSplitMessage.GetPGPMessage()

	ascString, err := pgpMessage.GetArmored()
	if err != nil {
		return err
	}
	err = os.WriteFile(path, []byte(ascString), 0600)
	return err
}

func InitSecrets(homeDir, baseDir, env string) (bool, error) {
	return initSecretsPrivate(homeDir, baseDir, env, initSettings)
}

func initSecretsPrivate(homeDir, baseDir, env string, initFn func() (*keySettings, error)) (bool, error) {
	initialized := false
	privateKey, _, e := GetPrivateKey(homeDir)
	if e != nil {
		ascKey := ""
		targetDir := filepath.Join(homeDir, ".xlrte")
		targetFile := filepath.Join(targetDir, "private-key.asc")
		err := os.MkdirAll(targetDir, 0750)
		if err != nil {
			return initialized, err
		}
		settings, err := initFn()
		if err != nil {
			return initialized, err
		}
		ascKey, err = helper.GenerateKey(settings.Name, settings.Email, []byte(settings.Passphrase), "rsa", 4096)
		if err != nil {
			return initialized, err
		}

		_, err = os.Create(targetFile)
		if err != nil {
			return initialized, err
		}
		err = os.WriteFile(targetFile, []byte(ascKey), 0600)
		if err != nil {
			return initialized, err
		}
		privateKey, err = crypto.NewKeyFromArmored(ascKey)
		if err != nil {
			return initialized, err
		}
		initialized = true
	}

	e = initEnv(baseDir, env, privateKey)
	if e != nil {
		return false, e
	}
	return initialized, nil
}

func GetPrivateKey(homeDir string) (*crypto.Key, string, error) {
	ascKey := ""
	targetDir := filepath.Join(homeDir, ".xlrte")

	targetFile := filepath.Join(targetDir, "private-key.asc")
	if os.Getenv("XLRTE_PRIVATE_KEY") != "" {
		ascKey = os.Getenv("XLRTE_PRIVATE_KEY")
	} else {
		data, err := ioutil.ReadFile(filepath.Clean(targetFile))
		if err != nil {
			return nil, "", err
		}
		ascKey = string(data)
	}
	theKey, err := crypto.NewKeyFromArmored(ascKey)
	if err != nil {
		return nil, "", err
	}
	return theKey, ascKey, nil
}

func getPublicKeys(baseDir, env string) ([]*crypto.Key, error) {
	pubKeyDir := filepath.Join(baseDir, "environments", env, "pubkeys")
	allKeys := []*crypto.Key{}
	err := filepath.Walk(pubKeyDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".asc") {
			data, err := ioutil.ReadFile(filepath.Clean(path))
			if err != nil {
				return err
			}
			key, err := crypto.NewKeyFromArmored(string(data))
			if err != nil {
				return err
			}
			allKeys = append(allKeys, key)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allKeys, nil
}

func initEnv(baseDir, env string, key *crypto.Key) error {
	envDir := filepath.Join(baseDir, "environments", env)
	secretsDir := filepath.Join(envDir, "secrets")

	err := os.MkdirAll(secretsDir, 0750)
	if err != nil {
		return err
	}
	return writePubKey(baseDir, env, key)

}

func writePubKey(baseDir, env string, key *crypto.Key) error {
	envDir := filepath.Join(baseDir, "environments", env)
	keyDir := filepath.Join(envDir, "pubkeys")
	err := os.MkdirAll(keyDir, 0750)
	if err != nil {
		return err
	}

	identity := identityStr(key)

	pubKeyFile := filepath.Join(keyDir, identity)
	_, err = os.Stat(pubKeyFile)
	if !os.IsNotExist(err) {
		err = os.Remove(pubKeyFile)
		if err != nil {
			return err
		}
	}

	pubKey, err := key.GetArmoredPublicKey()
	if err != nil {
		return err
	}
	err = os.WriteFile(pubKeyFile, []byte(pubKey), 0600)
	if err != nil {
		return err
	}
	return nil
}

func identityStr(key *crypto.Key) string {
	identity := ""
	for k := range key.GetEntity().Identities {
		identity = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(k, "<", ""), ">", ""), " ", "-"))

	}
	return identity + ".asc"
}

func getPassphrase() (string, error) {
	if os.Getenv("XLRTE_PASSPHRASE") != "" {
		return os.Getenv("XLRTE_PASSPHRASE"), nil
	}
	secretValue := ""
	prompt := &survey.Password{Message: "Please enter your private key passphrase:"}
	err := survey.AskOne(prompt, &secretValue)
	if err != nil {
		return "", err
	}
	err = os.Setenv("XLRTE_PASSPHRASE", secretValue)
	if err != nil {
		return "", err
	}
	return secretValue, nil
}

func initSettings() (*keySettings, error) {
	captured := ""

	var qs = []*survey.Question{
		{
			Name:      "Name",
			Prompt:    &survey.Input{Message: "What is your name?"},
			Validate:  survey.Required,
			Transform: survey.Title,
		},
		{
			Name:      "Email",
			Prompt:    &survey.Input{Message: "What is your e-mail?"},
			Validate:  survey.Required,
			Transform: survey.Title,
		},
		{
			Name:   "Passphrase",
			Prompt: &survey.Password{Message: "Please enter a passphrase for your Private Key:"},
			Validate: func(val interface{}) error {
				captured = fmt.Sprintf("%v", val)
				if captured == "" {
					return fmt.Errorf("passphrase must be set. \nIf you do not want to repeatedly enter it, set it as the environment variable XLRTE_PASSPHRASE")
				}
				return nil
			},
		},
		{
			Name:   "ConfirmPassphrase",
			Prompt: &survey.Password{Message: "Please confirm your passphrase:"},
			Validate: func(val interface{}) error {
				if val != captured {
					return fmt.Errorf("passphrase does not match")
				}
				return nil
			},
		},
	}

	answers := keySettings{}

	fmt.Println("Initializing the secrets-system, please provide the following details for your private key")
	err := survey.Ask(qs, &answers)
	if err != nil {
		return nil, err
	}
	return &answers, nil
}
