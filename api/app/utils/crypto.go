package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"os"
)

var bytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func Encrypt(text *string) error {
	block, err := aes.NewCipher([]byte(os.Getenv("API_SECRET")))
	if err != nil {
		return err
	}
	plainText := []byte(*text)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	*text = Encode(cipherText)
	return nil
}

func Decrypt(text *string) error {
	block, err := aes.NewCipher([]byte(os.Getenv("API_SECRET")))
	if err != nil {
		return err
	}
	cipherText := Decode(*text)
	cfb := cipher.NewCFBDecrypter(block, bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	*text = string(plainText)
	return nil
}
