# abnf-parser

Go library for Parsing the syntax defined in [ABNF](https://datatracker.ietf.org/doc/html/rfc5234).

## 1. Concept

### 1.1. Finder

This library provides Finders.  
Finder is the struct that has the methods named `Find` and `Copy`.

```go
type Finder interface {
	Find(data []byte) (found bool, end int)
	Copy() Finder
}
```

#### 1.1.1. Finder.Find

`Finder.Find(data []byte) (found bool, end int)` method finds the specific ABNF syntax such as [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [DIGIT](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1), [Concatenation](https://datatracker.ietf.org/doc/html/rfc5234#section-3.1), [Alternatives](https://datatracker.ietf.org/doc/html/rfc5234#section-3.2), etc. from the beginning of the `data`.  
If it finds the syntax, it returns `true` as `found` and the end of the syntax as `end`.

##### Example

For example, this library provides `AlphaFinder`.  
This Finder finds [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data`.  
When you call its `Find` method with data `[]byte{ 'a', 'b', 'c', }`, it search [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from the beginning of `data`. Then it returns `true, 1` because `data` has `'a'` at the beginning.

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	alpha := abnfp.NewAlphaFinder()
	found, end := alpha.Find(data)
	fmt.Printf("%v, %v\n", found, end) // -> true, 1
}
```

#### 1.1.2. Finder.Copy

`Finder.Copy() Finder` method returns the copy of the Finder.  
Some Finders have other finders as its child. This method also copies them. Note that this is the deep copy.

##### Example

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	alpha1 := abnfp.NewAlphaFinder()
	alpha2 := alpha1.Copy()

	var data []byte = []byte{'a', 'b', 'c'}
	found, end := alpha2.Find(data)
	fmt.Printf("%v, %v\n", found, end) // -> true, 1
}
```

### 1.2. Parse

The only utility provided by this library other than `Finder` is `Parse` function.

```go
func Parse(data []byte, finder Finder) (parsed []byte, remaining []byte)
```

`Parse` function returns the parsed data as `parsed` and the remaining data as `remaining`.

#### Example

For example, when `data` is `[]byte{'a', 'b', 'c'}` and `finder` is `AlphaFinder`,  
`Parse` function parses [ALPHA](https://datatracker.ietf.org/doc/html/rfc5234#appendix-B.1) from `data` and returns `[]byte{'a'}, []byte{'b', 'c'}`

```go
package main

import (
	"fmt"

	abnfp "github.com/um7a/abnf-parser"
)

func main() {
	var data []byte = []byte{'a', 'b', 'c'}
	parsed, remaining := abnfp.Parse(data, abnfp.NewAlphaFinder())
	fmt.Printf("parsed: %s, remaining: %s\n", parsed, remaining)
	// -> parsed: a, remaining: bc
}
```
