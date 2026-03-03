package vault

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const serviceName = "supashift"

type Vault interface {
	Backend() string
	SetToken(profile, token string) error
	GetToken(profile string) (string, error)
	DeleteToken(profile string) error
	ListTokens() (map[string]string, error)
}

type Manager struct {
	selected Vault
	keyring  *KeyringVault
	file     *FileVault
}

func NewManager(configDir, backend string) (*Manager, error) {
	kv := &KeyringVault{}
	fv := NewFileVault(filepath.Join(configDir, "vault.age"))

	if backend == "keyring" {
		if !kv.Available() {
			return nil, errors.New("keyring no disponible")
		}
		return &Manager{selected: kv, keyring: kv, file: fv}, nil
	}
	if backend == "file" {
		return &Manager{selected: fv, keyring: kv, file: fv}, nil
	}
	if kv.Available() {
		return &Manager{selected: kv, keyring: kv, file: fv}, nil
	}
	return &Manager{selected: fv, keyring: kv, file: fv}, nil
}

func (m *Manager) Vault() Vault { return m.selected }

func (m *Manager) Doctor() (bool, string) {
	if m.keyring.Available() {
		return true, "keyring del sistema disponible (recomendado)"
	}
	return false, "keyring no disponible; usar fallback cifrado con age + passphrase"
}

type KeyringVault struct{}

func (k *KeyringVault) Backend() string { return "keyring" }

func (k *KeyringVault) key(profile string) string {
	return "profile:" + profile
}

func (k *KeyringVault) Available() bool {
	probeKey := "probe"
	probeVal := "ok"
	if err := keyring.Set(serviceName, probeKey, probeVal); err != nil {
		return false
	}
	_, _ = keyring.Get(serviceName, probeKey)
	_ = keyring.Delete(serviceName, probeKey)
	return true
}

func (k *KeyringVault) SetToken(profile, token string) error {
	return keyring.Set(serviceName, k.key(profile), token)
}

func (k *KeyringVault) GetToken(profile string) (string, error) {
	val, err := keyring.Get(serviceName, k.key(profile))
	if err != nil {
		return "", fmt.Errorf("token no encontrado en keyring para %s", profile)
	}
	return val, nil
}

func (k *KeyringVault) DeleteToken(profile string) error {
	err := keyring.Delete(serviceName, k.key(profile))
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}
	return nil
}

func (k *KeyringVault) ListTokens() (map[string]string, error) {
	return nil, errors.New("listar tokens no soportado en backend keyring")
}

type FileVault struct {
	path       string
	passphrase string
}

func NewFileVault(path string) *FileVault {
	return &FileVault{path: path}
}

func (f *FileVault) Backend() string { return "file-age" }

func (f *FileVault) SetToken(profile, token string) error {
	items, err := f.load()
	if err != nil {
		return err
	}
	items[profile] = token
	return f.save(items)
}

func (f *FileVault) GetToken(profile string) (string, error) {
	items, err := f.load()
	if err != nil {
		return "", err
	}
	v, ok := items[profile]
	if !ok {
		return "", fmt.Errorf("token no encontrado para %s", profile)
	}
	return v, nil
}

func (f *FileVault) DeleteToken(profile string) error {
	items, err := f.load()
	if err != nil {
		return err
	}
	delete(items, profile)
	return f.save(items)
}

func (f *FileVault) ListTokens() (map[string]string, error) {
	return f.load()
}

func (f *FileVault) load() (map[string]string, error) {
	if _, err := os.Stat(f.path); errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	b, err := os.ReadFile(f.path)
	if err != nil {
		return nil, err
	}
	pass, err := f.getPassphrase()
	if err != nil {
		return nil, err
	}
	id, err := age.NewScryptIdentity(pass)
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(bytes.NewReader(b), id)
	if err != nil {
		return nil, errors.New("no se pudo descifrar vault (passphrase inválida o archivo corrupto)")
	}
	plain, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	items := map[string]string{}
	if len(plain) == 0 {
		return items, nil
	}
	if err := json.Unmarshal(plain, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (f *FileVault) save(items map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0o700); err != nil {
		return err
	}
	plain, err := json.Marshal(items)
	if err != nil {
		return err
	}
	pass, err := f.getPassphrase()
	if err != nil {
		return err
	}
	r, err := age.NewScryptRecipient(pass)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	w, err := age.Encrypt(buf, r)
	if err != nil {
		return err
	}
	if _, err := w.Write(plain); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := os.WriteFile(f.path, buf.Bytes(), 0o600); err != nil {
		return err
	}
	return nil
}

func (f *FileVault) getPassphrase() (string, error) {
	if f.passphrase != "" {
		return f.passphrase, nil
	}
	if env := os.Getenv("SUPASHIFT_PASSPHRASE"); strings.TrimSpace(env) != "" {
		f.passphrase = env
		return f.passphrase, nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("SUPASHIFT_PASSPHRASE requerido en modo no interactivo")
	}
	fmt.Fprint(os.Stderr, "Passphrase del vault: ")
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	f.passphrase = strings.TrimSpace(string(b))
	if f.passphrase == "" {
		return "", errors.New("passphrase vacía")
	}
	return f.passphrase, nil
}

func ConfirmReveal() bool {
	fmt.Fprint(os.Stderr, "Confirmar reveal del token? (type YES): ")
	in := bufio.NewReader(os.Stdin)
	v, _ := in.ReadString('\n')
	return strings.TrimSpace(v) == "YES"
}
