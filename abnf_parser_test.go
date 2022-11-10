package abnfp

import "testing"

func equals[C comparable](testName string, t *testing.T, expected C, actual C) {
	if actual != expected {
		t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
	}
}

func TestParse(t *testing.T) {
	bytesEqual := func(testName string, t *testing.T, expected []byte, actual []byte) {
		if len(expected) != len(actual) {
			t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
		}
		for i, e := range expected {
			if e != actual[i] {
				t.Errorf("%v: expected: %v, actual: %v", testName, expected[i], actual[i])
			}
		}
	}

	type TestCase struct {
		testName          string
		data              []byte
		finder            FindFunc
		expectedFound     bool
		expectedParsed    []byte
		expectedRemaining []byte
	}

	tests := []TestCase{
		{
			testName:          "Finder: ALPHA, data: []byte{}",
			data:              []byte{},
			finder:            FindAlpha,
			expectedFound:     false,
			expectedParsed:    []byte{},
			expectedRemaining: []byte{},
		},
		{
			testName:          "Finder: ALPHA, data: []byte(\"a\")",
			data:              []byte("a"),
			finder:            FindAlpha,
			expectedFound:     true,
			expectedParsed:    []byte("a"),
			expectedRemaining: []byte{},
		},
		{
			testName:          "Finder: ALPHA, data: []byte(\"abc\")",
			data:              []byte("abc"),
			finder:            FindAlpha,
			expectedFound:     true,
			expectedParsed:    []byte("a"),
			expectedRemaining: []byte("bc"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualParsed, actualRemaining := Parse(testCase.data, testCase.finder)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			bytesEqual(testCase.testName, t, testCase.expectedParsed, actualParsed)
			bytesEqual(testCase.testName, t, testCase.expectedRemaining, actualRemaining)
		})
	}
}

