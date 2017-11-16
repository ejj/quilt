package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	cliPath "github.com/kelda/kelda/cli/path"
	"github.com/kelda/kelda/util"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

func TestDefaultKeys(t *testing.T) {
	util.AppFs = afero.NewMemMapFs()

	// Don't pull in keys from the host OS. Setting this environment variable
	// is safe because it won't affect the parent shell.
	os.Setenv("SSH_AUTH_SOCK", "")

	dir, err := homedir.Dir()
	assert.NoError(t, err, "Failed to get homedir")

	sshDir := filepath.Join(dir, ".ssh")
	err = util.AppFs.MkdirAll(sshDir, 0600)
	assert.NoError(t, err, "Failed to create SSH directory")

	for _, key := range []string{"id_rsa", "id_dsa", "ignored"} {
		err := writeRandomKey(filepath.Join(sshDir, key), false)
		assert.NoError(t, err, "Failed to write key")
	}

	err = util.AppFs.MkdirAll(filepath.Dir(cliPath.DefaultSSHKeyPath), 0600)
	assert.NoError(t, err)

	err = writeRandomKey(cliPath.DefaultSSHKeyPath, false)
	assert.NoError(t, err)

	signers := defaultSigners()
	assert.Len(t, signers, 3)
}

func TestEncryptedKey(t *testing.T) {
	util.AppFs = afero.NewMemMapFs()

	dir, err := homedir.Dir()
	assert.NoError(t, err, "Failed to get homedir")

	sshDir := filepath.Join(dir, ".ssh")
	err = util.AppFs.MkdirAll(sshDir, 0600)
	assert.NoError(t, err, "Failed to create SSH directory")

	keyPath := filepath.Join(sshDir, "key")
	err = writeRandomKey(keyPath, true)
	assert.NoError(t, err, "Failed to write key")

	_, err = signerFromFile(keyPath)

	assert.Error(t, err, "ssh: password protected keys are "+
		"not supported, try adding the key to ssh-agent first using "+
		"`ssh-add`")
}

func writeRandomKey(path string, encrypt bool) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if encrypt {
		// Generate a random passphrase to encrypt the key
		passphrase := make([]byte, 10)
		_, err := rand.Read(passphrase)
		if err != nil {
			return err
		}

		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes,
			passphrase, x509.PEMCipherAES256)
		if err != nil {
			return err
		}
	}

	f, err := util.AppFs.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return pem.Encode(f, block)
}
