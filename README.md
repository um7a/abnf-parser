# abnf-parser

Go library for Parsing the syntax defined in [ABNF](https://datatracker.ietf.org/doc/html/rfc5234).

## 1. Concept

### 1.1. FindFunc

This library provides the functions whose types are `FindFunc`.

```go
type FindFunc func(data []byte) (ends []int)
```

`FindFunc` finds the specific ABNF syntax such as [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [DIGIT](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [Concatenation](https://datatracker.ietf.org/doc/html/rfc5234#section-3.1), [Alternatives](https://datatracker.ietf.org/doc/html/rfc5234#section-3.2), etc. from `data`.  
If `FindFunc` finds the syntax, it returns the ends of the syntax as `ends`.

#### Example

For example, this library provides `FindAlpha` function.

```go
func FindAlpha(data []byte) (ends []int)
```

This function finds [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
When you call it with data `[]byte{ 'a', 'b', 'c', }`, it search [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from the beginning of `data`.  
Then it returns `[]int{1}` because `data` has `'a'` at the beginning.

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	ends := abnfp.FindAlpha(data)
	fmt.Printf("%v\n", ends) // -> [1]
}
```

Note that `FindFunc` only finds the syntax at the beginning.  
This means that when you call `FindAlpha` with data `[]byte{ '0', 'a' }`, it doesn't find `'a'` and returns `[]int{}` because there is `'0'` at the beginning of `data` and it is not [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1).

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'0', 'a'}
	ends := abnfp.FindAlpha(data)
	fmt.Printf("%v\n", ends) // -> []
}
```

### 1.2. Parse functions

The only utilities provided by this library other than `FindFunc` are the Parse functions.

```go
func ParseLongest(data []byte, findFunc FindFunc) (result ParseResult)
func ParseShortest(data []byte, findFunc FindFunc) (result ParseResult)
func ParseAll(data []byte, findFunc FindFunc) (results []ParseResult)
```

This functions parse the syntax specified by `findFunc` from `data` and it returns `ParseResult`.

```go
type ParseResult struct {
	Parsed    []byte
	Remaining []byte
}
```

`Parsed` is the parsed data and `Remaining` is the remaining data.

#### Example

For example, when `data` is `[]byte{'a', 'b', 'c'}` and `findFunc` is `FindAlpha`,  
`ParseLongest` function parse [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
Because `'a'` is [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), it returns `[]byte{'a'}` as `Parsed` and `[]byte{'b', 'c'}` as `Remaining`.

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	result := abnfp.ParseLongest(data, abnfp.FindAlpha)

	fmt.Printf("result.Parsed: %s\n", result.Parsed)
	// -> result.Parsed: a
	fmt.Printf("result.Remaining: %s\n", result.Remaining)
	// -> result.Remaining: bc
}
```

Depending on which Parse function is selected, the result will vary.  
For example, if you parse the data `[]byte{'a', 'b', 'c'}` as the ABNF syntax `*ALPHA`,  
The parsing result for each function is as follows.

| mode            | Parsed                                                                             | Remaining                                                                          |
| --------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `ParseLongest`  | `[]byte{'a', 'b' 'c'}`                                                             | `[]byte{}`                                                                         |
| `ParseShortest` | `[]byte{}`                                                                         | `[]byte{'a', 'b', 'c'}`                                                            |
| `ParseAll`      | `[]byte{}`,</br>`[]byte{'a'}`,</br>`[]byte{'a', 'b'}`,</br>`[]byte{'a', 'b', 'c'}` | `[]byte{'a', 'b', 'c'}`,</br>`[]byte{'b', 'c'}`,</br>`[]byte{'c'}`,</br>`[]byte{}` |

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}

	result := abnfp.ParseLongest(
		data,
		abnfp.NewFindVariableRepetition(abnfp.FindAlpha),
	)
	fmt.Printf("result.Parsed: %s, result.Remaining: %s\n", result.Parsed, result.Remaining)
	// -> result.Parsed: abc, result.Remaining:

	result = abnfp.ParseShortest(
		data,
		abnfp.NewFindVariableRepetition(abnfp.FindAlpha),
	)
	fmt.Printf("result.Parsed: %s, result.Remaining: %s\n", result.Parsed, result.Remaining)
	// -> result.Parsed: , result.Remaining: abc

	results := abnfp.ParseAll(
		data,
		abnfp.NewFindVariableRepetition(abnfp.FindAlpha),
	)
	for _, result := range results {
		fmt.Printf("result.Parsed: %s, result.Remaining: %s\n", result.Parsed, result.Remaining)
	}
	// -> result.Parsed: , result.Remaining: abc
	// -> result.Parsed: a, result.Remaining: bc
	// -> result.Parsed: ab, result.Remaining: c
	// -> result.Parsed: abc, result.Remaining:
}
```