func TestFindCrLf(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}
	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\")",
			data:          []byte("a"),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"\\r\")",
			data:          []byte("\r"),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"\\r\\n\")",
			data:          []byte("\r\n"),
			expectedFound: true,
			expectedEnd:   2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindCrLf(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindConcatenation(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		finders       []FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		//
		// Concatenation: ALPHA ALPHA
		//
		{
			testName:      "Concatenation: ALPHA ALPHA, data: []byte{}",
			data:          []byte{},
			finders:       []FindFunc{FindAlpha, FindAlpha},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA ALPHA, data: []byte{'a'}",
			data:          []byte{'a'},
			finders:       []FindFunc{FindAlpha, FindAlpha},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA ALPHA, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			finders:       []FindFunc{FindAlpha, FindAlpha},
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "Concatenation: ALPHA ALPHA, data: []byte{'1'}",
			data:          []byte{'1'},
			finders:       []FindFunc{FindAlpha, FindAlpha},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA ALPHA, data: []byte{'1', '2'}",
			data:          []byte{'1', '2'},
			finders:       []FindFunc{FindAlpha, FindAlpha},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA ALPHA, data: []byte{'a', '1'}",
			data:          []byte{'a', '1'},
			finders:       []FindFunc{FindAlpha, FindAlpha},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA ALPHA, data: []byte{'1', 'a'}",
			data:          []byte{'1', 'a'},
			finders:       []FindFunc{FindAlpha, FindAlpha},
			expectedFound: false,
			expectedEnd:   0,
		},
		//
		// Concatenation: ALPHA DIGIT
		//
		{
			testName:      "Concatenation: ALPHA DIGIT, data: []byte{}",
			data:          []byte{},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA DIGIT, data: []byte{'a'}",
			data:          []byte{'a'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA DIGIT, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA DIGIT, data: []byte{'1'}",
			data:          []byte{'1'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA DIGIT, data: []byte{'1', '2'}",
			data:          []byte{'1', '2'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Concatenation: ALPHA DIGIT, data: []byte{'a', '1'}",
			data:          []byte{'a', '1'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "Concatenation: ALPHA DIGIT, data: []byte{'1', 'a'}",
			data:          []byte{'1', 'a'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: false,
			expectedEnd:   0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findConcatenation := CreateFindConcatenation(testCase.finders)
			actualFound, actualEnd := findConcatenation(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindAlternatives(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		finders       []FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		//
		// Alternatives: ALPHA DIGIT
		//
		{
			testName:      "Alternatives: ALPHA DIGIT, data: []byte{}",
			data:          []byte{},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "Alternatives: ALPHA DIGIT, data: []byte{'a'}",
			data:          []byte{'a'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "Alternatives: ALPHA DIGIT, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "Alternatives: ALPHA DIGIT, data: []byte{'1'}",
			data:          []byte{'1'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "Alternatives: ALPHA DIGIT, data: []byte{'1', '2'}",
			data:          []byte{'1', '2'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "Alternatives: ALPHA DIGIT, data: []byte{'a', '1'}",
			data:          []byte{'a', '1'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "Alternatives: ALPHA DIGIT, data: []byte{'1', 'a'}",
			data:          []byte{'1', 'a'},
			finders:       []FindFunc{FindAlpha, FindDigit},
			expectedFound: true,
			expectedEnd:   1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findAlternatives := CreateFindAlternatives(testCase.finders)
			actualFound, actualEnd := findAlternatives(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindValueRangeAlternatives(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		rangeStart    byte
		rangeEnd      byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "RangeStart: 'a', RangeEnd 'x', data: []byte{}",
			data:          []byte{},
			rangeStart:    'a',
			rangeEnd:      'x',
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "RangeStart: 'a', RangeEnd 'x', data: []byte{'`'}",
			data:          []byte{'`'},
			rangeStart:    'a',
			rangeEnd:      'x',
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "RangeStart: 'a', RangeEnd 'x', data: []byte{'a'}",
			data:          []byte{'a'},
			rangeStart:    'a',
			rangeEnd:      'x',
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "RangeStart: 'a', RangeEnd 'x', data: []byte{'x'}",
			data:          []byte{'x'},
			rangeStart:    'a',
			rangeEnd:      'x',
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "RangeStart: 'a', RangeEnd 'x', data: []byte{'{'}",
			data:          []byte{'{'},
			rangeStart:    'a',
			rangeEnd:      'x',
			expectedFound: false,
			expectedEnd:   0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findValueRangeAlternatives := CreateFindValueRangeAlternatives(testCase.rangeStart, testCase.rangeEnd)
			actualFound, actualEnd := findValueRangeAlternatives(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindVariableRepetitionMinMax(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		min           int
		max           int
		finder        FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "min: 0, max: 1, finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			min:           0,
			max:           1,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "min: 1, max: 1, finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			min:           1,
			max:           1,
			finder:        FindAlpha,
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "min: 1, max: 1, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			min:           1,
			max:           1,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "min: 1, max: 2, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			min:           1,
			max:           2,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "min: 2, max: 3, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			min:           2,
			max:           3,
			finder:        FindAlpha,
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "min: 2, max: 3, finder: FindAlpha, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			min:           2,
			max:           3,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "min: 2, max: 3, finder: FindAlpha, data: []byte{'a', 'b', 'c', 'd'}",
			data:          []byte{'a', 'b', 'c', 'd'},
			min:           2,
			max:           3,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   3,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findVariableRepetitionMinMax := CreateFindVariableRepetitionMinMax(testCase.min, testCase.max, testCase.finder)
			actualFound, actualEnd := findVariableRepetitionMinMax(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindVariableRepetitionMin(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		min           int
		finder        FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "min: 0, finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			min:           0,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "min: 1, finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			min:           1,
			finder:        FindAlpha,
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "min: 1, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			min:           1,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "min: 2, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			min:           2,
			finder:        FindAlpha,
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "min: 2, finder: FindAlpha, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			min:           2,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "min: 2, finder: FindAlpha, data: []byte{'a', 'b', 'c', 'd'}",
			data:          []byte{'a', 'b', 'c', 'd'},
			min:           2,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   4,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findVariableRepetitionMin := CreateFindVariableRepetitionMin(testCase.min, testCase.finder)
			actualFound, actualEnd := findVariableRepetitionMin(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindVariableRepetitionMax(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		max           int
		finder        FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "max: 1, finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			max:           1,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "max: 1, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			max:           1,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "max: 2, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			max:           2,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "max: 3, finder: FindAlpha, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			max:           3,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findVariableRepetitionMax := CreateFindVariableRepetitionMax(testCase.max, testCase.finder)
			actualFound, actualEnd := findVariableRepetitionMax(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindVariableRepetition(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		finder        FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "finder: FindAlpha, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "finder: FindAlpha, data: []byte{'a', 'b', 'c'}",
			data:          []byte{'a', 'b', 'c'},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   3,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findVariableRepetition := CreateFindVariableRepetition(testCase.finder)
			actualFound, actualEnd := findVariableRepetition(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindSpecificRepetition(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		count         int
		finder        FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "count: 0, finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			count:         0,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "count: 1, finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			count:         1,
			finder:        FindAlpha,
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "count: 1, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			count:         1,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "count: 1, finder: FindAlpha, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			count:         1,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "count: 2, finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			count:         2,
			finder:        FindAlpha,
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "count: 2, finder: FindAlpha, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			count:         2,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "count: 2, finder: FindAlpha, data: []byte{'a', 'b', 'c'}",
			data:          []byte{'a', 'b', 'c'},
			count:         2,
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findSpecificRepetition := CreateFindSpecificRepetition(testCase.count, testCase.finder)
			actualFound, actualEnd := findSpecificRepetition(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFindOptionalSequence(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		finder        FindFunc
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "finder: FindAlpha, data: []byte{}",
			data:          []byte{},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "finder: FindAlpha, data: []byte{'a'}",
			data:          []byte{'a'},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "finder: FindAlpha, data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			finder:        FindAlpha,
			expectedFound: true,
			expectedEnd:   1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			findOptionalSequence := CreateFindOptionalSequence(testCase.finder)
			actualFound, actualEnd := findOptionalSequence(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindAlpha(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a'}",
			data:          []byte{'a'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'1'}",
			data:          []byte{'1'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'1', '2'}",
			data:          []byte{'1', '2'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'A'}",
			data:          []byte{'A'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'A', 'B'}",
			data:          []byte{'A', 'B'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'1', 'a'}",
			data:          []byte{'1', 'a'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a', '1'}",
			data:          []byte{'a', '1'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'1', 'A'}",
			data:          []byte{'1', 'A'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'A', '1'}",
			data:          []byte{'A', '1'},
			expectedFound: true,
			expectedEnd:   1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindAlpha(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindDigit(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a'}",
			data:          []byte{'a'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'1'}",
			data:          []byte{'1'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'1', '2'}",
			data:          []byte{'1', '2'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'A'}",
			data:          []byte{'A'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'A', 'B'}",
			data:          []byte{'A', 'B'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'1', 'a'}",
			data:          []byte{'1', 'a'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'a', '1'}",
			data:          []byte{'a', '1'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'1', 'A'}",
			data:          []byte{'1', 'A'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'A', '1'}",
			data:          []byte{'A', '1'},
			expectedFound: false,
			expectedEnd:   0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindDigit(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindDQuote(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\")",
			data:          []byte("a"),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"'\")",
			data:          []byte("'"),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"\"\")",
			data:          []byte("\""),
			expectedFound: true,
			expectedEnd:   1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindDQuote(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindHexDig(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a'}",
			data:          []byte{'a'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a', 'b'}",
			data:          []byte{'a', 'b'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'1'}",
			data:          []byte{'1'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'1', '2'}",
			data:          []byte{'1', '2'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'A'}",
			data:          []byte{'A'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'A', 'B'}",
			data:          []byte{'A', 'B'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'1', 'a'}",
			data:          []byte{'1', 'a'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'a', '1'}",
			data:          []byte{'a', '1'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'1', 'A'}",
			data:          []byte{'1', 'A'},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{'A', '1'}",
			data:          []byte{'A', '1'},
			expectedFound: true,
			expectedEnd:   1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindHexDig(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindHTab(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\")",
			data:          []byte("a"),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x20}",
			data:          []byte{0x20},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x09}",
			data:          []byte{0x09},
			expectedFound: true,
			expectedEnd:   1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindHTab(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindOctet(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x00}",
			data:          []byte{0x00},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0xff}",
			data:          []byte{0xff},
			expectedFound: true,
			expectedEnd:   1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindOctet(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindSp(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\")",
			data:          []byte("a"),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x20}",
			data:          []byte{0x20},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0x09}",
			data:          []byte{0x09},
			expectedFound: false,
			expectedEnd:   0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindSp(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestFindVChar(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}",
			data:          []byte{},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x20}",
			data:          []byte{0x20},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x21}",
			data:          []byte{0x21},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0x7e}",
			data:          []byte{0x7e},
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0x7f}",
			data:          []byte{0x7f},
			expectedFound: false,
			expectedEnd:   0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := FindVChar(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestCreateFind(t *testing.T) {
	type TestCase struct {
		testName      string
		data          []byte
		target        []byte
		expectedFound bool
		expectedEnd   int
	}

	tests := []TestCase{
		{
			testName:      "data: []byte{}, target: []byte{'b', 'c'}",
			data:          []byte{},
			target:        []byte{'b', 'c'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a', 'b'}, target: []byte{'b', 'c'}",
			data:          []byte{'a', 'b'},
			target:        []byte{'b', 'c'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'b', 'c'}, target: []byte{'b', 'c'}",
			data:          []byte{'b', 'c'},
			target:        []byte{'b', 'c'},
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "data: []byte{'c', 'd'}, target: []byte{'b', 'c'}",
			data:          []byte{'c', 'd'},
			target:        []byte{'b', 'c'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'a', 'b', 'c'}, target: []byte{'b', 'c'}",
			data:          []byte{'a', 'b', 'c'},
			target:        []byte{'b', 'c'},
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{'b', 'c', 'd'}, target: []byte{'b', 'c'}",
			data:          []byte{'b', 'c', 'd'},
			target:        []byte{'b', 'c'},
			expectedFound: true,
			expectedEnd:   2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			find := CreateFind(testCase.target)
			actualFound, actualEnd := find(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}
