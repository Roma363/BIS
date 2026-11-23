package main

import (
	"crypto/rand"
	"errors"
	"math/big"
)

func generateLargePrime(digits int) (*big.Int, error) {
	if digits < 2 {
		return nil, errors.New("At least two digits are required")
	}

	lower := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits-1)), nil)
	upper := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	rangeLimit := new(big.Int).Sub(upper, lower)

	for {
		candidate, err := rand.Int(rand.Reader, rangeLimit)
		if err != nil {
			return nil, err
		}
		candidate.Add(candidate, lower)
		candidate.SetBit(candidate, 0, 1)
		if candidate.ProbablyPrime(40) {
			return candidate, nil
		}
	}
}

type RsaKeys struct {
	p, q   *big.Int
	n, phi *big.Int
	e, d   *big.Int
}

func generateRsaKeys(digits int) (RsaKeys, error) {
	var keys RsaKeys
	keys.e = big.NewInt(65537)
	one := big.NewInt(1)

	for {
		p, err := generateLargePrime(digits)
		if err != nil {
			return RsaKeys{}, err
		}
		pMinus1 := new(big.Int).Sub(p, one)
		g := new(big.Int).GCD(nil, nil, keys.e, pMinus1)
		if g.Cmp(one) == 0 {
			keys.p = p
			break
		}
	}

	for {
		q, err := generateLargePrime(digits)
		if err != nil {
			return RsaKeys{}, err
		}
		if q.Cmp(keys.p) == 0 {
			continue
		}
		qMinus1 := new(big.Int).Sub(q, one)
		g := new(big.Int).GCD(nil, nil, keys.e, qMinus1)
		if g.Cmp(one) == 0 {
			keys.q = q
			break
		}
	}

	keys.n = new(big.Int).Mul(keys.p, keys.q)
	pMinus1 := new(big.Int).Sub(keys.p, one)
	qMinus1 := new(big.Int).Sub(keys.q, one)
	keys.phi = new(big.Int).Mul(pMinus1, qMinus1)

	keys.d = new(big.Int).ModInverse(keys.e, keys.phi)
	if keys.d == nil {
		return RsaKeys{}, errors.New("No modular inverse: e and phi are not coprime")
	}
	return keys, nil
}
