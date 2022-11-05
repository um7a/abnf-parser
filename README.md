# abnf-parser

Go library for Parsing the syntax defined in [ABNF](https://datatracker.ietf.org/doc/html/rfc5234).

## Concept

This library provides the functions whose type is `FindFunc`.

```go
type FindFunc func(data []byte) (found bool, end int)
```

`FindFunc` find the specific ABNF syntax from `data`.  
If `FindFunc` find the syntax, return `true` as `found` and the end of the syntax as `end`.

### Example

For exmaple, this library provides `FindAlpha` function.

```go
func FindAlpha(data []byte) (found bool, end int)
```

This function find [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
When you call `FindAlpha` with data `[]byte{ 'a', 'b', 'c', }`,  
`FindAlpha` function search [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from the start of `data`, and it find `'a'`.  
So it returns `true` as `found` and `0` as `end`.

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	found, end := FindAlpha(data)
	fmt.Printf("found: %v, end: %v\n", found, end) // -> true, 0
}
```

Please note that `FindFunc` find syntax only from the start of `data`.  
This means that when you call `FindAlpha` with data `[]byte{ '0', 'a' }`,  
`FindAlpha` doesn't find `'a'` and returns `false` as `found`.

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'0', 'a'}
	found, end := FindAlpha(data)
	fmt.Printf("found: %v, end: %v\n", found, end) // -> false, 0
}
```
