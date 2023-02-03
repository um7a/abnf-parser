# abnf-parser

Go library for Parsing the syntax defined in [ABNF](https://datatracker.ietf.org/doc/html/rfc5234).

## 1. Concept

### 1.1. Finder

This library provides Finders.  
Finder is the struct which has the methods named `find` and `copy`.

```go
type Finder interface {
	find(data []byte) []int
	copy() Finder
}
```

#### 1.1.1. Finder.find

`Finder.find(data)` method finds the specific ABNF syntax such as [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [DIGIT](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [Concatenation](https://datatracker.ietf.org/doc/html/rfc5234#section-3.1), [Alternatives](https://datatracker.ietf.org/doc/html/rfc5234#section-3.2), etc. from `data`.  
If it find the syntax, it return `[]int` which has the ends of the syntax.

##### Example

For example, this library provides `AlphaFinder`.  
This Finder finds [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
When you call it's `find` method with data `[]byte{ 'a', 'b', 'c', }`, it search [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from the beginning of `data`. Then it returns `[]int{1}` because `data` has `'a'` at the beginning.

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	alpha := abnfp.NewAlphaFinder()
	ends := alpha.find(data)
	fmt.Printf("%v\n", ends) // -> {1}
}
```

Note that `find` method only finds the syntax at the beginning.  
This means that when you call `Alpha.find` method with data `[]byte{ '0', 'a' }`, it doesn't find `'a'` and returns `false` because there is `'0'` at the beginning of `data` and it is not [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1).

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'0', 'a'}
	alpha := abnfp.NewAlphaFinder()
	ends := alpha.find(data)
	fmt.Printf("%v\n", ends) // -> {}
}
```

#### 1.1.2. Finder.copy

`Finder.copy` method returns the copy of the Finder.  
Some of Finders has other finders as it's child. it also copies them. Note that this copy is the deep copy.

##### Example

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	alpha1 := abnfp.NewAlphaFinder()
	alpha2 := alpha1.copy()

	var data []byte = []byte{'a', 'b', 'c'}
	ends := alpha2.find(data)
	fmt.Printf("%v\n", ends) // -> {1}
}
```

### 1.2. Parse

The only utility provided by this library other than `Finder` is `Parse` function.

```go
func Parse(data []byte, finder Finder) (results []ParseResult)
```

`Parse` function returns `[]ParseResult`. It's element `ParseResult` is the type like the following.

```go
type ParseResult struct {
	parsed    []byte
	remaining []byte
}
```

This function parses the syntax specified by `finder` from `data []byte`.  
If it find the syntax, return `ParseResult` whose `parsed` is the parsed data and `remaining` is the remaining data.

#### Example

For example, when `data` is `[]byte{'a', 'b', 'c'}` and `finder` is `AlphaFinder`,  
`Parse` function parses [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data` and returns `{{parsed: {'a'}, remaining: {'b', 'c'}}}`

```go
package main

import (
	"fmt"

	"github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	result := Parse(data, abnfp.FindAlpha)
	fmt.Printf("parsed: %s, remaining: %s\n", result.parsed, result.remaining)
	// -> {'a'}, {'b', 'c'}
}
```
