package main

import (
	"errors"
	"fmt"
	"math/bits"
	"strings"
)

const limbBits = 32

// BigInt — невід’ємне велике ціле число (little-endian limbs).
type BigInt struct {
	limbs []uint32
}

// -------------------- Створення --------------------

// New створює число 0 у канонічному внутрішньому поданні.
func New() *BigInt {
	return &BigInt{limbs: []uint32{0}}
}

// NewFromUint64 створює BigInt із малого невід’ємного числа.
func NewFromUint64(n uint64) *BigInt {
	if n == 0 {
		return New()
	}

	limbs := make([]uint32, 0, 2)
	for n > 0 {
		limbs = append(limbs, uint32(n))
		n >>= limbBits
	}
	return &BigInt{limbs: limbs}
}

// NewFromString парсить десятковий рядок у внутрішнє limb-подання BigInt.
func NewFromString(s string) (*BigInt, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty string")
	}

	res := New()
	for _, c := range s {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("invalid digit: %c", c)
		}
		digit := uint32(c - '0')
		// res = res * 10 + digit
		carry := uint64(digit)
		for i := 0; i < len(res.limbs); i++ {
			val := uint64(res.limbs[i])*10 + carry
			res.limbs[i] = uint32(val)
			carry = val >> limbBits
		}
		if carry > 0 {
			res.limbs = append(res.limbs, uint32(carry))
		}
	}
	res.trim()
	return res, nil
}

// -------------------- Допоміжні --------------------

// trim прибирає старші нульові limb-и і зберігає канонічний нуль.
func (x *BigInt) trim() {
	for len(x.limbs) > 1 && x.limbs[len(x.limbs)-1] == 0 {
		x.limbs = x.limbs[:len(x.limbs)-1]
	}
	if len(x.limbs) == 0 {
		x.limbs = []uint32{0}
	}
}

// IsZero повертає true, якщо x == 0.
func (x *BigInt) IsZero() bool {
	return len(x.limbs) == 1 && x.limbs[0] == 0
}

