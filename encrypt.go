package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

func AesCBCEncrypt(rawData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	padding := blockSize - len(rawData)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	rawData = append(rawData, padtext...)
	cipherText := make([]byte, blockSize+len(rawData))
	iv := cipherText[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[blockSize:], rawData)

	return cipherText, nil
}

func AesCBCDncrypt(encryptData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	if len(encryptData) < blockSize {
		panic("ciphertext too short")
	}
	iv := encryptData[:blockSize]
	encryptData = encryptData[blockSize:]
	if len(encryptData)%blockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptData, encryptData)
	length := len(encryptData)
	unpadding := int(encryptData[length-1])
	encryptData = encryptData[:(length - unpadding)]
	return encryptData, nil
}

type Encryptor struct {
	block cipher.Block
}

func (e *Encryptor) aesCBCEncrypt(rawData []byte) ([]byte, error) {
	blockSize := e.block.BlockSize()
	padding := blockSize - len(rawData)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	rawData = append(rawData, padtext...)
	cipherText := make([]byte, blockSize+len(rawData))
	iv := cipherText[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	mode := cipher.NewCBCEncrypter(e.block, iv)
	mode.CryptBlocks(cipherText[blockSize:], rawData)

	return cipherText, nil
}
func (e *Encryptor) aesCBCDncrypt(encryptData []byte) ([]byte, error) {
	blockSize := e.block.BlockSize()
	if len(encryptData) < blockSize {
		panic("ciphertext too short")
	}
	iv := encryptData[:blockSize]
	encryptData = encryptData[blockSize:]
	if len(encryptData)%blockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(e.block, iv)
	mode.CryptBlocks(encryptData, encryptData)
	length := len(encryptData)
	unpadding := int(encryptData[length-1])
	encryptData = encryptData[:(length - unpadding)]
	return encryptData, nil
}
func (e *Encryptor) Encrypt(rawDatatext string) (string, error) {
	rawData := []byte(rawDatatext)
	data, err := e.aesCBCEncrypt(rawData)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (e *Encryptor) Dncrypt(rawData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(rawData)
	if err != nil {
		return "", err
	}
	dnData, err := e.aesCBCDncrypt(data)
	if err != nil {
		return "", err
	}
	return string(dnData), nil
}
func (e *Encryptor) SetSecret(key string) (err error) {
	keybytes := []byte(key)
	l := len(keybytes)
	if l < 1 {
		return errors.New("secret should not be null")
	}
	if l > 32 {
		return fmt.Errorf("secret should not be larger then 32,but %d", l)
	}
	var blockkey []byte
	if l <= 16 {
		blockkey = make([]byte, 16)
	} else if l <= 24 {
		blockkey = make([]byte, 24)
	} else if l <= 32 {
		blockkey = make([]byte, 32)
	}
	copy(blockkey, keybytes)
	e.block, err = aes.NewCipher(blockkey)
	return err
}
func NewEncryptor(secret string) *Encryptor {
	r := &Encryptor{}
	if err := r.SetSecret(secret); err != nil {
		panic(err)
	}
	return r
}
