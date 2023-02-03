package abnfp

import "testing"

type TestCase struct {
	testName     string
	data         []byte
	finder       Finder
	expectedEnds []int
}

func sliceEquals[C comparable](testName string, t *testing.T, expected []C, actual []C) {
	if len(expected) != len(actual) {
		t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
	}
	for i, e := range expected {
		if e != actual[i] {
			t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
		}
	}
}

func equals[C comparable](testName string, t *testing.T, expected C, actual C) {
	if actual != expected {
		t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
	}
}

func execTest(tests []TestCase, t *testing.T) {
	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualEnds := testCase.finder.Find(testCase.data)
			sliceEquals(testCase.testName, t, testCase.expectedEnds, actualEnds)
		})
	}
}

func TestParse(t *testing.T) {
	type TestCase struct {
		testName             string
		data                 []byte
		finder               Finder
		expectedParseResults []ParseResult
	}

	tests := []TestCase{
		{
			testName:             "data: []byte{}, parse \"a\"",
			data:                 []byte{},
			finder:               NewByteFinder('a'),
			expectedParseResults: []ParseResult{},
		},
		{
			testName: "data: []byte(\"a\"), parse \"a\"",
			data:     []byte("a"),
			finder:   NewByteFinder('a'),
			expectedParseResults: []ParseResult{
				{Parsed: []byte("a"), Remaining: []byte{}},
			},
		},
		{
			testName: "data: []byte(\"abc\"), parse \"a\"",
			data:     []byte("abc"),
			finder:   NewByteFinder('a'),
			expectedParseResults: []ParseResult{
				{Parsed: []byte("a"), Remaining: []byte("bc")},
			},
		},
		{
			testName: "data: []byte(\"aa\"), parse \"*aa\"",
			data:     []byte("aa"),
			finder: NewConcatenationFinder([]Finder{
				NewOptionalSequenceFinder(NewByteFinder('a')),
				NewByteFinder('a'),
			}),
			expectedParseResults: []ParseResult{
				{Parsed: []byte("a"), Remaining: []byte("a")},
				{Parsed: []byte("aa"), Remaining: []byte{}},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualParseResults := Parse(testCase.data, testCase.finder)
			if len(testCase.expectedParseResults) != len(actualParseResults) {
				t.Errorf("%v: expected: %v, actual: %v",
					testCase.testName,
					testCase.expectedParseResults,
					actualParseResults,
				)
			}
			for i, expectedParseResult := range testCase.expectedParseResults {
				sliceEquals(
					testCase.testName,
					t,
					expectedParseResult.Parsed,
					actualParseResults[i].Parsed,
				)
				sliceEquals(
					testCase.testName,
					t,
					expectedParseResult.Remaining,
					actualParseResults[i].Remaining,
				)
			}
		})
	}
}

func TestByteFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"a\"",
			data:         []byte{},
			finder:       NewByteFinder('a'),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"a\"",
			data:         []byte("a"),
			finder:       NewByteFinder('a'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"b\"), find \"a\"",
			data:         []byte("b"),
			finder:       NewByteFinder('a'),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"a\"",
			data:         []byte("ab"),
			finder:       NewByteFinder('a'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"b\"",
			data:         []byte("ab"),
			finder:       NewByteFinder('b'),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestBytesFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"ab\"",
			data:         []byte{},
			finder:       NewBytesFinder([]byte("ab")),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"ab\"",
			data:         []byte("a"),
			finder:       NewBytesFinder([]byte("ab")),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"ab\"",
			data:         []byte("ab"),
			finder:       NewBytesFinder([]byte("ab")),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"abc\"), find \"ab\"",
			data:         []byte("abc"),
			finder:       NewBytesFinder([]byte("ab")),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"abc\"), find \"bc\"",
			data:         []byte("abc"),
			finder:       NewBytesFinder([]byte("bc")),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestCrLfFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}",
			data:         []byte{},
			finder:       NewCrLfFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\")",
			data:         []byte("a"),
			finder:       NewCrLfFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"\\r\")",
			data:         []byte("\r"),
			finder:       NewCrLfFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"\\r\\n\")",
			data:         []byte("\r\n"),
			finder:       NewCrLfFinder(),
			expectedEnds: []int{2},
		},
	}
	execTest(tests, t)
}

