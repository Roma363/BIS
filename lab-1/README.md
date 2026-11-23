# Lab 1 — BigInt and fast modular exponentiation

This project implements non-negative big integer arithmetic for the lab work:

- addition
- multiplication
- squaring
- modulo reduction
- fast modular exponentiation (`square-and-multiply`)

The implementation is aimed at inputs up to about 50 decimal digits.

## Run

```powershell
go test ./...
go run .
```

## Notes

- Numbers are non-negative only.
- `main.go` demonstrates the operations and compares them with Go's `math/big`.
- Tests cross-check the custom implementation against `math/big` for deterministic and random cases.

