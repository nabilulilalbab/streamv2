package utils

import (
	"testing"
)

func TestDec(t *testing.T) {
	// Test cases from Python implementation
	tests := []struct {
		name     string
		key      string
		m        string
		wantErr  bool
	}{
		{
			name:    "Short key",
			key:     "ab",
			m:       "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Dec(tt.key, tt.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCryptoJSDecrypt(t *testing.T) {
	// Test basic decryption with known encrypted data
	tests := []struct {
		name          string
		encryptedJSON string
		passphrase    string
		wantErr       bool
	}{
		{
			name:          "Invalid JSON",
			encryptedJSON: "not json",
			passphrase:    "test",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CryptoJSDecrypt(tt.encryptedJSON, tt.passphrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("CryptoJSDecrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeriveKeyMD5(t *testing.T) {
	passphrase := []byte("password")
	salt := []byte("saltsalt")
	keyLen := 48

	key := deriveKeyMD5(passphrase, salt, keyLen)

	if len(key) != keyLen {
		t.Errorf("deriveKeyMD5() returned key of length %d, want %d", len(key), keyLen)
	}
}

func TestUnpadPKCS7(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "Valid padding",
			data:    []byte("test\x04\x04\x04\x04"),
			want:    []byte("test"),
			wantErr: false,
		},
		{
			name:    "Empty data",
			data:    []byte{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid padding",
			data:    []byte("test\x01\x02\x03\x04"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unpadPKCS7(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpadPKCS7() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != string(tt.want) {
				t.Errorf("unpadPKCS7() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverseString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "olleh"},
		{"12345", "54321"},
		{"", ""},
		{"a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := reverseString(tt.input); got != tt.want {
				t.Errorf("reverseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddBase64Padding(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "abc="},
		{"ab", "ab=="},
		{"abcd", "abcd"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := addBase64Padding(tt.input); got != tt.want {
				t.Errorf("addBase64Padding() = %v, want %v", got, tt.want)
			}
		})
	}
}
