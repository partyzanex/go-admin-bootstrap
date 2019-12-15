package encrypt_test

import (
	"github.com/partyzanex/go-admin-bootstrap/encrypt"
	"github.com/partyzanex/testutils"
	"testing"
)

func TestEncrypt_Encrypt(t *testing.T) {
	key, iv := encrypt.KeysFromString("qwerty")

	e, err := encrypt.New("aes-256-cbc", key, iv)
	testutils.Err(t, "encrypt.New", err)

	exp := "test"
	out, err := e.Encrypt([]byte(exp))
	testutils.Err(t, "e.Encrypt", err)

	got, err := e.Decrypt(out)
	testutils.Err(t, "e.Decrypt", err)

	testutils.AssertEqual(t, "result", exp, string(got))
}
