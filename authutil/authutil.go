package authutil

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func MakeSMSToken() (string, []byte, error) {
	bytes := make([]byte, 3)
	_, err := rand.Read(bytes)
	token := hex.EncodeToString(bytes)
	return token, bytes, err
}

func MakeAuthToken() (string, []byte, error) {
	bytes := make([]byte, 24)
	_, err := rand.Read(bytes)
	token := hex.EncodeToString(bytes)
	return token, bytes, err
}

func HashApiKey(apiKey string, salt []byte) (string, []byte) {
	var passwordBytes = []byte(apiKey)
	var sha512Hasher = sha512.New()
	passwordBytes = append(passwordBytes, salt...)
	sha512Hasher.Write(passwordBytes)
	var hashedPasswordBytes = sha512Hasher.Sum(nil)
	var hashedPasswordHex = hex.EncodeToString(hashedPasswordBytes)
	return hashedPasswordHex, hashedPasswordBytes
}

func HashTokenBytes(bytesR []byte) string {
	var sha512Hasher = sha512.New()
	sha512Hasher.Write(bytesR)
	var hashedTokenBytes = sha512Hasher.Sum(nil)
	return hex.EncodeToString(hashedTokenBytes)
}

func Sha1Hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bytesR := h.Sum(nil)
	return hex.EncodeToString(bytesR)
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

// IP ADDR TO STORAGE
func InetAton(ip string) (ipInt uint32) {
	ipByte := net.ParseIP(ip).To4()
	for i := 0; i < len(ipByte); i++ {
		ipInt |= uint32(ipByte[i])
		if i < 3 {
			ipInt <<= 8
		}
	}
	return
}

// IPADDR FROM STORAGE
func InetNtoa(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8),
		byte(ip))
}

func GeneratePW(rawPWStr string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPWStr), 14)
	hashStr := string(bytes)
	return hashStr, err
}
