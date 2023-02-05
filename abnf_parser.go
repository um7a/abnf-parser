package abnfp

type ParseMode int

type ParseResult struct {
	Parsed    []byte
	Remaining []byte
}

type FindFunc func(data []byte) (ends []int)

var PARSE_LONGEST ParseMode = 0
var PARSE_SHORTEST ParseMode = 1
var PARSE_ALL ParseMode = 2

func getBiggest(src []int) (biggest int) {
	for _, n := range src {
		if n > biggest {
			biggest = n
		}
	}
	return
}

func getSmallest(src []int) (smallest int) {
	for i, n := range src {
		if i == 0 {
			smallest = n
			continue
		}
		if n < smallest {
			smallest = n
		}
	}
	return
}

func sliceUnique(src []int) (unique []int) {
	m := map[int]bool{}
	for _, end := range src {
		if !m[end] {
			m[end] = true
			unique = append(unique, end)
		}
	}
	return unique
}

func Parse(data []byte, findFunc FindFunc, mode ParseMode) (results []ParseResult) {
	ends := findFunc(data)
	if len(ends) == 0 {
		return
	}
	if mode == PARSE_LONGEST {
		longestEnd := getBiggest(ends)
		ends = []int{longestEnd}
	} else if mode == PARSE_SHORTEST {
		shortestEnd := getSmallest(ends)
		ends = []int{shortestEnd}
	}
	for _, end := range ends {
		parsed := data[:end]
		remaining := data[end:]
		results = append(results, ParseResult{Parsed: parsed, Remaining: remaining})
	}
	return
}

func NewFindByte(target byte) FindFunc {
	findByte := func(data []byte) (ends []int) {
		if len(data) > 0 && data[0] == target {
			ends = []int{1}
		}
		return
	}
	return findByte
}

func NewFindBytes(target []byte) FindFunc {
	targetCopy := make([]byte, len(target))
	copy(targetCopy, target)
	findBytes := func(data []byte) (ends []int) {
		if len(targetCopy) > len(data) {
			return []int{}
		}

		var tmpEnds []int
		for i, t := range targetCopy {
			findByte := NewFindByte(t)
			if i == 0 {
				tmpEnds = findByte(data)
			} else {
				tmpEnds = findByte(data[i:])
			}
			if len(tmpEnds) == 0 {
				return
			}
		}
		ends = []int{len(targetCopy)}
		return
	}
	return findBytes
}

// RFC5234 - 2.3. Terminal Values
// A concatenated string of such values is specified compactly, using a
// period (".") to indicate a separation of characters within that
// value.  Hence:
//
//  CRLF =  %d13.10
//

func FindCrLf(data []byte) (ends []int) {
	findCrLf := NewFindBytes([]byte("\r\n"))
	return findCrLf(data)
}

// RFC5234 - 3.1. Concatenation: Rule1 Rule2
// A rule can define a simple, ordered string of values (i.e., a
// concatenation of contiguous characters) by listing a sequence of rule
// names. For example:
//
//  foo = %x61 ; a
//  bar = %x62 ; b
//  mumble = foo bar foo
//
// So that the rule <mumble> matches the lowercase string "aba".

func NewFindConcatenation(findFuncs []FindFunc) FindFunc {
	findConcatenation := func(data []byte) (ends []int) {
		for childCount, childFindFunc := range findFuncs {
			if childCount == 0 {
				ends = childFindFunc(data)
			} else {
				pastEnds := ends
				ends = []int{}
				for _, pastEnd := range pastEnds {
					var partialEnds []int
					var remaining []byte
					if pastEnd == 0 {
						remaining = data
					} else {
						remaining = data[pastEnd:]
					}
					partialEnds = childFindFunc(remaining)
					for i := 0; i < len(partialEnds); i++ {
						partialEnds[i] += pastEnd
					}
					if len(partialEnds) > 0 {
						ends = append(ends, partialEnds...)
					}
				}
			}
			if len(ends) == 0 {
				return
			}
		}
		ends = sliceUnique(ends)
		return
	}
	return findConcatenation
}

// RFC5234 - 3.2. Alternatives: Rule1 / Rule2
// Elements separated by a forward slash ("/") are alternatives.
// Therefore,
//
//  foo / bar
//
// will accept <foo> or <bar>.

func NewFindAlternatives(findFuncs []FindFunc) FindFunc {
	findAlternatives := func(data []byte) (ends []int) {
		var tmpEnds []int
		for _, childFindFunc := range findFuncs {
			tmpEnds = childFindFunc(data)
			if len(tmpEnds) != 0 {
				ends = append(ends, tmpEnds...)
			}
		}
		return
	}
	return findAlternatives
}

// RFC5234 - 3.4. Value Range Alternatives: %c##-##
// A range of alternative numeric values can be specified compactly,
// using a dash ("-") to indicate the range of alternative values.
// Hence:
//
//  DIGIT = %x30-39
//
// is equivalent to:
//
//  DIGIT = "0" / "1" / "2" / "3" / "4" / "5" / "6" / "7" / "8" / "9"
//

func NewFindValueRangeAlternatives(rangeStart byte, rangeEnd byte) FindFunc {
	findValueRangeAlternatives := func(data []byte) (ends []int) {
		if len(data) == 0 {
			return
		}
		if data[0] >= rangeStart && data[0] <= rangeEnd {
			ends = []int{1}
		}
		return
	}
	return findValueRangeAlternatives
}

