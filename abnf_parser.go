package abnfp

type FindFunc func(data []byte) (found bool, end int)

func Parse(data []byte, finder FindFunc) (found bool, parsed []byte, remaining []byte) {
	end := 0
	found, end = finder(data)
	if !found {
		return
	}
	parsed = data[:end]
	remaining = data[end:]
	return
}

// RFC5234 - 2.3. Terminal Values
// A concatenated string of such values is specified compactly, using a
// period (".") to indicate a separation of characters within that
// value.  Hence:
//
//  CRLF =  %d13.10
//

func FindCrLf(data []byte) (found bool, end int) {
	findCrLf := CreateFind([]byte("\r\n"))
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

func CreateFindConcatenation(finders []FindFunc) FindFunc {
	findConcatenation := func(data []byte) (found bool, end int) {
		ruleFound := false
		ruleEnd := 0
		for _, findFunc := range finders {
			if ruleEnd == 0 {
				// Find first rule
				ruleFound, ruleEnd = findFunc(data)
			} else {
				pastRuleEnd := ruleEnd
				// Find rules other than the first one
				ruleFound, ruleEnd = findFunc(data[ruleEnd:])
				ruleEnd += pastRuleEnd
			}
			if !ruleFound {
				return
			}
		}
		found = true
		end = ruleEnd
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

func CreateFindAlternatives(finders []FindFunc) FindFunc {
	findAlternatives := func(data []byte) (found bool, end int) {
		for _, findFunc := range finders {
			found, end = findFunc(data)
			if found {
				break
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

func CreateFindValueRangeAlternatives(rangeStart byte, rangeEnd byte) FindFunc {
	findValueRangeAlternatives := func(data []byte) (found bool, end int) {
		if len(data) == 0 {
			return
		}
		if data[0] >= rangeStart && data[0] <= rangeEnd {
			found = true
			end = 1
			return
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

func CreateFindVariableRepetitionMinMax(min int, max int, finder FindFunc) FindFunc {
	findVariableRepetition := func(data []byte) (found bool, end int) {
		matchCount := 0
		for {
			if matchCount == 0 {
				found, end = finder(data)
			} else {
				pastEnd := end
				found, end = finder(data[pastEnd:])
				end += pastEnd
			}
			if !found {
				break
			}
			matchCount++
			if max >= 0 && matchCount >= max {
				break
			}
			if end >= len(data) {
				break
			}
		}

		if matchCount < min {
			found = false
			end = 0
			return
		}
		found = true
		return
	}
	return findVariableRepetition
}

func CreateFindVariableRepetitionMin(min int, finder FindFunc) FindFunc {
	max := -1
	return CreateFindVariableRepetitionMinMax(min, max, finder)
}

func CreateFindVariableRepetitionMax(max int, finder FindFunc) FindFunc {
	min := 0
	return CreateFindVariableRepetitionMinMax(min, max, finder)
}

func CreateFindVariableRepetition(finder FindFunc) FindFunc {
	min := 0
	max := -1
	return CreateFindVariableRepetitionMinMax(min, max, finder)
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

func CreateFindSpecificRepetition(count int, finder FindFunc) FindFunc {
	min := count
	max := count
	return CreateFindVariableRepetitionMinMax(min, max, finder)
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

func CreateFindOptionalSequence(finder FindFunc) FindFunc {
	min := 0
	max := 1
	return CreateFindVariableRepetitionMinMax(min, max, finder)
}

// RFC5234 - B.1. Core Rules
//
//  ALPHA = %x41-5A / %x61-7A ; A-Z / a-z
//

func FindAlpha(data []byte) (found bool, end int) {
	findAlpha := CreateFindAlternatives([]FindFunc{
		CreateFindValueRangeAlternatives(0x41, 0x5a),
		CreateFindValueRangeAlternatives(0x61, 0x7a),
	})
	return findAlpha(data)
}

// RFC5234 - B.1. Core Rules
//
//  DIGIT = %x30-39 ; 0-9
//

func FindDigit(data []byte) (found bool, end int) {
	findDigit := CreateFindValueRangeAlternatives(0x30, 0x39)
	return findDigit(data)
}

// RFC5234 - B.1. Core Rules
//
//  DQUOTE = %x22
//  ; " (Double Quote)
//

func FindDQuote(data []byte) (found bool, end int) {
	if len(data) > 0 && data[0] == 0x22 {
		found = true
		end = 1
	}
	return
}

// RFC5234 - B.1. Core Rules
//
//  HEXDIG = DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
//

func FindHexDig(data []byte) (found bool, end int) {
	findHexDig := CreateFindAlternatives([]FindFunc{
		FindDigit,
		createFindByte('A'),
		createFindByte('B'),
		createFindByte('C'),
		createFindByte('D'),
		createFindByte('E'),
		createFindByte('F'),
	})
	return findHexDig(data)
}

// RFC5234 - B.1. Core Rules
//
//  HTAB = %x09
//  ; horizontal tab
//

func FindHTab(data []byte) (found bool, end int) {
	if len(data) > 0 && data[0] == 0x09 {
		found = true
		end = 1
	}
	return
}

// RFC5234 - B.1. Core Rules
//
//  OCTET = %x00-FF
//  ; 8 bits of data
//

func FindOctet(data []byte) (found bool, end int) {
	if len(data) > 0 && data[0] >= 0x00 && data[0] <= 0xff {
		found = true
		end = 1
	}
	return
}

// RFC5234 - B.1. Core Rules
//
//  SP =  %x20
//

func FindSp(data []byte) (found bool, end int) {
	if len(data) > 0 && data[0] == 0x20 {
		found = true
		end = 1
	}
	return
}

// RFC5234 - B.1. Core Rules
//
//  VCHAR = %x21-7E
//  ; visible (printing) characters
//

func FindVChar(data []byte) (found bool, end int) {
	if len(data) > 0 && data[0] >= 0x21 && data[0] <= 0x7e {
		found = true
		end = 1
	}
	return
}

func createFindByte(target byte) FindFunc {
	findByte := func(data []byte) (found bool, end int) {
		if len(data) > 0 && data[0] == target {
			return true, 1
		}
		return false, 0
	}
	return findByte
}

func CreateFind(target []byte) FindFunc {
	find := func(data []byte) (found bool, end int) {
		for i, t := range target {
			findByte := createFindByte(t)
			if i == 0 {
				found, end = findByte(data)
			} else {
				found, end = findByte(data[end:])
			}
			if !found {
				found = false
				end = 0
				return
			}
			if i < len(target)-1 && i == len(data)-1 {
				found = false
				end = 0
				return
			}
		}
		end = len(target)
		return
	}
	return find
}
