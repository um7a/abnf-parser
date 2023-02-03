package abnfp

type ParseResult struct {
	parsed    []byte
	remaining []byte
}

func Parse(data []byte, finder Finder) (results []ParseResult) {
	ends := finder.find(data)
	if len(ends) == 0 {
		return
	}
	for _, end := range ends {
		parsed := data[:end]
		remaining := data[end:]
		results = append(results, ParseResult{parsed: parsed, remaining: remaining})
	}
	return
}

type Finder interface {
	find(data []byte) []int
	copy() Finder
}

type ByteFinder struct {
	target byte
}

func NewByteFinder(target byte) *ByteFinder {
	return &ByteFinder{target: target}
}

func (finder ByteFinder) find(data []byte) (ends []int) {
	if len(data) > 0 && data[0] == finder.target {
		ends = []int{1}
	}
	return
}

func (finder ByteFinder) copy() Finder {
	return NewByteFinder(finder.target)
}

type BytesFinder struct {
	target []byte
}

func NewBytesFinder(target []byte) *BytesFinder {
	targetCopy := make([]byte, len(target))
	copy(targetCopy, target)
	return &BytesFinder{target: targetCopy}
}

func (finder BytesFinder) find(data []byte) (ends []int) {
	if len(finder.target) > len(data) {
		return []int{}
	}

	var tmpEnds []int
	for i, t := range finder.target {
		byteFinder := ByteFinder{target: t}
		if i == 0 {
			tmpEnds = byteFinder.find(data)
		} else {
			tmpEnds = byteFinder.find(data[i:])
		}
		if len(tmpEnds) == 0 {
			return
		}
	}
	ends = []int{len(finder.target)}
	return
}

func (finder BytesFinder) copy() Finder {
	return NewBytesFinder(finder.target)
}

// RFC5234 - 2.3. Terminal Values
// A concatenated string of such values is specified compactly, using a
// period (".") to indicate a separation of characters within that
// value.  Hence:
//
//  CRLF =  %d13.10
//

type CrLfFinder struct {
}

func NewCrLfFinder() *CrLfFinder {
	return &CrLfFinder{}
}

func (finder CrLfFinder) find(data []byte) (ends []int) {
	bytesFinder := BytesFinder{target: []byte("\r\n")}
	return bytesFinder.find(data)
}

func (finder CrLfFinder) copy() Finder {
	return NewCrLfFinder()
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
}

