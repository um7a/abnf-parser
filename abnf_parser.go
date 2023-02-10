package abnfp

import "fmt"

type ParseResult struct {
	Parsed    []byte
	Remaining []byte
}

type VariableFinder interface {
	Find(data []byte) (found bool, end int)
	Copy() Finder
	Recalculate(data []byte) (found bool, end int)
}

type Finder interface {
	Find(data []byte) (found bool, end int)
	Copy() Finder
}

func Parse(data []byte, finder Finder) (parsed []byte, remaining []byte) {
	found, end := finder.Find(data)
	if !found {
		return []byte{}, data
	}
	return data[:end], data[end:]
}

type ByteFinder struct {
	target byte
}

func (finder ByteFinder) Find(data []byte) (found bool, end int) {
	if len(data) == 0 {
		return
	}
	if data[0] == finder.target {
		found = true
		end = 1
	}
	return
}

func (finder ByteFinder) Copy() Finder {
	return ByteFinder{target: finder.target}
}

func NewByteFinder(target byte) *ByteFinder {
	return &ByteFinder{target: target}
}

type BytesFinder struct {
	target []byte
}

func (finder BytesFinder) Find(data []byte) (found bool, end int) {
	if len(finder.target) == 0 {
		return
	}
	if len(data) < len(finder.target) {
		return
	}
	remaining := data
	for _, t := range finder.target {
		found, _ = NewByteFinder(t).Find(remaining)
		if !found {
			end = 0
			return
		}
		remaining = remaining[1:]
	}
	return true, len(finder.target)
}

func (finder BytesFinder) Copy() Finder {
	targetCopy := append([]byte{}, finder.target...)
	return BytesFinder{target: targetCopy}
}

func NewBytesFinder(target []byte) *BytesFinder {
	targetCopy := append([]byte{}, target...)
	return &BytesFinder{target: targetCopy}
}

// RFC5234 - 2.3. Terminal Values
// A concatenated string of such values is specified compactly, using a
// period (".") to indicate a separation of characters within that
// value.  Hence:
//
//  CRLF =  %d13.10
//

type CrLfFinder struct{}

func (finder CrLfFinder) Find(data []byte) (found bool, end int) {
	return NewBytesFinder([]byte("\r\n")).Find(data)
}

func (finder CrLfFinder) Copy() Finder {
	return CrLfFinder{}
}

