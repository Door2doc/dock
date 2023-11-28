package password

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func roundTrip(plain string) error {
	p := Password(plain)
	encoded, err := json.Marshal(p)
	if err != nil {
		return err
	}
	var decoded Password
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return err
	}
	if !reflect.DeepEqual(p, decoded) {
		return fmt.Errorf("Invalid roundtrip, got %q, but want %q", decoded.PlainText(), p.PlainText())
	}
	return nil
}

func TestPassword_MarshalJSON(t *testing.T) {
	plaintext := "secret"
	p := Password(plaintext)
	j, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(j), plaintext) {
		t.Errorf("Password not marshalled safely, got %q but don't want to see %q", j, plaintext)
	}
	if err := roundTrip(plaintext); err != nil {
		t.Error(err)
	}
}

func TestPassword_JSON(t *testing.T) {
	f := func(plain string) bool {
		return roundTrip(plain) == nil
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestPassword_String(t *testing.T) {
	f := func(plain string) bool {
		p := Password(plain)
		s := p.String()
		return s == "[redacted]"
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestPassword_PlainText(t *testing.T) {
	f := func(plain string) bool {
		p := Password(plain)
		s := p.PlainText()
		return s == plain
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
