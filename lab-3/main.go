package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
)

func ask(reader *bufio.Reader, prompt, defaultValue string) (string, error) {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil && (err != io.EOF || line == "") {
		return "", errors.New("Cannot read console input")
	}
	answer := strings.TrimRight(line, "\r\n")
	if answer == "" {
		answer = defaultValue
	}
	if answer == "" {
		return "", errors.New("No value provided")
	}
	return answer, nil
}

func isDigitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func run() (int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Laboratory 3. RSA file encryption")

	lengthStr, err := ask(reader, "Decimal digits in each of p and q (10-100, Enter = 40): ", "40")
	if err != nil {
		return 1, err
	}
	if !isDigitsOnly(lengthStr) {
		return 1, errors.New("The digit count must be an integer")
	}
	digits, err := strconv.Atoi(lengthStr)
	if err != nil || digits < 10 || digits > 100 {
		return 1, errors.New("Enter a digit count from 10 to 100")
	}

	fmt.Println("Generating keys using 40 Miller-Rabin rounds...")
	keys, err := generateRsaKeys(digits)
	if err != nil {
		return 1, err
	}

	check := new(big.Int).Mod(new(big.Int).Mul(keys.e, keys.d), keys.phi)
	if check.Cmp(big.NewInt(1)) != 0 {
		return 1, errors.New("Key calculation failed")
	}

	fmt.Printf("\nPublic key (e, n):\ne = %s\nn = %s\n", keys.e.String(), keys.n.String())
	fmt.Printf("\nPrivate key (d, n):\nd = %s\nn = %s\n", keys.d.String(), keys.n.String())
	fmt.Println("Verification: (e*d) mod phi = 1")

	source, err := ask(reader, "\nOriginal text filename: ", "")
	if err != nil {
		return 1, err
	}
	encrypted, err := ask(reader, "Encrypted output filename: ", "")
	if err != nil {
		return 1, err
	}
	restored, err := ask(reader, "Restored output filename: ", "")
	if err != nil {
		return 1, err
	}

	outputs := []string{encrypted, restored}
	for i := 0; i < len(outputs); i++ {
		if err := requireNewFile(outputs[i]); err != nil {
			return 1, err
		}
		for j := 0; j < i; j++ {
			if outputs[i] == outputs[j] {
				return 1, errors.New("Output filenames must be different")
			}
		}
	}

	if _, err := readFile(source); err != nil {
		return 1, err
	}

	publicKey := FileKey{exponent: keys.e, modulus: keys.n}
	privateKey := FileKey{exponent: keys.d, modulus: keys.n}

	if err := encryptFile(source, encrypted, publicKey); err != nil {
		return 1, err
	}
	if err := decryptFile(encrypted, restored, privateKey); err != nil {
		return 1, err
	}

	fmt.Printf("\nEncrypted file: %s\nRestored file: %s\n",
		encrypted, restored)

	ok, err := compareFiles(source, restored)
	if err != nil {
		return 1, err
	}
	if !ok {
		return 1, nil
	}
	return 0, nil
}

func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(code)
}