func NewCrLfFinder() *CrLfFinder {
	return &CrLfFinder{}
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

type ConcatenationFinder struct {
	childFinders []Finder
	childEnds    []int
}

func (finder *ConcatenationFinder) Find(data []byte) (found bool, end int) {
	remaining := data

	for i := 0; i < len(finder.childFinders); i++ {
		fmt.Printf("### data: %s, remaining: %s, i: %v\n", data, remaining, i)
		// If the childFinder has already calculated, skip.
		if len(finder.childEnds) > i {
			remaining = remaining[finder.childEnds[i]:]
			continue
		}

		childFinder := finder.childFinders[i]
		childFound, childEnd := childFinder.Find(remaining)
		fmt.Printf("### finder count : %v, found : %v, end : %v\n", i, childFound, childEnd)
		if childFound {
			remaining = remaining[childEnd:]
			if i != 0 {
				childEnd += finder.childEnds[i-1]
			}
			finder.childEnds = append(finder.childEnds, childEnd)
			continue
		}
		fmt.Printf("### call Recalculate. data : %v\n", data)
		otherFound, _ := finder.Recalculate(data)
		fmt.Printf("### return from Recalculate. result: %v\n", otherFound)
		if otherFound {
			return finder.Find(data)
		} else {
			return false, 0
		}
	}
	return true, finder.childEnds[len(finder.childEnds)-1]
}

func (finder ConcatenationFinder) Copy() Finder {
	childFindersCopy := []Finder{}
	for _, childFinder := range finder.childFinders {
		childFindersCopy = append(childFindersCopy, childFinder.Copy())
	}
	return &ConcatenationFinder{childFinders: childFindersCopy, childEnds: []int{}}
}

func (finder *ConcatenationFinder) Recalculate(data []byte) (found bool, end int) {
	var remaining []byte
	childEnds := finder.childEnds
	for i := len(finder.childEnds) - 1; i >= 0; i-- {
		// restore remaining and childEnds
		if i == 0 {
			remaining = data
			childEnds = []int{}
		} else {
			remaining = data[childEnds[i-1]:]
			childEnds = childEnds[:i]
		}

		childFinder := finder.childFinders[i]
		switch cf := childFinder.(type) {
		case VariableFinder:
			otherFound, otherEnd := cf.Recalculate(remaining)
			fmt.Printf("### VariableFinder found. otherFound: %v, otherEnd: %v\n", otherFound, otherEnd)
			if !otherFound {
				continue
			}
			if i != 0 {
				otherEnd += childEnds[i-1]
			}
			childEnds = append(childEnds, otherEnd)
			finder.childEnds = childEnds
			return true, otherEnd
		}
	}
	return false, 0
}

func NewConcatenationFinder(finders []Finder) *ConcatenationFinder {
	findersCopy := []Finder{}
	for _, finder := range finders {
		findersCopy = append(findersCopy, finder.Copy())
	}
	return &ConcatenationFinder{childFinders: findersCopy}
}

// RFC5234 - 3.2. Alternatives: Rule1 / Rule2
// Elements separated by a forward slash ("/") are alternatives.
// Therefore,
//
//  foo / bar
//
// will accept <foo> or <bar>.

type AlternativesFinder struct {
	childFinders     []Finder
	remainingFinders []Finder
}

func (finder *AlternativesFinder) Find(data []byte) (found bool, end int) {
	finder.remainingFinders = finder.childFinders
	return finder.Recalculate(data)
}

func (finder AlternativesFinder) Copy() Finder {
	childFindersCopy := []Finder{}
	for _, childFinder := range finder.childFinders {
		childFindersCopy = append(childFindersCopy, childFinder.Copy())
	}
	return &AlternativesFinder{childFinders: childFindersCopy, remainingFinders: childFindersCopy}
}

func (finder *AlternativesFinder) Recalculate(data []byte) (found bool, end int) {
	for _, childFinder := range finder.remainingFinders {
		finder.remainingFinders = finder.remainingFinders[1:]
		childFound, childEnd := childFinder.Find(data)
		if childFound {
			found = childFound
			end = childEnd
			break
		}
	}
	return
}

func NewAlternativesFinder(finders []Finder) *AlternativesFinder {
	findersCopy := []Finder{}
	for _, finder := range finders {
		findersCopy = append(findersCopy, finder.Copy())
	}
	return &AlternativesFinder{childFinders: findersCopy, remainingFinders: findersCopy}
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

type ValueRangeAlternativesFinder struct {
	rangeStart byte
	rangeEnd   byte
}

func (finder *ValueRangeAlternativesFinder) Find(data []byte) (found bool, end int) {
	if len(data) == 0 {
		return
	}
	if data[0] >= finder.rangeStart && data[0] <= finder.rangeEnd {
		return true, 1
	}
	return false, 0
}

func (finder *ValueRangeAlternativesFinder) Copy() Finder {
	return &ValueRangeAlternativesFinder{rangeStart: finder.rangeStart, rangeEnd: finder.rangeEnd}
}

func NewValueRangeAlternativesFinder(rangeStart byte, rangeEnd byte) *ValueRangeAlternativesFinder {
	return &ValueRangeAlternativesFinder{rangeStart: rangeStart, rangeEnd: rangeEnd}
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

type VariableRepetitionMinMaxFinder struct {
	childFinder Finder
	min         int
	max         int
	childEnds   []int
}

func (finder *VariableRepetitionMinMaxFinder) Find(data []byte) (found bool, end int) {
	remaining := data
	childEnds := []int{}
	for {
		// If no more remaining return
		if len(remaining) == 0 {
			if len(finder.childEnds) == 0 && finder.min > 0 {
				return false, 0
			}
			break
		}

		// If match count is bigger than max, return
		if finder.max >= 0 && len(childEnds) >= finder.max {
			break
		}

		// NOTE
		// childFinder might have its state.
		// It's dangerous to use the same childFinder in this for loop. So copy it.
		childFinder := finder.childFinder.Copy()
		childFound, childEnd := childFinder.Find(remaining)
		if !childFound {
			if len(finder.childEnds) == 0 && finder.min > 0 {
				return false, 0
			}
			break
		}
		remaining = remaining[childEnd:]
		if len(childEnds) > 0 {
			childEnd += childEnds[len(childEnds)-1]
		}
		childEnds = append(childEnds, childEnd)
		if len(childEnds) >= finder.min {
			finder.childEnds = append(finder.childEnds, childEnd)
		}
	}
	if finder.min == 0 {
		finder.childEnds = append([]int{0}, finder.childEnds...)
	}
	return true, finder.childEnds[len(finder.childEnds)-1]
}

func (finder VariableRepetitionMinMaxFinder) Copy() Finder {
	return &VariableRepetitionMinMaxFinder{
		childFinder: finder.childFinder.Copy(),
		min:         finder.min,
		max:         finder.max,
		childEnds:   append([]int{}, finder.childEnds...),
	}
}

func (finder *VariableRepetitionMinMaxFinder) Recalculate(data []byte) (found bool, end int) {
	if len(finder.childEnds) == 1 {
		return false, 0
	}
	fmt.Printf("### len(finder.childEnds): %v\n", len(finder.childEnds))
	finder.childEnds = finder.childEnds[:len(finder.childEnds)-1]
	fmt.Printf("### len(finder.childEnds): %v\n", len(finder.childEnds))
	return true, finder.childEnds[len(finder.childEnds)-1]
}

func NewVariableRepetitionMinMaxFinder(min int, max int, finder Finder) *VariableRepetitionMinMaxFinder {
	return &VariableRepetitionMinMaxFinder{min: min, max: max, childFinder: finder}
}

func NewVariableRepetitionMinFinder(min int, finder Finder) *VariableRepetitionMinMaxFinder {
	return NewVariableRepetitionMinMaxFinder(min, -1, finder)
}

func NewVariableRepetitionMaxFinder(max int, finder Finder) *VariableRepetitionMinMaxFinder {
	return NewVariableRepetitionMinMaxFinder(0, max, finder)
}

func NewVariableRepetitionFinder(finder Finder) *VariableRepetitionMinMaxFinder {
	return NewVariableRepetitionMinMaxFinder(0, -1, finder)
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

func NewSpecificRepetitionFinder(count int, finder Finder) *VariableRepetitionMinMaxFinder {
	return NewVariableRepetitionMinMaxFinder(count, count, finder)
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

func NewOptionalSequenceFinder(finder Finder) *VariableRepetitionMinMaxFinder {
	return NewVariableRepetitionMinMaxFinder(0, 1, finder)
}

// RFC5234 - B.1. Core Rules
//
//  ALPHA = %x41-5A / %x61-7A ; A-Z / a-z
//

func NewAlphaFinder() *AlternativesFinder {
	return NewAlternativesFinder([]Finder{
		NewValueRangeAlternativesFinder(0x41, 0x5a),
		NewValueRangeAlternativesFinder(0x61, 0x7a),
	})
}

// RFC5234 - B.1. Core Rules
//
//  DIGIT = %x30-39 ; 0-9
//

func NewDigitFinder() *ValueRangeAlternativesFinder {
	return NewValueRangeAlternativesFinder(0x30, 0x39)
}

// RFC5234 - B.1. Core Rules
//
//  DQUOTE = %x22
//  ; " (Double Quote)
//

func NewDQuoteFinder() *ByteFinder {
	return NewByteFinder('"')
}

// RFC5234 - B.1. Core Rules
//
//  HEXDIG = DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
//

func NewHexDigFinder() *AlternativesFinder {
	return NewAlternativesFinder([]Finder{
		NewDigitFinder(),
		NewByteFinder('A'),
		NewByteFinder('B'),
		NewByteFinder('C'),
		NewByteFinder('D'),
		NewByteFinder('E'),
		NewByteFinder('F'),
	})
}

// RFC5234 - B.1. Core Rules
//
//  HTAB = %x09
//  ; horizontal tab
//

func NewHTabFinder() *ByteFinder {
	return NewByteFinder(0x09)
}

// RFC5234 - B.1. Core Rules
//
//  OCTET = %x00-FF
//  ; 8 bits of data
//

func NewOctetFinder() *ValueRangeAlternativesFinder {
	return NewValueRangeAlternativesFinder(0x00, 0xff)
}

// RFC5234 - B.1. Core Rules
//
//  SP =  %x20
//

func NewSpFinder() *ByteFinder {
	return NewByteFinder(0x20)
}

// RFC5234 - B.1. Core Rules
//
//  VCHAR = %x21-7E
//  ; visible (printing) characters
//

func NewVCharFinder() *ValueRangeAlternativesFinder {
	return NewValueRangeAlternativesFinder(0x21, 0x7e)
}
