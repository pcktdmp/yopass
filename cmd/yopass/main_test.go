package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestCLI(t *testing.T) {
	if !pingDemoServer() {
		t.Skip("skipping CLI integration tests - could not ping demo server")
	}

	msg := "yopass CLI integration test message"
	stdin, err := tempFile(msg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(stdin.Name())
	defer stdin.Close()

	out := bytes.Buffer{}
	err = encrypt(stdin, &out)
	if err != nil {
		t.Fatalf("expected no encryption error, got %q", err)
	}
	if !strings.HasPrefix(out.String(), viper.GetString("url")) {
		t.Fatalf("expected encrypt to return secret URL, got %q", out.String())
	}

	viper.Set("decrypt", out.String())
	out.Reset()
	err = decrypt(&out)
	if err != nil {
		t.Fatalf("expected no decryption error, got %q", err)
	}
	if out.String() != msg {
		t.Fatalf("expected secret to match original %q, got %q", msg, out.String())
	}
}

func TestCLIFileUpload(t *testing.T) {
	if !pingDemoServer() {
		t.Skip("skipping CLI integration tests - could not ping demo server")
	}

	msg := "yopass CLI integration test file upload"
	file, err := tempFile(msg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	viper.Set("file", file.Name())

	out := bytes.Buffer{}
	err = encrypt(nil, &out)
	if err != nil {
		t.Fatalf("expected no encryption error, got %q", err)
	}
	if !strings.HasPrefix(out.String(), viper.GetString("url")) {
		t.Fatalf("expected encrypt to return secret URL, got %q", out.String())
	}

	viper.Set("decrypt", out.String())
	out.Reset()
	err = decrypt(&out)
	if err != nil {
		t.Fatalf("expected no decryption error, got %q", err)
	}
	// Note yopass decrypt currently always prints the content to stdout. This
	// could be changed to create a file, but will need to handle the case that
	// the file already exists.
	if out.String() != msg {
		t.Fatalf("expected secret to match original %q, got %q", msg, out.String())
	}
}

func TestExpiration(t *testing.T) {
	tests := []struct {
		input  string
		output int32
	}{
		{
			"1h",
			3600,
		},
		{
			"1d",
			86400,
		},
		{
			"1w",
			604800,
		},
		{
			"invalid",
			0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := expiration(tc.input)
			if got != tc.output {
				t.Fatalf("Expected %d; got %d", tc.output, got)
			}
		})
	}
}
func TestCLIParse(t *testing.T) {
	tests := []struct {
		args   []string
		exit   int
		output string
	}{
		{
			args:   []string{},
			exit:   -1,
			output: "",
		},
		{
			args:   []string{"--one-time=false"},
			exit:   -1,
			output: "",
		},
		{
			args:   []string{"-h"},
			exit:   0,
			output: "Yopass - Secure sharing for secrets, passwords and files",
		},
		{
			args:   []string{"--help"},
			exit:   0,
			output: "Yopass - Secure sharing for secrets, passwords and files",
		},
		{
			args:   []string{"--decrypt"},
			exit:   1,
			output: "flag needs an argument: --decrypt",
		},
		{
			args:   []string{"--unknown"},
			exit:   1,
			output: "unknown flag: --unknown",
		},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.args, "_"), func(t *testing.T) {
			stderr := bytes.Buffer{}
			exit := parse(test.args, &stderr)

			if test.exit != exit {
				t.Errorf("expected parse to exit with %d, got %d", test.exit, exit)
			}
			if test.output != stderr.String() && (test.output != "" && !strings.HasPrefix(stderr.String(), test.output)) {
				t.Errorf("expected parse to print %q, got: %q", test.output, stderr.String())
			}
		})
	}
}

func pingDemoServer() bool {
	resp, err := http.Get(viper.GetString("url"))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func tempFile(s string) (*os.File, error) {
	f, err := ioutil.TempFile("", "yopass-")
	if err != nil {
		return nil, err
	}
	if _, err := f.Write([]byte(s)); err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	return f, nil
}
