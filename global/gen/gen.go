package gen

import (
	"crypto/rand"
	"fmt"
	"time"
)

func RandIntStr(leng int) (string, error) {
	codes := make([]byte, leng)
	if _, err := rand.Read(codes); err != nil {
		return "", err
	}

	for i := 0; i < leng; i++ {
		codes[i] = uint8(48 + (codes[i] % 10))
	}

	return string(codes), nil
}

func NanoSecGen() (error, string) {
	var err error = nil
	var retStr string
	retStr = fmt.Sprint(time.Now().Nanosecond())[:7]
	return err, retStr
}
