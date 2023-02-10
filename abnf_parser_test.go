package abnfp

import "testing"

type TestCase struct {
	testName      string
	data          []byte
	finder        Finder
	expectedFound bool
	expectedEnd   int
}

func equals[C comparable](testName string, t *testing.T, expected C, actual C) {
	if actual != expected {
		t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
	}
}

func sliceEquals[C comparable](testName string, t *testing.T, expected []C, actual []C) {
	if len(expected) != len(actual) {
		t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
		return
	}
	for i, e := range expected {
		if e != actual[i] {
			t.Errorf("%v: expected: %v, actual: %v", testName, expected, actual)
			return
		}
	}
}

func execFinderTest(tests []TestCase, t *testing.T) {
	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualFound, actualEnd := testCase.finder.Find(testCase.data)
			equals(testCase.testName, t, testCase.expectedFound, actualFound)
			equals(testCase.testName, t, testCase.expectedEnd, actualEnd)
		})
	}
}

func TestParse(t *testing.T) {
	type TestCase struct {
		testName          string
		data              []byte
		finder            Finder
		expectedParsed    []byte
		expectedRemaining []byte
	}

	tests := []TestCase{
		{
			testName:          "data: []byte{}, parse \"a\"",
			data:              []byte{},
			finder:            NewByteFinder('a'),
			expectedParsed:    []byte(""),
			expectedRemaining: []byte(""),
		},
		{
			testName:          "data: []byte(\"a\"), parse \"a\"",
			data:              []byte("a"),
			finder:            NewByteFinder('a'),
			expectedParsed:    []byte("a"),
			expectedRemaining: []byte(""),
		},
		{
			testName:          "data: []byte(\"abc\"), parse \"a\"",
			data:              []byte("abc"),
			finder:            NewByteFinder('a'),
			expectedParsed:    []byte("a"),
			expectedRemaining: []byte("bc"),
		},
		{
			testName:          "data: []byte(\"aa\"), parse \"*a\"",
			data:              []byte("aa"),
			finder:            NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedParsed:    []byte("aa"),
			expectedRemaining: []byte(""),
		},
		{
			testName: "data: []byte(\"aa\"), parse \"*aa\"",
			data:     []byte("aa"),
			finder: NewConcatenationFinder([]Finder{
				NewVariableRepetitionFinder(NewByteFinder('a')),
				NewByteFinder('a'),
			}),
			expectedParsed:    []byte("aa"),
			expectedRemaining: []byte(""),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			parsed, remaining := Parse(testCase.data, testCase.finder)
			sliceEquals(
				testCase.testName,
				t,
				testCase.expectedParsed,
				parsed,
			)
			sliceEquals(
				testCase.testName,
				t,
				testCase.expectedRemaining,
				remaining,
			)
		})
	}
}

func TestByteFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find \"a\"",
			data:          []byte{},
			finder:        NewByteFinder('a'),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"a\"",
			data:          []byte("a"),
			finder:        NewByteFinder('a'),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"b\"), find \"a\"",
			data:          []byte("b"),
			finder:        NewByteFinder('a'),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"ab\"), find \"a\"",
			data:          []byte("ab"),
			finder:        NewByteFinder('a'),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"ab\"), find \"b\"",
			data:          []byte("ab"),
			finder:        NewByteFinder('b'),
			expectedFound: false,
			expectedEnd:   0,
		},
	}
	execFinderTest(tests, t)
}

func TestBytesFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find \"ab\"",
			data:          []byte{},
			finder:        NewBytesFinder([]byte("ab")),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"ab\"",
			data:          []byte("a"),
			finder:        NewBytesFinder([]byte("ab")),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"ab\"), find \"ab\"",
			data:          []byte("ab"),
			finder:        NewBytesFinder([]byte("ab")),
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "data: []byte(\"abc\"), find \"ab\"",
			data:          []byte("abc"),
			finder:        NewBytesFinder([]byte("ab")),
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "data: []byte(\"abc\"), find \"bc\"",
			data:          []byte("abc"),
			finder:        NewBytesFinder([]byte("bc")),
			expectedFound: false,
			expectedEnd:   0,
		},
	}
	execFinderTest(tests, t)
}

func TestCrLfFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find CRLF",
			data:          []byte{},
			finder:        NewCrLfFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find CRLF",
			data:          []byte("a"),
			finder:        NewCrLfFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"\\r\"), find CRLF",
			data:          []byte("\r"),
			finder:        NewCrLfFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"\\r\\n\"), find CRLF",
			data:          []byte("\r\n"),
			finder:        NewCrLfFinder(),
			expectedFound: true,
			expectedEnd:   2,
		},
	}
	execFinderTest(tests, t)
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
				NewAlphaFinder(),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName: "data: []byte(\"a\"), find ALPHA ALPHA",
			data:     []byte("a"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder(),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName: "data: []byte(\"ab\"), find ALPHA ALPHA",
			data:     []byte("ab"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder(),
			}),
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName: "data: []byte(\"1\"), find ALPHA ALPHA",
			data:     []byte("1"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder(),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName: "data: []byte(\"12\"), find ALPHA ALPHA",
			data:     []byte("12"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder(),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName: "data: []byte(\"a1\"), find ALPHA ALPHA",
			data:     []byte("a1"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder(),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName: "data: []byte(\"1a\"), find ALPHA ALPHA",
			data:     []byte("1a"),
			finder: NewConcatenationFinder([]Finder{
				NewAlphaFinder(),
				NewAlphaFinder(),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		//
		// Concatenation *ALPHA ALPHA
		//
		// NOTE
		// In this test case, ConcatenationFinder calls its Recalculate() once.
		//
		{
			testName: "data: []byte(\"a\"), find *ALPHA ALPHA",
			data:     []byte("a"),
			finder: NewConcatenationFinder([]Finder{
				NewVariableRepetitionFinder(NewAlphaFinder()),
				NewAlphaFinder(),
			}),
			expectedFound: true,
			expectedEnd:   1,
		},
		//
		// Concatenation *ALPHA ALPHA
		//
		// NOTE
		// In this test case, ConcatenationFinder calls its Recalculate() once.
		//
		{
			testName: "data: []byte(\"aa\"), find *ALPHA ALPHA",
			data:     []byte("aa"),
			finder: NewConcatenationFinder([]Finder{
				NewVariableRepetitionFinder(NewAlphaFinder()),
				NewAlphaFinder(),
			}),
			expectedFound: true,
			expectedEnd:   2,
		},
		//
		// Concatenation *ALPHA ALPHA
		//
		// NOTE
		// In this test case, ConcatenationFinder calls its Recalculate() twice.
		//
		{
			testName: "data: []byte(\"aa\"), find *ALPHA ALPHA ALPHA",
			data:     []byte("aa"),
			finder: NewConcatenationFinder([]Finder{
				NewVariableRepetitionFinder(NewAlphaFinder()),
				NewAlphaFinder(),
				NewAlphaFinder(),
			}),
			expectedFound: true,
			expectedEnd:   2,
		},
	}
	execFinderTest(tests, t)
}

func TestAlternativesFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName: "data: []byte{}, find a | b",
			data:     []byte{},
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName: "data: []byte(\"a\"), find a | b",
			data:     []byte("a"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName: "data: []byte(\"ab\"), find a | b",
			data:     []byte("ab"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName: "data: []byte(\"b\"), find a | b",
			data:     []byte("b"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName: "data: []byte(\"ba\"), find a | b",
			data:     []byte("ba"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName: "data: []byte(\"c\"), find a | b",
			data:     []byte("c"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName: "data: []byte(\"ca\"), find a | b",
			data:     []byte("ca"),
			finder: NewAlternativesFinder([]Finder{
				NewByteFinder('a'),
				NewByteFinder('b'),
			}),
			expectedFound: false,
			expectedEnd:   0,
		},
	}
	execFinderTest(tests, t)
}

func TestValueRangeAlternativesFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, RangeStart: 'a', RangeEnd 'x'",
			data:          []byte{},
			finder:        NewValueRangeAlternativesFinder('a', 'x'),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), RangeStart: 'a', RangeEnd 'x'",
			data:          []byte("a"),
			finder:        NewValueRangeAlternativesFinder('a', 'x'),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"x\"), RangeStart: 'a', RangeEnd 'x'",
			data:          []byte{'x'},
			finder:        NewValueRangeAlternativesFinder('a', 'x'),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"{\"), RangeStart: 'a', RangeEnd 'x'",
			data:          []byte{'{'},
			finder:        NewValueRangeAlternativesFinder('a', 'x'),
			expectedFound: false,
			expectedEnd:   0,
		},
	}
	execFinderTest(tests, t)
}

func TestVariableRepetitionMinMaxFinder(t *testing.T) {
	tests := []TestCase{
		//
		// NewVariableRepetitionMinMaxFinder
		//
		{
			testName:      "data: []byte{}, find \"0*1a\"",
			data:          []byte{},
			finder:        NewVariableRepetitionMinMaxFinder(0, 1, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{}, find \"1*2a\"",
			data:          []byte{},
			finder:        NewVariableRepetitionMinMaxFinder(1, 2, NewByteFinder('a')),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"1*2a\"",
			data:          []byte("a"),
			finder:        NewVariableRepetitionMinMaxFinder(1, 2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"a\"), find \"2*3a\"",
			data:          []byte("a"),
			finder:        NewVariableRepetitionMinMaxFinder(2, 3, NewByteFinder('a')),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"aa\"), find \"2*3a\"",
			data:          []byte("aa"),
			finder:        NewVariableRepetitionMinMaxFinder(2, 3, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "data: []byte(\"aaaa\"), find \"2*3a\"",
			data:          []byte("aaaa"),
			finder:        NewVariableRepetitionMinMaxFinder(2, 3, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   3,
		},
	}
	execFinderTest(tests, t)
}

func TestVariableRepetitionMinFinder(t *testing.T) {
	tests := []TestCase{
		//
		// NewVariableRepetitionMinFinder
		//
		{
			testName:      "data: []byte{}, find \"0*a\"",
			data:          []byte{},
			finder:        NewVariableRepetitionMinFinder(0, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{}, find \"1*a\"",
			data:          []byte{},
			finder:        NewVariableRepetitionMinFinder(1, NewByteFinder('a')),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"1*a\"",
			data:          []byte("a"),
			finder:        NewVariableRepetitionMinFinder(1, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"a\"), find \"2*a\"",
			data:          []byte("a"),
			finder:        NewVariableRepetitionMinFinder(2, NewByteFinder('a')),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"aa\"), find \"2*a\"",
			data:          []byte("aa"),
			finder:        NewVariableRepetitionMinFinder(2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "data: []byte(\"aaa\"), find \"2*a\"",
			data:          []byte("aaa"),
			finder:        NewVariableRepetitionMinFinder(2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   3,
		},
	}
	execFinderTest(tests, t)
}

func TestVariableRepetitionMaxFinder(t *testing.T) {
	tests := []TestCase{
		//
		// NewVariableRepetitionMaxFinder
		//
		{
			testName:      "data: []byte{}, find \"*1a\"",
			data:          []byte{},
			finder:        NewVariableRepetitionMaxFinder(1, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"*2a\"",
			data:          []byte("a"),
			finder:        NewVariableRepetitionMaxFinder(2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"aa\"), find \"*2a\"",
			data:          []byte("aa"),
			finder:        NewVariableRepetitionMaxFinder(2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "data: []byte(\"aaa\"), find \"*2a\"",
			data:          []byte("aaa"),
			finder:        NewVariableRepetitionMaxFinder(2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   2,
		},
	}
	execFinderTest(tests, t)
}

func TestVariableRepetitionFinder(t *testing.T) {
	tests := []TestCase{
		//
		// NewVariableRepetitionFinder
		//
		{
			testName:      "data: []byte{}, find \"*a\"",
			data:          []byte{},
			finder:        NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"*a\"",
			data:          []byte("a"),
			finder:        NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"aa\"), find \"*a\"",
			data:          []byte("aa"),
			finder:        NewVariableRepetitionFinder(NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   2,
		},
		//
		// combination with Concatenation
		//
		{
			testName: "data: []byte(\"a1b2c3\"), find \"*(ALPHA | DIGIT)\"",
			data:     []byte("a1b2c3"),
			finder: NewVariableRepetitionFinder(
				NewAlternativesFinder([]Finder{
					NewAlphaFinder(),
					NewDigitFinder(),
				}),
			),
			expectedFound: true,
			expectedEnd:   6,
		},
		//
		// combination with Concatenation and VariableRepetition
		//
		{
			testName: "data: []byte(\"a1bc3\"), find \"*(ALPHA *DIGIT)\"",
			data:     []byte("a1bc3"),
			finder: NewVariableRepetitionFinder(
				NewConcatenationFinder([]Finder{
					NewAlphaFinder(),
					NewVariableRepetitionFinder(
						NewDigitFinder(),
					),
				}),
			),
			expectedFound: true,
			expectedEnd:   5,
		},
	}
	execFinderTest(tests, t)
}

func TestSpecificRepetitionFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find \"0a\"",
			data:          []byte{},
			finder:        NewSpecificRepetitionFinder(0, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{}, find \"1a\"",
			data:          []byte{},
			finder:        NewSpecificRepetitionFinder(1, NewByteFinder('a')),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"1a\"",
			data:          []byte("a"),
			finder:        NewSpecificRepetitionFinder(1, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"ab\"), find \"1a\"",
			data:          []byte("ab"),
			finder:        NewSpecificRepetitionFinder(1, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"a\"), find \"2a\"",
			data:          []byte("a"),
			finder:        NewSpecificRepetitionFinder(2, NewByteFinder('a')),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"aa\"), find \"2a\"",
			data:          []byte("aa"),
			finder:        NewSpecificRepetitionFinder(2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   2,
		},
		{
			testName:      "data: []byte(\"aaa\"), find \"2a\"",
			data:          []byte("aaa"),
			finder:        NewSpecificRepetitionFinder(2, NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   2,
		},
	}
	execFinderTest(tests, t)
}

func TestOptionalSequenceFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find \"[ a ]\"",
			data:          []byte{},
			finder:        NewOptionalSequenceFinder(NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find \"[ a ]\"",
			data:          []byte("a"),
			finder:        NewOptionalSequenceFinder(NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"ab\"), find \"[ a ]\"",
			data:          []byte("ab"),
			finder:        NewOptionalSequenceFinder(NewByteFinder('a')),
			expectedFound: true,
			expectedEnd:   1,
		},
	}
	execFinderTest(tests, t)
}

func TestAlphaFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find ALPHA",
			data:          []byte{},
			finder:        NewAlphaFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find ALPHA",
			data:          []byte("a"),
			finder:        NewAlphaFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"ab\"), find ALPHA",
			data:          []byte("ab"),
			finder:        NewAlphaFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"1\"), find ALPHA",
			data:          []byte("1"),
			finder:        NewAlphaFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"12\"), find ALPHA",
			data:          []byte("12"),
			finder:        NewAlphaFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"A\"), find ALPHA",
			data:          []byte("A"),
			finder:        NewAlphaFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"AB\"), find ALPHA",
			data:          []byte("AB"),
			finder:        NewAlphaFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"1A\"), find ALPHA",
			data:          []byte("1A"),
			finder:        NewAlphaFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a1\"), find ALPHA",
			data:          []byte("a1"),
			finder:        NewAlphaFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"1A\"), find ALPHA",
			data:          []byte("1A"),
			finder:        NewAlphaFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"A1\"), find ALPHA",
			data:          []byte("A1"),
			finder:        NewAlphaFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
	}
	execFinderTest(tests, t)
}

func TestDigitFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find DIGIT",
			data:          []byte{},
			finder:        NewDigitFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find DIGIT",
			data:          []byte("a"),
			finder:        NewDigitFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"ab\"), find DIGIT",
			data:          []byte("ab"),
			finder:        NewDigitFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"1\"), find DIGIT",
			data:          []byte("1"),
			finder:        NewDigitFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"12\"), find DIGIT",
			data:          []byte("12"),
			finder:        NewDigitFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"A\"), find DIGIT",
			data:          []byte("A"),
			finder:        NewDigitFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"AB\"), find DIGIT",
			data:          []byte("AB"),
			finder:        NewDigitFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"1a\"), find DIGIT",
			data:          []byte("1a"),
			finder:        NewDigitFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"a1\"), find DIGIT",
			data:          []byte("a1"),
			finder:        NewDigitFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"1A\"), find DIGIT",
			data:          []byte("1A"),
			finder:        NewDigitFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"A1\"), find DIGIT",
			data:          []byte("A1"),
			finder:        NewDigitFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
	}
	execFinderTest(tests, t)
}

func TestDQuoteFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find DQUOTE",
			data:          []byte{},
			finder:        NewDQuoteFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find DQUOTE",
			data:          []byte("a"),
			finder:        NewDQuoteFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"'\"), find DQUOTE",
			data:          []byte("'"),
			finder:        NewDQuoteFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"\"\"), find DQUOTE",
			data:          []byte("\""),
			finder:        NewDQuoteFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
	}
	execFinderTest(tests, t)
}

func TestHexDigFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find HEXDIG",
			data:          []byte{},
			finder:        NewHexDigFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find HEXDIG",
			data:          []byte("a"),
			finder:        NewHexDigFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"ab\"), find HEXDIG",
			data:          []byte("ab"),
			finder:        NewHexDigFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"1\"), find HEXDIG",
			data:          []byte("1"),
			finder:        NewHexDigFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"12\"), find HEXDIG",
			data:          []byte("12"),
			finder:        NewHexDigFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"A\"), find HEXDIG",
			data:          []byte("A"),
			finder:        NewHexDigFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"AB\")",
			data:          []byte("AB"),
			finder:        NewHexDigFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"1a\"), find HEXDIG",
			data:          []byte("1a"),
			finder:        NewHexDigFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"a1\"), find HEXDIG",
			data:          []byte("a1"),
			finder:        NewHexDigFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"1A\")",
			data:          []byte("1A"),
			finder:        NewHexDigFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte(\"A1\")",
			data:          []byte("A1"),
			finder:        NewHexDigFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
	}
	execFinderTest(tests, t)
}

func TestHTabFinder(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find HTAB",
			data:          []byte{},
			finder:        NewHTabFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find HTAB",
			data:          []byte("a"),
			finder:        NewHTabFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x20}, find HTAB",
			data:          []byte{0x20},
			finder:        NewHTabFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x09}, find HTAB",
			data:          []byte{0x09},
			finder:        NewHTabFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
	}
	execFinderTest(tests, t)
}

func TestFindOctet(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find Octet",
			data:          []byte{},
			finder:        NewOctetFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x00}, find Octet",
			data:          []byte{0x00},
			finder:        NewOctetFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0xff}, find Octet",
			data:          []byte{0xff},
			finder:        NewOctetFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
	}
	execFinderTest(tests, t)
}

func TestFindSp(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find SP",
			data:          []byte{},
			finder:        NewSpFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte(\"a\"), find SP",
			data:          []byte("a"),
			finder:        NewSpFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x20}, find SP",
			data:          []byte{0x20},
			finder:        NewSpFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0x09}, find SP",
			data:          []byte{0x09},
			finder:        NewSpFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
	}
	execFinderTest(tests, t)
}

func TestFindVChar(t *testing.T) {
	tests := []TestCase{
		{
			testName:      "data: []byte{}, find VCHAR",
			data:          []byte{},
			finder:        NewVCharFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x20}, find VCHAR",
			data:          []byte{0x20},
			finder:        NewVCharFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
		{
			testName:      "data: []byte{0x21}, find VCHAR",
			data:          []byte{0x21},
			finder:        NewVCharFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0x7e}, find VCHAR",
			data:          []byte{0x7e},
			finder:        NewVCharFinder(),
			expectedFound: true,
			expectedEnd:   1,
		},
		{
			testName:      "data: []byte{0x7f}, find VCHAR",
			data:          []byte{0x7f},
			finder:        NewVCharFinder(),
			expectedFound: false,
			expectedEnd:   0,
		},
	}
	execFinderTest(tests, t)
}
