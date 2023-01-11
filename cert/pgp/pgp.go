package pgp

import (
	"bytes"
	"crypto"
	"io"
	"os"

	"github.com/deweppro/go-errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/openpgp/packet"
)

type (
	Config struct {
		Name, Email, Comment string
	}

	Cert struct {
		Public  []byte
		Private []byte
	}
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	store struct {
		key     *openpgp.Entity
		conf    *packet.Config
		headers map[string]string
	}

	Signer interface {
		SetKey(b []byte, passwd string) error
		SetKeyFromFile(filename string, passwd string) error
		SetHash(hash crypto.Hash, bits int)
		PublicKey() ([]byte, error)
		PublicKeyBase64() ([]byte, error)
		Sign(in io.Reader, out io.Writer) error
	}
)

func New() Signer {
	return &store{
		conf: &packet.Config{
			DefaultHash: crypto.SHA512,
			RSABits:     4096,
		},
		headers: make(map[string]string),
	}
}

func (v *store) SetKey(b []byte, passwd string) error {
	r := bytes.NewReader(b)
	return v.readKey(r, passwd)
}

func (v *store) SetHash(hash crypto.Hash, bits int) {
	v.conf.DefaultHash = hash
	v.conf.RSABits = bits
}

func (v *store) SetHeaders(headers ...string) error {
	h, err := createHeaders(headers)
	if err != nil {
		return err
	}
	v.headers = mergeHeaders(v.headers, h)
	return nil
}

func (v *store) SetKeyFromFile(filename string, passwd string) error {
	r, err := os.Open(filename)
	if err != nil {
		return errors.WrapMessage(err, "read key from file")
	}
	return v.readKey(r, passwd)
}

func (v *store) PublicKey() ([]byte, error) {
	if v.key == nil {
		return nil, errors.New("key is empty")
	}

	var buf bytes.Buffer
	if err := v.key.Serialize(&buf); err != nil {
		return nil, errors.WrapMessage(err, "serialize public key")
	}
	return buf.Bytes(), nil
}

func (v *store) PublicKeyBase64() ([]byte, error) {
	if v.key == nil {
		return nil, errors.New("key is empty")
	}

	var buf bytes.Buffer
	enc, err := armor.Encode(&buf, openpgp.PublicKeyType, v.headers)
	if err != nil {
		return nil, errors.WrapMessage(err, "init armor encoder")
	}
	if err = v.key.Serialize(enc); err != nil {
		return nil, errors.WrapMessage(err, "serialize public key")
	}
	if err = enc.Close(); err != nil {
		return nil, errors.WrapMessage(err, "close armor encoder")
	}
	return buf.Bytes(), nil
}

func (v *store) readKey(r io.ReadSeeker, passwd string) error {
	block, err := armor.Decode(r)
	if err != nil {
		return errors.WrapMessage(err, "armor decode key")
	}
	if block.Type != openpgp.PrivateKeyType {
		return errors.WrapMessage(err, "invalid key type")
	}
	if _, err = r.Seek(0, 0); err != nil {
		return errors.WrapMessage(err, "seek key file")
	}
	keys, err := openpgp.ReadArmoredKeyRing(r)
	if err != nil {
		return errors.WrapMessage(err, "read armored key")
	}
	v.key = keys[0]
	if v.key.PrivateKey.Encrypted {
		if err = v.key.PrivateKey.Decrypt([]byte(passwd)); err != nil {
			return errors.WrapMessage(err, "invalid password")
		}
	}
	v.headers = mergeHeaders(v.headers, block.Header)
	return nil
}

func (v *store) Sign(in io.Reader, out io.Writer) error {
	if v.key == nil {
		return errors.New("key is empty")
	}

	w, err := clearsign.Encode(out, v.key.PrivateKey, v.conf)
	if err != nil {
		return errors.WrapMessage(err, "init")
	}
	if _, err = io.Copy(w, in); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func generatePrivateKey(key *openpgp.Entity, w io.Writer, headers map[string]string) error {
	enc, err := armor.Encode(w, openpgp.PrivateKeyType, headers)
	if err != nil {
		return errors.WrapMessage(err, "init armor encoder")
	}
	defer enc.Close() //nolint: errcheck

	if err = key.SerializePrivate(enc, nil); err != nil {
		return errors.WrapMessage(err, "serialize private key")
	}

	return nil
}

func generatePublicKey(key *openpgp.Entity, w io.Writer, headers map[string]string) error {
	enc, err := armor.Encode(w, openpgp.PublicKeyType, headers)
	if err != nil {
		return errors.WrapMessage(err, "create OpenPGP armor")
	}
	defer enc.Close() //nolint: errcheck

	if err = key.Serialize(enc); err != nil {
		return errors.WrapMessage(err, "serialize public key")
	}

	return nil
}

func setupIdentities(key *openpgp.Entity, c *packet.Config) error {
	// Sign all the identities
	for _, id := range key.Identities {
		id.SelfSignature.PreferredCompression = []uint8{1, 2, 3, 0}
		id.SelfSignature.PreferredHash = []uint8{2, 8, 10, 1, 3, 9, 11}
		id.SelfSignature.PreferredSymmetric = []uint8{9, 8, 7, 3, 2}

		if err := id.SelfSignature.SignUserId(id.UserId.Id, key.PrimaryKey, key.PrivateKey, c); err != nil {
			return err
		}
	}
	return nil
}

func createHeaders(v []string) (map[string]string, error) {
	if len(v)%2 != 0 {
		return nil, errors.New("odd headers count")
	}
	result := make(map[string]string, len(v)/2)
	for i := 0; i < len(v); i += 2 {
		result[v[i]] = v[i+1]
	}
	return result, nil
}

func mergeHeaders(h ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range h {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func NewCert(c Config, hash crypto.Hash, bits int, headers ...string) (*Cert, error) {
	h, err := createHeaders(headers)
	if err != nil {
		return nil, errors.WrapMessage(err, "parse headers")
	}

	conf := &packet.Config{
		DefaultHash: hash,
		RSABits:     bits,
	}

	key, err := openpgp.NewEntity(c.Name, c.Comment, c.Email, conf)
	if err != nil {
		return nil, errors.WrapMessage(err, "generate entity")
	}

	if err = setupIdentities(key, conf); err != nil {
		return nil, errors.WrapMessage(err, "setup entity")
	}

	var priv bytes.Buffer
	if err = generatePrivateKey(key, &priv, h); err != nil {
		return nil, errors.WrapMessage(err, "generate private key")
	}

	var pub bytes.Buffer
	if err = generatePublicKey(key, &pub, h); err != nil {
		return nil, errors.WrapMessage(err, "generate public key")
	}

	return &Cert{
		Public:  pub.Bytes(),
		Private: priv.Bytes(),
	}, nil
}

func NewCertSHA512(c Config, headers ...string) (*Cert, error) {
	return NewCert(c, crypto.SHA512, 4096, headers...)
}
