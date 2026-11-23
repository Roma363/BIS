package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const (
	primeDigits = 25
	testRounds  = 10
)

var one = big.NewInt(1)
var two = big.NewInt(2)

// isProbablePrimeFermat перевіряє кандидата за малою теоремою Ферма.
func isProbablePrimeFermat(n *big.Int, rounds int) bool {
	if n.Cmp(two) < 0 {
		return false
	}
	if n.Cmp(two) == 0 || n.Cmp(big.NewInt(3)) == 0 {
		return true
	}
	if new(big.Int).Mod(n, two).Sign() == 0 {
		return false
	}

	max := new(big.Int).Sub(n, two)
	for i := 0; i < rounds; i++ {
		a, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(fmt.Errorf("генерація основи тесту Ферма: %w", err))
		}
		a.Add(a, two)
		power := new(big.Int).Exp(a, new(big.Int).Sub(n, one), n)
		if power.Cmp(one) != 0 {
			return false
		}
	}
	return true
}

// isProbablePrimeRabinMiller виконує додаткову перевірку Рабіна—Міллера.
func isProbablePrimeRabinMiller(n *big.Int, rounds int) bool {
	if n.Cmp(two) < 0 {
		return false
	}
	if n.Cmp(two) == 0 || n.Cmp(big.NewInt(3)) == 0 {
		return true
	}
	if new(big.Int).Mod(n, two).Sign() == 0 {
		return false
	}

	d := new(big.Int).Sub(n, one)
	s := 0
	for new(big.Int).Mod(d, two).Sign() == 0 {
		d.Div(d, two)
		s++
	}

	max := new(big.Int).Sub(n, big.NewInt(3))
	nMinusOne := new(big.Int).Sub(n, one)
	for i := 0; i < rounds; i++ {
		a, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(fmt.Errorf("генерація основи тесту Рабіна—Міллера: %w", err))
		}
		a.Add(a, two)
		x := new(big.Int).Exp(a, d, n)
		if x.Cmp(one) == 0 || x.Cmp(nMinusOne) == 0 {
			continue
		}

		probablyPrime := false
		for j := 1; j < s; j++ {
			x.Mul(x, x).Mod(x, n)
			if x.Cmp(nMinusOne) == 0 {
				probablyPrime = true
				break
			}
		}
		if !probablyPrime {
			return false
		}
	}
	return true
}

// generateLargePrime генерує просте число із заданою кількістю цифр.
func generateLargePrime(numDigits, rounds int) (*big.Int, error) {
	if numDigits < 2 {
		return nil, errors.New("кількість цифр простого числа має бути не меншою за 2")
	}

	min := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(numDigits-1)), nil)
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(numDigits)), nil)
	rangeSize := new(big.Int).Sub(max, min)

	for {
		candidate, err := rand.Int(rand.Reader, rangeSize)
		if err != nil {
			return nil, fmt.Errorf("генерація кандидата: %w", err)
		}
		candidate.Add(candidate, min)
		candidate.SetBit(candidate, 0, 1)

		if isProbablePrimeFermat(candidate, rounds) &&
			isProbablePrimeRabinMiller(candidate, rounds) {
			return candidate, nil
		}
	}
}

// modInverse повертає обернений елемент e за модулем phi.
func modInverse(e, phi *big.Int) (*big.Int, error) {
	oldR, r := new(big.Int).Set(e), new(big.Int).Set(phi)
	oldS, s := big.NewInt(1), big.NewInt(0)

	for r.Sign() != 0 {
		quotient := new(big.Int).Div(oldR, r)
		oldR, r = r, new(big.Int).Sub(oldR, new(big.Int).Mul(quotient, r))
		oldS, s = s, new(big.Int).Sub(oldS, new(big.Int).Mul(quotient, s))
	}

	if oldR.Cmp(one) != 0 {
		return nil, errors.New("обернений елемент не існує")
	}
	return new(big.Int).Mod(oldS, phi), nil
}

type rsaKey struct {
	exponent *big.Int
	modulus  *big.Int
}

// generateRSAKeys створює відкритий і закритий ключі RSA.
func generateRSAKeys(numDigits, rounds int) (*big.Int, *big.Int, rsaKey, rsaKey, error) {
	p, err := generateLargePrime(numDigits, rounds)
	if err != nil {
		return nil, nil, rsaKey{}, rsaKey{}, err
	}

	var q *big.Int
	for {
		q, err = generateLargePrime(numDigits, rounds)
		if err != nil {
			return nil, nil, rsaKey{}, rsaKey{}, err
		}
		if q.Cmp(p) != 0 {
			break
		}
	}

	n := new(big.Int).Mul(p, q)
	phi := new(big.Int).Mul(
		new(big.Int).Sub(p, one),
		new(big.Int).Sub(q, one),
	)

	e := big.NewInt(65537)
	if new(big.Int).GCD(nil, nil, e, phi).Cmp(one) != 0 {
		e = big.NewInt(3)
	}
	d, err := modInverse(e, phi)
	if err != nil {
		return nil, nil, rsaKey{}, rsaKey{}, err
	}

	return p, q, rsaKey{exponent: e, modulus: n}, rsaKey{exponent: d, modulus: n}, nil
}

func (key rsaKey) apply(message *big.Int) *big.Int {
	return new(big.Int).Exp(message, key.exponent, key.modulus)
}

func main() {
	p, q, publicKey, privateKey, err := generateRSAKeys(primeDigits, testRounds)
	if err != nil {
		panic(err)
	}

	message := big.NewInt(123456789)
	ciphertext := publicKey.apply(message)
	decrypted := privateKey.apply(ciphertext)

	fmt.Println("=== Генератор великих простих чисел та RSA ===")
	fmt.Printf("Розмір простих чисел: %d десяткових цифр\n", primeDigits)
	fmt.Printf("Раундів перевірки: %d\n\n", testRounds)
	fmt.Printf("Згенероване p: %s\n", p)
	fmt.Printf("Згенероване q: %s\n\n", q)
	fmt.Printf("Відкритий ключ (e, n): (%s, %s)\n", publicKey.exponent, publicKey.modulus)
	fmt.Printf("Закритий ключ (d, n): (%s, %s)\n", privateKey.exponent, privateKey.modulus)
	fmt.Printf("Повідомлення: %s\n", message)
	fmt.Printf("Зашифровано: %s\n", ciphertext)
	fmt.Printf("Розшифровано: %s\n", decrypted)

	if decrypted.Cmp(message) != 0 {
		panic("помилка перевірки RSA")
	}
	fmt.Println("Перевірка RSA: OK")
}
