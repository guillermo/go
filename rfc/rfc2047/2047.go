// Convert rfc2047 email headers into its utf-8 version
// Not all the rfc is implemented but should work in 99% of the cases.
package rfc2047

import (
	"bytes"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

const regString = "=\\?([^?]+)\\?(B|Q|b|q)\\?([^?]+)\\?="
const qregString = "=([0-9A-F]{2})"

var reg *regexp.Regexp
var qreg *regexp.Regexp

func init() {
	reg = regexp.MustCompile(regString)
	qreg = regexp.MustCompile(qregString)
}

// Convert will return an utf-8 encoded string from a rfc2047 encoded string
func Convert(in string) (out string) {
	s := &Rfc2047{S: in}
	return s.String()
}

type Rfc2047 struct {
	S      string  // Input String
	Errors []error // Hold all the errors produced while calling String()
}

// String return the utf-8 encoded version of the string.
// Any error found on the way will be stored in the Errors slice.
func (s *Rfc2047) String() string {
	return reg.ReplaceAllStringFunc(string(s.S), s.transform)
}

func (s *Rfc2047) transform(input string) string {
	subMatchs := reg.FindStringSubmatch(input)
	if len(subMatchs) != 4 {
		s.appendError(fmt.Errorf("Unexpected input: %s", input))
	}
	charset := subMatchs[1]
	encoding := subMatchs[2]
	text := subMatchs[3]

	// Base64 encoding
	if encoding == "b" || encoding == "B" {
		return s.bdec(charset, encoding, text)
	}
	return s.qdec(charset, encoding, text)
}

func (s *Rfc2047) bdec(charset, encoding, text string) string {
	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		s.appendError(err)
	}
	return string(s.convertEncoding(charset, data))
}

func (s *Rfc2047) qdec(char, encoding, text string) string {
	text = strings.Replace(text, "_", " ", -1)
	decoded := qreg.ReplaceAllFunc([]byte(text), func(input []byte) []byte {
		input = input[1:len(input)]
		out := s.convertEncoding(char, s.dehex(input))
		return out
	})
	return string(decoded)
}

func (s *Rfc2047) dehex(in []byte) (out []byte) {
	out = make([]byte, hex.DecodedLen(len(in)))
	_, err := hex.Decode(out, in)
	if err != nil {
		s.appendError(err)
	}
	return
}

func (s *Rfc2047) convertEncoding(char string, input []byte) []byte {
	reader, err := charset.NewReader(char, bytes.NewReader(input))
	if err != nil {
		s.appendError(fmt.Errorf("Transcoding: %v : %s", input, err))
		return input
	}

	data, err := ioutil.ReadAll(reader)
	s.appendError(err)
	return data
}

func (s *Rfc2047) appendError(err error) {
	if err != nil {
		s.Errors = append(s.Errors, err)
	}
}
