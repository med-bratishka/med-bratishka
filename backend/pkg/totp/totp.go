package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDigits = 6
	defaultPeriod = 30
)

func GenerateSecret() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return enc.EncodeToString(raw), nil
}

func BuildURL(issuer, account, secret string) string {
	label := url.PathEscape(issuer + ":" + account)
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", issuer)
	values.Set("algorithm", "SHA1")
	values.Set("digits", strconv.Itoa(defaultDigits))
	values.Set("period", strconv.Itoa(defaultPeriod))
	return "otpauth://totp/" + label + "?" + values.Encode()
}

func Validate(secret, code string, now time.Time, allowedSkew int) bool {
	code = strings.TrimSpace(code)
	if len(code) != defaultDigits {
		return false
	}
	counter := now.Unix() / defaultPeriod
	for offset := -allowedSkew; offset <= allowedSkew; offset++ {
		if hotp(secret, uint64(counter+int64(offset))) == code {
			return true
		}
	}
	return false
}

func hotp(secret string, counter uint64) string {
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	key, err := enc.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return ""
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	binCode := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)
	otp := binCode % int(math.Pow10(defaultDigits))
	return fmt.Sprintf("%0*d", defaultDigits, otp)
}
