package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
)

type FileKey struct {
	exponent *big.Int
	modulus  *big.Int
}

func readFile(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("Cannot open or read file: %s", name)
	}
	return data, nil
}

func requireNewFile(name string) error {
	if name == "" {
		return errors.New("No filename provided")
	}
	_, err := os.Stat(name)
	if err == nil {
		return fmt.Errorf("File already exists. Choose another name: %s", name)
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeFile(name string, data []byte) error {
	if err := requireNewFile(name); err != nil {
		return err
	}
	if err := os.WriteFile(name, data, 0644); err != nil {
		return fmt.Errorf("Cannot create or write file: %s", name)
	}
	return nil
}

func validateKey(key FileKey) error {
	if key.exponent == nil || key.modulus == nil {
		return errors.New("Invalid key")
	}
	if key.exponent.Cmp(big.NewInt(1)) <= 0 || key.modulus.Cmp(big.NewInt(255)) <= 0 {
		return errors.New("Invalid key: exponent must exceed 1 and modulus must exceed 255")
	}
	return nil
}

func encryptFile(source, destination string, key FileKey) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := requireNewFile(destination); err != nil {
		return err
	}
	original, err := readFile(source)
	if err != nil {
		return err
	}

	outFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Cannot create file: %s", destination)
	}
	writer := bufio.NewWriter(outFile)

	if _, err := fmt.Fprintf(writer, "RSA_FILE_V1\n%s\n%d\n", key.modulus.String(), len(original)); err != nil {
		outFile.Close()
		return errors.New("Cannot write encrypted file")
	}

	cache := make(map[byte]string)
	for _, b := range original {
		enc, exists := cache[b]
		if !exists {
			block := new(big.Int).Exp(big.NewInt(int64(b)), key.exponent, key.modulus)
			enc = block.String()
			cache[b] = enc
		}
		if _, err := writer.WriteString(enc + "\n"); err != nil {
			outFile.Close()
			return errors.New("Cannot write encrypted file")
		}
	}

	if err := writer.Flush(); err != nil {
		outFile.Close()
		return errors.New("Cannot write encrypted file")
	}
	if err := outFile.Close(); err != nil {
		return errors.New("Cannot write encrypted file")
	}
	return nil
}

func decryptFile(source, destination string, key FileKey) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := requireNewFile(destination); err != nil {
		return err
	}

	inFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Cannot open file: %s", source)
	}
	defer inFile.Close()

	info, err := inFile.Stat()
	if err != nil {
		return errors.New("Cannot inspect encrypted file")
	}

	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanWords)

	nextToken := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	magic, ok1 := nextToken()
	modulusStr, ok2 := nextToken()
	countText, ok3 := nextToken()
	if !ok1 || !ok2 || !ok3 || magic != "RSA_FILE_V1" {
		return errors.New("Unknown encrypted file format")
	}
	if modulusStr != key.modulus.String() {
		return errors.New("The file was encrypted with a different RSA modulus")
	}
	if !isDigitsOnly(countText) {
		return errors.New("Invalid block count")
	}
	count, err := strconv.ParseUint(countText, 10, 64)
	if err != nil {
		return errors.New("Invalid block count")
	}
	if count > uint64(info.Size()) {
		return errors.New("Invalid block count: count exceeds the encrypted file size")
	}

	limit255 := big.NewInt(255)
	restored := make([]byte, 0)
	cache := make(map[string]byte)

	for i := uint64(0); i < count; i++ {
		token, ok := nextToken()
		if !ok {
			if err := scanner.Err(); err != nil {
				return errors.New("Cannot read encrypted file")
			}
			return errors.New("The encrypted file is truncated")
		}
		if len(token) > len(key.modulus.String()) {
			return errors.New("The encrypted block is too large")
		}
		block, ok := new(big.Int).SetString(token, 10)
		if !ok || block.Sign() < 0 || block.String() != token {
			return errors.New("Invalid encrypted block")
		}
		if block.Cmp(key.modulus) >= 0 {
			return errors.New("Invalid encrypted block")
		}

		decoded, exists := cache[token]
		if !exists {
			byteVal := new(big.Int).Exp(block, key.exponent, key.modulus)
			if byteVal.Cmp(limit255) > 0 {
				return errors.New("Invalid decoded byte: check the key and encrypted file")
			}
			v, err := strconv.Atoi(byteVal.String())
			if err != nil {
				return errors.New("Invalid decoded byte: check the key and encrypted file")
			}
			decoded = byte(v)
			cache[token] = decoded
		}
		restored = append(restored, decoded)
	}

	if _, extra := nextToken(); extra {
		return errors.New("Unexpected extra blocks in the encrypted file")
	}
	if err := scanner.Err(); err != nil {
		return errors.New("Cannot read encrypted file")
	}

	return writeFile(destination, restored)
}

func compareFiles(original, restored string) (bool, error) {
	origData, err := readFile(original)
	if err != nil {
		return false, err
	}
	restData, err := readFile(restored)
	if err != nil {
		return false, err
	}
	eq := bytes.Equal(origData, restData)
	if eq {
		fmt.Println("Verification: the files match byte for byte.")
	} else {
		fmt.Println("Verification: the files do NOT match!")
	}
	return eq, nil
}