func TestConcatenationFinder(t *testing.T) {
	tests := []TestCase{
		//
		// Concatenation: ALPHA ALPHA
		//
		{
			testName: "data: []byte{}, find ALPHA ALPHA",
			data:     []byte{},
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a\"), find ALPHA ALPHA",
			data:     []byte{'a'},
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"ab\"), find ALPHA ALPHA",
			data:     []byte("ab"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder()}),
			expectedEnds: []int{2},
		},
		{
			testName: "data: []byte(\"1\"), find ALPHA ALPHA",
			data:     []byte("1"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"12\"), find ALPHA ALPHA",
			data:     []byte("12"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a1\"), find ALPHA ALPHA",
			data:     []byte("a1"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"1a\"), find ALPHA ALPHA",
			data:     []byte("1a"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder()}),
			expectedEnds: []int{},
		},
		//
		// Concatenation: ALPHA DIGIT
		//
		{
			testName: "data: []byte{}, find ALPHA DIGIT",
			data:     []byte{},
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewDigitFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a\"), find ALPHA DIGIT",
			data:     []byte("a"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewDigitFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"ab\"), find ALPHA DIGIT",
			data:     []byte("ab"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewDigitFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"1\"), find ALPHA DIGIT",
			data:     []byte("1"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewDigitFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"12\"), find ALPHA DIGIT",
			data:     []byte("12"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewDigitFinder()}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a1\"), find ALPHA DIGIT",
			data:     []byte("a1"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewDigitFinder()}),
			expectedEnds: []int{2},
		},
		{
			testName: "data: []byte(\"1a\")",
			data:     []byte("1a"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewDigitFinder()}),
			expectedEnds: []int{},
		},
		//
		// Concatenation *ALPHA ALPHA
		//
		{
			testName: "data: []byte(\"aa\"), find *ALPHA ALPHA",
			data:     []byte("aa"),
			finder: NewConcatenationFinder([]Finder{
				NewVariableRepetitionFinder(NewAlphaFinder()),
				NewAlphaFinder(),
			}),
			expectedEnds: []int{1, 2},
		},
	}
	execTest(tests, t)
}

func TestAlternativesFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName: "data: []byte{}, find a / b",
			data:     []byte{},
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a\"), find a / b",
			data:     []byte("a"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"ab\"), find a / b",
			data:     []byte("ab"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"b\"), find a / b",
			data:     []byte("b"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"ba\"), find a / b",
			data:     []byte("ba"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"c\"), find a / b",
			data:     []byte("c"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"ca\")",
			data:     []byte("ca"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestValueRangeAlternativesFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, RangeStart: 'a', RangeEnd 'x'",
			data:         []byte{},
			finder:       NewValueRangeAlternativesFinder('a', 'x'),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), RangeStart: 'a', RangeEnd 'x'",
			data:         []byte("a"),
			finder:       NewValueRangeAlternativesFinder('a', 'x'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"x\"), RangeStart: 'a', RangeEnd 'x'",
			data:         []byte{'x'},
			finder:       NewValueRangeAlternativesFinder('a', 'x'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"{\"), RangeStart: 'a', RangeEnd 'x'",
			data:         []byte{'{'},
			finder:       NewValueRangeAlternativesFinder('a', 'x'),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestVariableRepetitionFinder(t *testing.T) {
	tests := []TestCase{
		//
		// NewVariableRepetitionMinMaxFinder
		//
		{
			testName:     "data: []byte{}, find \"0*1a\"",
			data:         []byte{},
			finder:       NewVariableRepetitionMinMaxFinder(0, 1, NewByteFinder('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte{}, find \"1*2a\"",
			data:         []byte{},
			finder:       NewVariableRepetitionMinMaxFinder(1, 2, NewByteFinder('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"1*2a\"",
			data:         []byte("a"),
			finder:       NewVariableRepetitionMinMaxFinder(1, 2, NewByteFinder('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a\"), find \"2*3a\"",
			data:         []byte("a"),
			finder:       NewVariableRepetitionMinMaxFinder(2, 3, NewByteFinder('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"2*3a\"",
			data:         []byte("aa"),
			finder:       NewVariableRepetitionMinMaxFinder(2, 3, NewByteFinder('a')),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"aaaa\"), find \"2*3a\"",
			data:         []byte("aaaa"),
			finder:       NewVariableRepetitionMinMaxFinder(2, 3, NewByteFinder('a')),
			expectedEnds: []int{2, 3},
		},
		//
		// NewVariableRepetitionMinFinder
		//
		{
			testName:     "data: []byte{}, find \"0*a\"",
			data:         []byte{},
			finder:       NewVariableRepetitionMinFinder(0, NewByteFinder('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte{}, find \"1*a\"",
			data:         []byte{},
			finder:       NewVariableRepetitionMinFinder(1, NewByteFinder('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"1*a\"",
			data:         []byte("a"),
			finder:       NewVariableRepetitionMinFinder(1, NewByteFinder('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a\"), find \"2*a\"",
			data:         []byte("a"),
			finder:       NewVariableRepetitionMinFinder(2, NewByteFinder('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"2*a\"",
			data:         []byte("aa"),
			finder:       NewVariableRepetitionMinFinder(2, NewByteFinder('a')),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"2*a\"",
			data:         []byte("aaa"),
			finder:       NewVariableRepetitionMinFinder(2, NewByteFinder('a')),
			expectedEnds: []int{2, 3},
		},
		//
		// NewVariableRepetitionMaxFinder
		//
		{
			testName:     "data: []byte{}, find \"*1a\"",
			data:         []byte{},
			finder:       NewVariableRepetitionMaxFinder(1, NewByteFinder('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte(\"a\"), find \"*2a\"",
			data:         []byte("a"),
			finder:       NewVariableRepetitionMaxFinder(2, NewByteFinder('a')),
			expectedEnds: []int{0, 1},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"*2a\"",
			data:         []byte("aa"),
			finder:       NewVariableRepetitionMaxFinder(2, NewByteFinder('a')),
			expectedEnds: []int{0, 1, 2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"*3a\"",
			data:         []byte("aaa"),
			finder:       NewVariableRepetitionMaxFinder(3, NewByteFinder('a')),
			expectedEnds: []int{0, 1, 2, 3},
		},
		{
			testName:     "data: []byte(\"aaaa\"), find \"*3a\"",
			data:         []byte("aaaa"),
			finder:       NewVariableRepetitionMaxFinder(3, NewByteFinder('a')),
			expectedEnds: []int{0, 1, 2, 3},
		},
		//
		// NewVariableRepetitionFinder
		//
		{
			testName:     "data: []byte{}, find \"*a\"",
			data:         []byte{},
			finder:       NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte(\"a\"), find \"*a\"",
			data:         []byte("a"),
			finder:       NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedEnds: []int{0, 1},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"*a\"",
			data:         []byte("aa"),
			finder:       NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedEnds: []int{0, 1, 2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"*a\"",
			data:         []byte("aaa"),
			finder:       NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedEnds: []int{0, 1, 2, 3},
		},
	}
	execTest(tests, t)
}

func TestSpecificRepetitionFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"0a\"",
			data:         []byte{},
			finder:       NewSpecificRepetitionFinder(0, NewByteFinder('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte{}, find \"1a\"",
			data:         []byte{},
			finder:       NewSpecificRepetitionFinder(1, NewByteFinder('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"1a\"",
			data:         []byte("a"),
			finder:       NewSpecificRepetitionFinder(1, NewByteFinder('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"1a\"",
			data:         []byte("ab"),
			finder:       NewSpecificRepetitionFinder(1, NewByteFinder('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a\"), find \"2a\"",
			data:         []byte("a"),
			finder:       NewSpecificRepetitionFinder(2, NewByteFinder('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"2a\"",
			data:         []byte("aa"),
			finder:       NewSpecificRepetitionFinder(2, NewByteFinder('a')),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"2a\"",
			data:         []byte("aaa"),
			finder:       NewSpecificRepetitionFinder(2, NewByteFinder('a')),
			expectedEnds: []int{2},
		},
	}
	execTest(tests, t)
}

func TestOptionalSequenceFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"[ a ]\"",
			data:         []byte{},
			finder:       NewOptionalSequenceFinder(NewByteFinder('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte(\"a\"), find \"[ a ]\"",
			data:         []byte("a"),
			finder:       NewOptionalSequenceFinder(NewByteFinder('a')),
			expectedEnds: []int{0, 1},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"[ a ]\"",
			data:         []byte("ab"),
			finder:       NewOptionalSequenceFinder(NewByteFinder('a')),
			expectedEnds: []int{0, 1},
		},
	}
	execTest(tests, t)
}

func TestAlphaFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find ALPHA",
			data:         []byte{},
			finder:       NewAlphaFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find ALPHA",
			data:         []byte("a"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"ab\"), find ALPHA",
			data:         []byte("ab"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1\"), find ALPHA",
			data:         []byte("1"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"12\"), find ALPHA",
			data:         []byte("12"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"A\"), find ALPHA",
			data:         []byte("A"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"AB\"), find ALPHA",
			data:         []byte("AB"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1A\"), find ALPHA",
			data:         []byte("1A"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a1\"), find ALPHA",
			data:         []byte("a1"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1A\"), find ALPHA",
			data:         []byte("1A"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"A1\"), find ALPHA",
			data:         []byte("A1"),
			finder:       NewAlphaFinder(),
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestDigitFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find DIGIT",
			data:         []byte{},
			finder:       NewDigitFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find DIGIT",
			data:         []byte("a"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find DIGIT",
			data:         []byte("ab"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1\"), find DIGIT",
			data:         []byte("1"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"12\"), find DIGIT",
			data:         []byte("12"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A\"), find DIGIT",
			data:         []byte("A"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"AB\"), find DIGIT",
			data:         []byte("AB"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1a\"), find DIGIT",
			data:         []byte("1a"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a1\"), find DIGIT",
			data:         []byte("a1"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1A\"), find DIGIT",
			data:         []byte("1A"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A1\"), find DIGIT",
			data:         []byte("A1"),
			finder:       NewDigitFinder(),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestDQuoteFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find DQUOTE",
			data:         []byte{},
			finder:       NewDQuoteFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find DQUOTE",
			data:         []byte("a"),
			finder:       NewDQuoteFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"'\"), find DQUOTE",
			data:         []byte("'"),
			finder:       NewDQuoteFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"\"\"), find DQUOTE",
			data:         []byte("\""),
			finder:       NewDQuoteFinder(),
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestHexDigFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find HEXDIG",
			data:         []byte{},
			finder:       NewHexDigFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find HEXDIG",
			data:         []byte("a"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find HEXDIG",
			data:         []byte("ab"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1\"), find HEXDIG",
			data:         []byte("1"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"12\"), find HEXDIG",
			data:         []byte("12"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A\"), find HEXDIG",
			data:         []byte("A"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"AB\")",
			data:         []byte("AB"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1a\"), find HEXDIG",
			data:         []byte("1a"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a1\"), find HEXDIG",
			data:         []byte("a1"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1A\")",
			data:         []byte("1A"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A1\")",
			data:         []byte("A1"),
			finder:       NewHexDigFinder(),
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestHTabFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find HTAB",
			data:         []byte{},
			finder:       NewHTabFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find HTAB",
			data:         []byte("a"),
			finder:       NewHTabFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x20}, find HTAB",
			data:         []byte{0x20},
			finder:       NewHTabFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x09}, find HTAB",
			data:         []byte{0x09},
			finder:       NewHTabFinder(),
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestOctetFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find Octet",
			data:         []byte{},
			finder:       NewOctetFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x00}, find Octet",
			data:         []byte{0x00},
			finder:       NewOctetFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0xff}, find Octet",
			data:         []byte{0xff},
			finder:       NewOctetFinder(),
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestFindSp(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find SP",
			data:         []byte{},
			finder:       NewSpFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find SP",
			data:         []byte("a"),
			finder:       NewSpFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x20}, find SP",
			data:         []byte{0x20},
			finder:       NewSpFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0x09}, find SP",
			data:         []byte{0x09},
			finder:       NewSpFinder(),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestFindVChar(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find VCHAR",
			data:         []byte{},
			finder:       NewVCharFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x20}, find VCHAR",
			data:         []byte{0x20},
			finder:       NewVCharFinder(),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x21}, find VCHAR",
			data:         []byte{0x21},
			finder:       NewVCharFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0x7e}, find VCHAR",
			data:         []byte{0x7e},
			finder:       NewVCharFinder(),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0x7f}, find VCHAR",
			data:         []byte{0x7f},
			finder:       NewVCharFinder(),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}
