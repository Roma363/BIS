package main

import (
	"fmt"
	"math/big"
	"os"
)

func mustParse(s string) *BigInt {
	x, err := NewFromString(s)
	if err != nil {
		panic(err)
	}
	return x
}

func verify(name string, got *BigInt, expected *big.Int) {
	gotBig, _ := new(big.Int).SetString(got.String(), 10)
	status := "OK"
	if gotBig.Cmp(expected) != 0 {
		status = "ПОМИЛКА"
		fmt.Printf("%-40s %s\n", name, status)
		fmt.Printf("  отримано:   %s\n", got)
		fmt.Printf("  очікувалось: %s\n", expected.String())
		return
	}
	fmt.Printf("%-40s %s\n", name, status)
}

func main() {
	// Ті самі тестові числа, що в Python/C прикладах
	aStr := "31415926535897932384626433832795028841971693993751"
	bStr := "27182818284590452353602874713526624977572470936999"
	mStr := "10000000000000000000000000000000000000000000000051"
	expStr := "123456789" // для powmod можна взяти менше або більше

	a := mustParse(aStr)
	b := mustParse(bStr)
	m := mustParse(mStr)
	exp := mustParse(expStr)

	fmt.Println("A =", a)
	fmt.Println("B =", b)
	fmt.Println("M =", m)
	fmt.Println()

	// === Операції ===
	sum := a.Add(b)
	prod := a.Mul(b)
	sq := a.Square()
	modRes := a.Mod(m)
	powRes := a.Exp(exp, m)

	fmt.Println("A + B     =", sum)
	fmt.Println("A * B     =", prod)
	fmt.Println("A²        =", sq)
	fmt.Println("A mod M   =", modRes)
	fmt.Println("A^E mod M =", powRes)
	fmt.Println()

	// === Перевірка з math/big ===
	fmt.Println("=== Перевірка коректності (звірка з math/big) ===")

	ai, _ := new(big.Int).SetString(aStr, 10)
	bi, _ := new(big.Int).SetString(bStr, 10)
	mi, _ := new(big.Int).SetString(mStr, 10)
	ei, _ := new(big.Int).SetString(expStr, 10)

	verify("Складання (A + B)", sum, new(big.Int).Add(ai, bi))
	verify("Множення (A * B)", prod, new(big.Int).Mul(ai, bi))
	verify("Зведення в квадрат (A²)", sq, new(big.Int).Mul(ai, ai))
	verify("Залишок за модулем (A mod M)", modRes, new(big.Int).Mod(ai, mi))
	verify("Швидке потенціювання (A^E mod M)", powRes, new(big.Int).Exp(ai, ei, mi))

	// Якщо хочеш CLI-режим як у C-версії — можна додати обробку os.Args
	_ = os.Args
}
