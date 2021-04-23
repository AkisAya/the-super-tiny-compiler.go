# the-super-tiny-compiler.go
Inspired by [the-super-tiny-compiler](https://github.com/jamiebuilds/the-super-tiny-compiler) which aims to show how most compilers work from end to end.

New to Golang and interestd in AST, I decided to implement the tiny compiler in go and further to support transforming C to Lisp too

## Goal:
Tranform statements bewteen List and C

### Example
1. add(1 substract(2 3)) -> add(1, substrac(2, 3))
2. add(1, substrac(2, 3)) -> add(1 substract(2 3))




## Related Repos
[the-super-tiny-compiler in go](https://github.com/hazbo/the-super-tiny-compiler)