// RFC5234 - 3.6. Variable Repetition: *Rule
// The operator "*" preceding an element indicates repetition. The full
// form is:
//
//  <a>*<b>element
//
// where <a> and <b> are optional decimal values, indicating at least
// <a> and at most <b> occurrences of the element.
//
// Default values are 0 and infinity so that *<element> allows any
// number, including zero; 1*<element> requires at least one;
// 3*3<element> allows exactly 3; and 1*2<element> allows one or two.

func NewFindVariableRepetitionMinMax(min int, max int, findFunc FindFunc) FindFunc {
	findVariableRepetitionMinMax := func(data []byte) (ends []int) {
		if min == 0 {
			ends = []int{0}
		}
		var currentEnds []int
		var matchCount int
		var pastEnds []int
		for {
			if matchCount == 0 {
				currentEnds = findFunc(data)
			} else {
				pastEnds = currentEnds
				currentEnds = []int{}
				for _, pastEnd := range pastEnds {
					currentEndsPerPastEnd := findFunc(data[pastEnd:])
					if len(currentEndsPerPastEnd) == 0 {
						continue
					}
					for i := 0; i < len(currentEndsPerPastEnd); i++ {
						currentEndsPerPastEnd[i] += pastEnd
					}
					currentEnds = append(currentEnds, currentEndsPerPastEnd...)
				}
			}
			if len(currentEnds) == 0 {
				break
			}
			matchCount++
			if matchCount >= min {
				ends = append(ends, currentEnds...)
			}
			if max >= 0 && matchCount >= max {
				break
			}
		}
		return
	}
	return findVariableRepetitionMinMax
}

func NewFindVariableRepetitionMin(min int, findFunc FindFunc) FindFunc {
	return NewFindVariableRepetitionMinMax(min, -1, findFunc)
}

func NewFindVariableRepetitionMax(max int, findFunc FindFunc) FindFunc {
	return NewFindVariableRepetitionMinMax(0, max, findFunc)
}

func NewFindVariableRepetition(findFunc FindFunc) FindFunc {
	return NewFindVariableRepetitionMinMax(0, -1, findFunc)
}

// RFC5234 - 3.7. Specific Repetition: nRule
// A rule of the form:
//
//  <n>element
//
// is equivalent to
//
//  <n>*<n>element
//

func NewFindSpecificRepetition(count int, findFunc FindFunc) FindFunc {
	findSpecificRepetition := func(data []byte) (ends []int) {
		min := count
		max := count
		findVariableRepetition := NewFindVariableRepetitionMinMax(min, max, findFunc)
		return findVariableRepetition(data)
	}
	return findSpecificRepetition
}

// RFC5234 - 3.8. Optional Sequence: [RULE]
// Square brackets enclose an optional element sequence:
//
//  [foo bar]
//
// is equivalent to
//
//  *1(foo bar).
//

func NewFindOptionalSequence(findFunc FindFunc) FindFunc {
	findOptionalSequence := func(data []byte) (ends []int) {
		min := 0
		max := 1
		findVariableRepetition := NewFindVariableRepetitionMinMax(min, max, findFunc)
		return findVariableRepetition(data)
	}
	return findOptionalSequence
}

// RFC5234 - B.1. Core Rules
//
//  ALPHA = %x41-5A / %x61-7A ; A-Z / a-z
//

func FindAlpha(data []byte) (ends []int) {
	findAlternatives := NewFindAlternatives([]FindFunc{
		NewFindValueRangeAlternatives(0x41, 0x5a),
		NewFindValueRangeAlternatives(0x61, 0x7a),
	})
	return findAlternatives(data)
}

// RFC5234 - B.1. Core Rules
//
//  DIGIT = %x30-39 ; 0-9
//

func FindDigit(data []byte) (ends []int) {
	findAlternatives := NewFindAlternatives([]FindFunc{
		NewFindValueRangeAlternatives(0x30, 0x39),
	})
	return findAlternatives(data)
}

// RFC5234 - B.1. Core Rules
//
//  DQUOTE = %x22
//  ; " (Double Quote)
//

func FindDQuote(data []byte) (ends []int) {
	findByte := NewFindByte(0x22)
	return findByte(data)
}

// RFC5234 - B.1. Core Rules
//
//  HEXDIG = DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
//

func FindHexDig(data []byte) (ends []int) {
	findAlternatives := NewFindAlternatives([]FindFunc{
		FindDigit,
		NewFindByte('A'),
		NewFindByte('B'),
		NewFindByte('C'),
		NewFindByte('D'),
		NewFindByte('E'),
		NewFindByte('F'),
	})
	return findAlternatives(data)
}

// RFC5234 - B.1. Core Rules
//
//  HTAB = %x09
//  ; horizontal tab
//

func FindHTab(data []byte) (ends []int) {
	findByte := NewFindByte(0x09)
	return findByte(data)
}

// RFC5234 - B.1. Core Rules
//
//  OCTET = %x00-FF
//  ; 8 bits of data
//

func FindOctet(data []byte) (ends []int) {
	findValueRangeAlternatives := NewFindValueRangeAlternatives(0x00, 0xFF)
	return findValueRangeAlternatives(data)
}

// RFC5234 - B.1. Core Rules
//
//  SP =  %x20
//

func FindSp(data []byte) (ends []int) {
	findByte := NewFindByte(0x20)
	return findByte(data)
}

// RFC5234 - B.1. Core Rules
//
//  VCHAR = %x21-7E
//  ; visible (printing) characters
//

func FindVChar(data []byte) (ends []int) {
	findValueRangeAlternatives := NewFindValueRangeAlternatives(0x21, 0x7E)
	return findValueRangeAlternatives(data)
}