func NewConcatenationFinder(finders []Finder) *ConcatenationFinder {
	var findersCopy []Finder

	for _, finder := range finders {
		findersCopy = append(findersCopy, finder.copy())
	}

	return &ConcatenationFinder{childFinders: findersCopy}
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

func (finder ConcatenationFinder) find(data []byte) (ends []int) {
	for childCount, childFinder := range finder.childFinders {
		if childCount == 0 {
			ends = childFinder.find(data)
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
				partialEnds = childFinder.find(remaining)
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

func (finder ConcatenationFinder) copy() Finder {
	return NewConcatenationFinder(finder.childFinders)
}

// RFC5234 - 3.2. Alternatives: Rule1 / Rule2
// Elements separated by a forward slash ("/") are alternatives.
// Therefore,
//
//  foo / bar
//
// will accept <foo> or <bar>.

type AlternativesFinder struct {
	childFinders []Finder
}

func NewAlternativesFinder(finders []Finder) *AlternativesFinder {
	var findersCopy []Finder

	for _, finder := range finders {
		findersCopy = append(findersCopy, finder.copy())
	}

	return &AlternativesFinder{childFinders: findersCopy}
}

func (finder AlternativesFinder) find(data []byte) (ends []int) {
	var tmpEnds []int
	for _, child := range finder.childFinders {
		tmpEnds = child.find(data)
		if len(tmpEnds) != 0 {
			ends = append(ends, tmpEnds...)
		}
	}
	return
}

func (finder AlternativesFinder) copy() Finder {
	return NewAlternativesFinder(finder.childFinders)
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

func NewValueRangeAlternativesFinder(rangeStart byte, rangeEnd byte) *ValueRangeAlternativesFinder {
	return &ValueRangeAlternativesFinder{rangeStart: rangeStart, rangeEnd: rangeEnd}
}

func (finder ValueRangeAlternativesFinder) find(data []byte) (ends []int) {
	if len(data) == 0 {
		return
	}
	if data[0] >= finder.rangeStart && data[0] <= finder.rangeEnd {
		ends = []int{1}
	}
	return
}

func (finder ValueRangeAlternativesFinder) copy() Finder {
	return NewValueRangeAlternativesFinder(finder.rangeStart, finder.rangeEnd)
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

type VariableRepetitionFinder struct {
	min         int
	max         int
	childFinder Finder
}

func NewVariableRepetitionMinMaxFinder(min int, max int, finder Finder) *VariableRepetitionFinder {
	finderCopy := finder.copy()
	return &VariableRepetitionFinder{min: min, max: max, childFinder: finderCopy}
}

func NewVariableRepetitionMinFinder(min int, finder Finder) *VariableRepetitionFinder {
	return NewVariableRepetitionMinMaxFinder(min, -1, finder)
}

func NewVariableRepetitionMaxFinder(max int, finder Finder) *VariableRepetitionFinder {
	return NewVariableRepetitionMinMaxFinder(0, max, finder)
}

func NewVariableRepetitionFinder(finder Finder) *VariableRepetitionFinder {
	return NewVariableRepetitionMinMaxFinder(0, -1, finder)
}

func (finder *VariableRepetitionFinder) find(data []byte) (ends []int) {
	if finder.min == 0 {
		ends = []int{0}
	}

	var tmpEnds []int
	var matchCount int
	var pastEnd int
	for {
		if matchCount == 0 {
			tmpEnds = finder.childFinder.find(data)
		} else {
			pastEnd = tmpEnds[0]
			tmpEnds = finder.childFinder.find(data[pastEnd:])
		}
		if len(tmpEnds) == 0 {
			break
		}
		matchCount++
		tmpEnds[0] += pastEnd
		if matchCount >= finder.min {
			ends = append(ends, tmpEnds...)
		}
		if finder.max >= 0 && matchCount >= finder.max {
			break
		}
		if tmpEnds[0] >= len(data) {
			break
		}
	}
	return
}

func (finder VariableRepetitionFinder) copy() Finder {
	return NewVariableRepetitionMinMaxFinder(finder.min, finder.max, finder.childFinder)
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

type SpecificRepetitionFinder struct {
	count       int
	childFinder Finder
}

func NewSpecificRepetitionFinder(count int, finder Finder) *SpecificRepetitionFinder {
	finderCopy := finder.copy()
	return &SpecificRepetitionFinder{count: count, childFinder: finderCopy}
}

func (finder SpecificRepetitionFinder) find(data []byte) []int {
	min := finder.count
	max := finder.count
	variableRepetitionFinder := NewVariableRepetitionMinMaxFinder(min, max, finder.childFinder)
	return variableRepetitionFinder.find(data)
}

func (finder SpecificRepetitionFinder) copy() Finder {
	return NewSpecificRepetitionFinder(finder.count, finder.childFinder)
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

type OptionalSequenceFinder struct {
	childFinder Finder
}

func NewOptionalSequenceFinder(finder Finder) *OptionalSequenceFinder {
	finderCopy := finder.copy()
	return &OptionalSequenceFinder{childFinder: finderCopy}
}

func (finder *OptionalSequenceFinder) find(data []byte) []int {
	min := 0
	max := 1
	variableRepetitionFinder := NewVariableRepetitionMinMaxFinder(min, max, finder.childFinder)
	return variableRepetitionFinder.find(data)
}

func (finder OptionalSequenceFinder) copy() Finder {
	return NewOptionalSequenceFinder(finder.childFinder)
}

// RFC5234 - B.1. Core Rules
//
//  ALPHA = %x41-5A / %x61-7A ; A-Z / a-z
//

type AlphaFinder struct {
}

func NewAlphaFinder() *AlphaFinder {
	return &AlphaFinder{}
}

func (finder AlphaFinder) find(data []byte) []int {
	alternativesFinder := NewAlternativesFinder([]Finder{
		NewValueRangeAlternativesFinder(0x41, 0x5a),
		NewValueRangeAlternativesFinder(0x61, 0x7a),
	})
	return alternativesFinder.find(data)
}

func (finder AlphaFinder) copy() Finder {
	return NewAlphaFinder()
}

// RFC5234 - B.1. Core Rules
//
//  DIGIT = %x30-39 ; 0-9
//

type DigitFinder struct {
}

func NewDigitFinder() *DigitFinder {
	return &DigitFinder{}
}

func (finder DigitFinder) find(data []byte) []int {
	valueRangeAlternativesFinder := NewValueRangeAlternativesFinder(0x30, 0x39)
	return valueRangeAlternativesFinder.find(data)
}

func (finder DigitFinder) copy() Finder {
	return NewDigitFinder()
}

// RFC5234 - B.1. Core Rules
//
//  DQUOTE = %x22
//  ; " (Double Quote)
//

type DQuoteFinder struct {
}

func NewDQuoteFinder() *DQuoteFinder {
	return &DQuoteFinder{}
}

func (finder DQuoteFinder) find(data []byte) []int {
	byteFinder := NewByteFinder(0x22)
	return byteFinder.find(data)
}

func (finder DQuoteFinder) copy() Finder {
	return NewDQuoteFinder()
}

// RFC5234 - B.1. Core Rules
//
//  HEXDIG = DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
//

type HexDigFinder struct {
}

func NewHexDigFinder() *HexDigFinder {
	return &HexDigFinder{}
}

func (finder HexDigFinder) find(data []byte) []int {
	alternativesFinder := NewAlternativesFinder([]Finder{
		NewDigitFinder(),
		NewByteFinder('A'),
		NewByteFinder('B'),
		NewByteFinder('C'),
		NewByteFinder('D'),
		NewByteFinder('E'),
		NewByteFinder('F'),
	})
	return alternativesFinder.find(data)
}

func (finder HexDigFinder) copy() Finder {
	return *NewHexDigFinder()
}

// RFC5234 - B.1. Core Rules
//
//  HTAB = %x09
//  ; horizontal tab
//

type HTabFinder struct {
}

func NewHTabFinder() *HTabFinder {
	return &HTabFinder{}
}

func (finder HTabFinder) find(data []byte) []int {
	byteFinder := NewByteFinder(0x09)
	return byteFinder.find(data)
}

func (finder HTabFinder) copy() Finder {
	return *NewHTabFinder()
}

// RFC5234 - B.1. Core Rules
//
//  OCTET = %x00-FF
//  ; 8 bits of data
//

type OctetFinder struct {
}

func NewOctetFinder() *OctetFinder {
	return &OctetFinder{}
}

func (finder OctetFinder) find(data []byte) []int {
	valueRangeAlternativesFinder := NewValueRangeAlternativesFinder(0x00, 0xFF)
	return valueRangeAlternativesFinder.find(data)
}

func (finder OctetFinder) copy() Finder {
	return *NewOctetFinder()
}

// RFC5234 - B.1. Core Rules
//
//  SP =  %x20
//

type SpFinder struct {
}

func NewSpFinder() *SpFinder {
	return &SpFinder{}
}

func (finder SpFinder) find(data []byte) []int {
	byteFinder := NewByteFinder(0x20)
	return byteFinder.find(data)
}

func (finder SpFinder) copy() Finder {
	return NewSpFinder()
}

// RFC5234 - B.1. Core Rules
//
//  VCHAR = %x21-7E
//  ; visible (printing) characters
//

type VCharFinder struct {
}

func NewVCharFinder() *VCharFinder {
	return &VCharFinder{}
}

func (finder VCharFinder) find(data []byte) []int {
	valueRangeAlternativesFinder := NewValueRangeAlternativesFinder(0x21, 0x7E)
	return valueRangeAlternativesFinder.find(data)
}

func (finder VCharFinder) copy() Finder {
	return *NewVCharFinder()
}
