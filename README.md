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

### 1.2. Parse

The only utility provided by this library other than `FindFunc` is `Parse` function.

```go
func Parse(data []byte, findFunc FindFunc, mode ParseMode) (results []ParseResult)
```

This function parses the syntax specified by `finder` from `data []byte`,  
and it returns `[]ParseResult` whose element `ParseResult` is the type defined as the following.

```go
type ParseResult struct {
	Parsed    []byte
	Remaining []byte
}
```

`Parsed` is the parsed data and `Remaining` is the remaining data.

#### Example

For example, when `data` is `[]byte{'a', 'b', 'c'}` and `finder` is `FindAlpha`,  
`Parse` function parse [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
Because `'a'` is [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), it returns `[]byte{'a'}` as `Parsed` and `[]byte{'b', 'c'}` as `Remaining`.

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	results := Parse(data, abnfp.FindAlpha, abnfp.PARSE_LONGEST)
	fmt.Printf("results: %s\n", results)
	// -> results: [{a bc}]
}
```

#### ParseMode

The third argument of `Parse` function is `ParseMode`.  
You can use one of the following values.

- `PARSE_LONGEST`
- `PARSE_SHORTEST`
- `PARSE_ALL`

Depending on which mode is selected, the result of the `Parse` function will vary.  
For example, if you parse the data `[]byte{'a', 'b', 'c'}` as the ABNF syntax `*ALPHA`,  
The parsing results for each mode are as follows.

| mode           | Parsed                                                                             | Remaining                                                                          |
| -------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| PARSE_LONGEST  | `[]byte{'a', 'b' 'c'}`                                                             | `[]byte{}`                                                                         |
| PARSE_SHORTEST | `[]byte{}`                                                                         | `[]byte{'a', 'b', 'c'}`                                                            |
| PARSE_ALL      | `[]byte{}`,</br>`[]byte{'a'}`,</br>`[]byte{'a', 'b'}`,</br>`[]byte{'a', 'b', 'c'}` | `[]byte{'a', 'b', 'c'}`,</br>`[]byte{'b', 'c'}`,</br>`[]byte{'c'}`,</br>`[]byte{}` |

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	var results []ParseResult

	results = Parse(
		data,
		abnfp.NewFindVariableRepetition(abnfp.FindAlpha),
		abnfp.PARSE_LONGEST)

	for _, result := range results {
		fmt.Printf("result.Parsed: %s, result.Remaining\n", result.Parsed, result.Remaining)
	}

	results = Parse(
		data,
		abnfp.NewFindVariableRepetition(abnfp.FindAlpha),
		abnfp.PARSE_SHORTEST)

	for _, result := range results {
		fmt.Printf("result.Parsed: %s, result.Remaining\n", result.Parsed, result.Remaining)
	}

	results = Parse(
		data,
		abnfp.NewFindVariableRepetition(abnfp.FindAlpha),
		abnfp.PARSE_ALL)

	for _, result := range results {
		fmt.Printf("result.Parsed: %s, result.Remaining\n", result.Parsed, result.Remaining)
	}
}

```