// Cmp порівнює x і y: -1 якщо x<y, 0 якщо x==y, 1 якщо x>y.
func (x *BigInt) Cmp(y *BigInt) int {
	if len(x.limbs) != len(y.limbs) {
		if len(x.limbs) < len(y.limbs) {
			return -1
		}
		return 1
	}
	for i := len(x.limbs) - 1; i >= 0; i-- {
		if x.limbs[i] != y.limbs[i] {
			if x.limbs[i] < y.limbs[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Clone повертає незалежну копію числа.
func (x *BigInt) Clone() *BigInt {
	limbs := make([]uint32, len(x.limbs))
	copy(limbs, x.limbs)
	return &BigInt{limbs: limbs}
}

// bit повертає значення i-го біта числа (від молодшого до старшого).
func (x *BigInt) bit(i int) bool {
	limb := i / limbBits
	if limb >= len(x.limbs) {
		return false
	}
	return ((x.limbs[limb] >> (uint(i) % limbBits)) & 1) == 1
}

// lsh1 зсуває x на 1 біт вліво прямо в поточному об'єкті.
func (x *BigInt) lsh1() {
	carry := uint32(0)
	for i := 0; i < len(x.limbs); i++ {
		nextCarry := x.limbs[i] >> 31
		x.limbs[i] = (x.limbs[i] << 1) | carry
		carry = nextCarry
	}
	if carry > 0 {
		x.limbs = append(x.limbs, carry)
	}
	x.trim()
}

// -------------------- Арифметика --------------------

// Add повертає суму a + b.
func (a *BigInt) Add(b *BigInt) *BigInt {
	maxLen := len(a.limbs)
	if len(b.limbs) > maxLen {
		maxLen = len(b.limbs)
	}
	res := make([]uint32, maxLen+1)
	var carry uint64
	for i := 0; i < maxLen; i++ {
		sum := carry
		if i < len(a.limbs) {
			sum += uint64(a.limbs[i])
		}
		if i < len(b.limbs) {
			sum += uint64(b.limbs[i])
		}
		res[i] = uint32(sum)
		carry = sum >> limbBits
	}
	res[maxLen] = uint32(carry)
	out := &BigInt{limbs: res}
	out.trim()
	return out
}

// Sub повертає різницю a - b; панікує, якщо a < b.
func (a *BigInt) Sub(b *BigInt) *BigInt {
	if a.Cmp(b) < 0 {
		panic("negative result is not supported")
	}
	res := make([]uint32, len(a.limbs))
	var borrow uint64
	for i := 0; i < len(a.limbs); i++ {
		ai := uint64(a.limbs[i])
		bi := uint64(0)
		if i < len(b.limbs) {
			bi = uint64(b.limbs[i])
		}
		subtrahend := bi + borrow
		if ai < subtrahend {
			res[i] = uint32((uint64(1) << limbBits) + ai - subtrahend)
			borrow = 1
		} else {
			res[i] = uint32(ai - subtrahend)
			borrow = 0
		}
	}
	out := &BigInt{limbs: res}
	out.trim()
	return out
}

// Mul повертає добуток a * b.
func (a *BigInt) Mul(b *BigInt) *BigInt {
	if a.IsZero() || b.IsZero() {
		return New()
	}
	res := make([]uint32, len(a.limbs)+len(b.limbs)+1)
	for i := 0; i < len(a.limbs); i++ {
		var carry uint64
		ai := uint64(a.limbs[i])
		// Множення по limb-ах з накопиченням у res[i+j].
		for j := 0; j < len(b.limbs); j++ {
			sum := ai*uint64(b.limbs[j]) + uint64(res[i+j]) + carry
			res[i+j] = uint32(sum)
			carry = sum >> limbBits
		}
		// Протягуємо залишковий перенос вправо.
		k := i + len(b.limbs)
		for carry > 0 {
			if k >= len(res) {
				res = append(res, 0)
			}
			sum := uint64(res[k]) + carry
			res[k] = uint32(sum)
			carry = sum >> limbBits
			k++
		}
	}
	out := &BigInt{limbs: res}
	out.trim()
	return out
}

// Square повертає квадрат a.
func (a *BigInt) Square() *BigInt {
	// Зараз використовуємо прямий виклик множення.
	return a.Mul(a)
}

// Mod повертає остачу a mod m; панікує, якщо m == 0.
func (a *BigInt) Mod(m *BigInt) *BigInt {
	if m.IsZero() {
		panic("modulus is zero")
	}

	if a.IsZero() || a.Cmp(m) < 0 {
		return a.Clone()
	}
	if m.Cmp(NewFromUint64(1)) == 0 {
		return New() // a mod 1 == 0
	}

	rem := New()

	for i := a.BitLen() - 1; i >= 0; i-- {
		rem.lsh1() // rem <<= 1

		if a.bit(i) {
			if len(rem.limbs) == 0 {
				rem.limbs = []uint32{1}
			} else {
				rem.limbs[0] |= 1
			}
		}

		if rem.Cmp(m) >= 0 {
			rem = rem.Sub(m)
		}
	}

	return rem
}

// Exp повертає base^exp mod modulus (бінарне піднесення до степеня).
func (base *BigInt) Exp(exp, modulus *BigInt) *BigInt {
	if modulus.IsZero() {
		panic("modulus is zero")
	}

	result := NewFromUint64(1).Mod(modulus)
	if exp.IsZero() {
		return result
	}
	cur := base.Mod(modulus)
	e := exp.Clone()

	// Для кожного біта exp: множимо (якщо біт 1) і квадратуємо базу.
	for !e.IsZero() {
		// якщо exp непарний
		if e.limbs[0]&1 == 1 {
			result = result.Mul(cur).Mod(modulus)
		}
		// exp >>= 1
		e.rsh1()
		if !e.IsZero() {
			cur = cur.Square().Mod(modulus)
		}
	}
	return result
}

// rsh1 ділить x на 2 на місці.
func (x *BigInt) rsh1() {
	carry := uint32(0)
	for i := len(x.limbs) - 1; i >= 0; i-- {
		next := x.limbs[i] & 1
		x.limbs[i] = (x.limbs[i] >> 1) | (carry << 31)
		carry = next
	}
	x.trim()
}

// String повертає десяткове рядкове подання числа.
func (x *BigInt) String() string {
	if x.IsZero() {
		return "0"
	}

	tmp := x.Clone()
	var digits []byte

	// Багаторазово ділимо на 10 і збираємо цифри остач у зворотному порядку.
	for !tmp.IsZero() {
		// Залишок
		rem := uint64(0)
		// Ділення "довгим стовпчиком" у базі 2^32.
		for i := len(tmp.limbs) - 1; i >= 0; i-- {
			val := (rem << limbBits) | uint64(tmp.limbs[i])
			tmp.limbs[i] = uint32(val / 10)
			rem = val % 10
		}
		digits = append(digits, byte('0'+rem))
		tmp.trim()
	}

	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

// -------------------- Додатково --------------------

// BitLen повертає кількість значущих бітів числа.
func (x *BigInt) BitLen() int {
	if x.IsZero() {
		return 0
	}
	// Повні limb-и дають базову довжину, останній додає старші значущі біти.
	return (len(x.limbs)-1)*limbBits + bits.Len32(x.limbs[len(x.limbs)-1])
}
