# abnf-parser

Go library for Parsing the syntax defined in [ABNF](https://datatracker.ietf.org/doc/html/rfc5234).

## 1. Concept

### 1.1. FindFunc

This library provides the functions whose type are `FindFunc`.

```go
type FindFunc func(data []byte) (found bool, end int)
```

`FindFunc` finds the specific ABNF syntax such as [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [DIGIT](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [Concatenation](https://datatracker.ietf.org/doc/html/rfc5234#section-3.1), [Alternatives](https://datatracker.ietf.org/doc/html/rfc5234#section-3.2), etc. from `data`.  
If `FindFunc` find the syntax, it return `true` as `found` and the end of syntax as `end`.

### Example

For example, this library provides `FindAlpha` function.

```go
func FindAlpha(data []byte) (found bool, end int)
```

This function finds [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
When you call `FindAlpha` with data `[]byte{ 'a', 'b', 'c', }`, it search [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from the beginning of `data`. Then it returns `true` as `found` and `1` as `end` because `data` has `'a'` at the beginning.

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	found, end := abnfp.FindAlpha(data)
	fmt.Printf("found: %v, end: %v\n", found, end) // -> true, 1
}
```

Note that `FindFunc` only finds the syntax at the beginning.  
This means that when you call `FindAlpha` with data `[]byte{ '0', 'a' }`, it doesn't find `'a'` and returns `false` because there is `'0'` at the beginning of `data` and it is not [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1).

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'0', 'a'}
	found, end := abnfp.FindAlpha(data)
	fmt.Printf("found: %v, end: %v\n", found, end) // -> false, 0
}
```

### 1.2. Parse

The only utility provided by this library other than `FindFunc` is `Parse` function.

```go
func Parse(data []byte, finder FindFunc) (found bool, parsed []byte, remaining []byte)
```

This function parses the syntax specified by `finder FindFunc` from `data []byte`.  
If this function find the syntax, return `true` as `found`, the parsed data as `parsed` and the remaining data as `remaining`.

### Example

For example, when `data` is `[]byte{'a', 'b', 'c'}` and `finder` is `FindAlpha`,  
`Parse` function parse [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
Because `'a'` is [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), it returns `true` as `found`, `[]byte{'a'}` as `parsed`, `[]byte{'b', 'c'}` as `remaining`.

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	found, parsed, remaining := Parse(data, abnfp.FindAlpha)
	fmt.Printf("found: %v, parsed: %s, remaining: %s\n", found, parsed, remaining)
	// -> true, a, bc
}
```
