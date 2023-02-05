package abnfp

import "testing"

type TestCase struct {
	testName     string
	data         []byte
	findFunc     FindFunc
	expectedEnds []int
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

func execTest(tests []TestCase, t *testing.T) {
	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualEnds := testCase.findFunc(testCase.data)
			sliceEquals(testCase.testName, t, testCase.expectedEnds, actualEnds)
		})
	}
}

func TestParse(t *testing.T) {
	type TestCase struct {
		testName             string
		data                 []byte
		findFunc             FindFunc
		parseMode            ParseMode
		expectedParseResults []ParseResult
	}

	tests := []TestCase{
		{
			testName:             "data: []byte{}, parse \"a\"",
			data:                 []byte{},
			findFunc:             NewFindByte('a'),
			parseMode:            PARSE_LONGEST,
			expectedParseResults: []ParseResult{},
		},
		{
			testName:  "data: []byte(\"a\"), parse \"a\"",
			data:      []byte("a"),
			findFunc:  NewFindByte('a'),
			parseMode: PARSE_LONGEST,
			expectedParseResults: []ParseResult{
				{Parsed: []byte("a"), Remaining: []byte{}},
			},
		},
		{
			testName:  "data: []byte(\"abc\"), parse \"a\"",
			data:      []byte("abc"),
			findFunc:  NewFindByte('a'),
			parseMode: PARSE_LONGEST,
			expectedParseResults: []ParseResult{
				{Parsed: []byte("a"), Remaining: []byte("bc")},
			},
		},
		{
			testName:  "data: []byte(\"aa\"), parse \"*a\", mode longest",
			data:      []byte("aa"),
			findFunc:  NewFindVariableRepetition(NewFindByte('a')),
			parseMode: PARSE_LONGEST,
			expectedParseResults: []ParseResult{
				{Parsed: []byte("aa"), Remaining: []byte{}},
			},
		},
		{
			testName:  "data: []byte(\"aa\"), parse \"*a\", mode shortest",
			data:      []byte("aa"),
			findFunc:  NewFindVariableRepetition(NewFindByte('a')),
			parseMode: PARSE_SHORTEST,
			expectedParseResults: []ParseResult{
				{Parsed: []byte{}, Remaining: []byte("aa")},
			},
		},
		{
			testName:  "data: []byte(\"aa\"), parse \"*a\", mode all",
			data:      []byte("aa"),
			findFunc:  NewFindVariableRepetition(NewFindByte('a')),
			parseMode: PARSE_ALL,
			expectedParseResults: []ParseResult{
				{Parsed: []byte{}, Remaining: []byte("aa")},
				{Parsed: []byte("a"), Remaining: []byte("a")},
				{Parsed: []byte("aa"), Remaining: []byte{}},
			},
		},

		{
			testName: "data: []byte(\"aa\"), parse \"*aa\", mode longest",
			data:     []byte("aa"),
			findFunc: NewFindConcatenation([]FindFunc{
				NewFindVariableRepetition(NewFindByte('a')),
				NewFindByte('a'),
			}),
			parseMode: PARSE_LONGEST,
			expectedParseResults: []ParseResult{
				{Parsed: []byte("aa"), Remaining: []byte{}},
			},
		},
		{
			testName: "data: []byte(\"aa\"), parse \"*aa\", mode shortest",
			data:     []byte("aa"),
			findFunc: NewFindConcatenation([]FindFunc{
				NewFindVariableRepetition(NewFindByte('a')),
				NewFindByte('a'),
			}),
			parseMode: PARSE_SHORTEST,
			expectedParseResults: []ParseResult{
				{Parsed: []byte("a"), Remaining: []byte("a")},
			},
		},
		{
			testName: "data: []byte(\"aa\"), parse \"*aa\", mode all",
			data:     []byte("aa"),
			findFunc: NewFindConcatenation([]FindFunc{
				NewFindVariableRepetition(NewFindByte('a')),
				NewFindByte('a'),
			}),
			parseMode: PARSE_ALL,
			expectedParseResults: []ParseResult{
				{Parsed: []byte("a"), Remaining: []byte("a")},
				{Parsed: []byte("aa"), Remaining: []byte{}},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			actualParseResults := Parse(testCase.data, testCase.findFunc, testCase.parseMode)
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

func TestFindByte(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"a\"",
			data:         []byte{},
			findFunc:     NewFindByte('a'),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"a\"",
			data:         []byte("a"),
			findFunc:     NewFindByte('a'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"b\"), find \"a\"",
			data:         []byte("b"),
			findFunc:     NewFindByte('a'),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"a\"",
			data:         []byte("ab"),
			findFunc:     NewFindByte('a'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"b\"",
			data:         []byte("ab"),
			findFunc:     NewFindByte('b'),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestFindBytes(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"ab\"",
			data:         []byte{},
			findFunc:     NewFindBytes([]byte("ab")),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"ab\"",
			data:         []byte("a"),
			findFunc:     NewFindBytes([]byte("ab")),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"ab\"",
			data:         []byte("ab"),
			findFunc:     NewFindBytes([]byte("ab")),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"abc\"), find \"ab\"",
			data:         []byte("abc"),
			findFunc:     NewFindBytes([]byte("ab")),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"abc\"), find \"bc\"",
			data:         []byte("abc"),
			findFunc:     NewFindBytes([]byte("bc")),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestFindCrLf(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}",
			data:         []byte{},
			findFunc:     FindCrLf,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\")",
			data:         []byte("a"),
			findFunc:     FindCrLf,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"\\r\")",
			data:         []byte("\r"),
			findFunc:     FindCrLf,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"\\r\\n\")",
			data:         []byte("\r\n"),
			findFunc:     FindCrLf,
			expectedEnds: []int{2},
		},
	}
	execTest(tests, t)
}

func TestFindConcatenation(t *testing.T) {
	tests := []TestCase{
		//
		// Concatenation: ALPHA ALPHA
		//
		{
			testName: "data: []byte{}, find ALPHA ALPHA",
			data:     []byte{},
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindAlpha,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a\"), find ALPHA ALPHA",
			data:     []byte{'a'},
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindAlpha,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"ab\"), find ALPHA ALPHA",
			data:     []byte("ab"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindAlpha,
			}),
			expectedEnds: []int{2},
		},
		{
			testName: "data: []byte(\"1\"), find ALPHA ALPHA",
			data:     []byte("1"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindAlpha,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"12\"), find ALPHA ALPHA",
			data:     []byte("12"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindAlpha,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a1\"), find ALPHA ALPHA",
			data:     []byte("a1"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindAlpha,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"1a\"), find ALPHA ALPHA",
			data:     []byte("1a"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindAlpha,
			}),
			expectedEnds: []int{},
		},
		//
		// Concatenation: ALPHA DIGIT
		//
		{
			testName: "data: []byte{}, find ALPHA DIGIT",
			data:     []byte{},
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a\"), find ALPHA DIGIT",
			data:     []byte("a"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"ab\"), find ALPHA DIGIT",
			data:     []byte("ab"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"1\"), find ALPHA DIGIT",
			data:     []byte("1"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"12\"), find ALPHA DIGIT",
			data:     []byte("12"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a1\"), find ALPHA DIGIT",
			data:     []byte("a1"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			}),
			expectedEnds: []int{2},
		},
		{
			testName: "data: []byte(\"1a\")",
			data:     []byte("1a"),
			findFunc: NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			}),
			expectedEnds: []int{},
		},
		//
		// Concatenation *ALPHA ALPHA
		//
		{
			testName: "data: []byte(\"aa\"), find *ALPHA ALPHA",
			data:     []byte("aa"),
			findFunc: NewFindConcatenation([]FindFunc{
				NewFindVariableRepetition(FindAlpha),
				FindAlpha,
			}),
			expectedEnds: []int{1, 2},
		},
	}
	execTest(tests, t)
}

func TestFindAlternatives(t *testing.T) {
	tests := []TestCase{
		{
			testName: "data: []byte{}, find a / b",
			data:     []byte{},
			findFunc: NewFindAlternatives([]FindFunc{
				NewFindByte('a'),
				NewFindByte('b'),
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"a\"), find a / b",
			data:     []byte("a"),
			findFunc: NewFindAlternatives([]FindFunc{
				NewFindByte('a'),
				NewFindByte('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"ab\"), find a / b",
			data:     []byte("ab"),
			findFunc: NewFindAlternatives([]FindFunc{
				NewFindByte('a'),
				NewFindByte('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"b\"), find a / b",
			data:     []byte("b"),
			findFunc: NewFindAlternatives([]FindFunc{
				NewFindByte('a'),
				NewFindByte('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"ba\"), find a / b",
			data:     []byte("ba"),
			findFunc: NewFindAlternatives([]FindFunc{
				NewFindByte('a'),
				NewFindByte('b'),
			}),
			expectedEnds: []int{1},
		},
		{
			testName: "data: []byte(\"c\"), find a / b",
			data:     []byte("c"),
			findFunc: NewFindAlternatives([]FindFunc{
				NewFindByte('a'),
				NewFindByte('b'),
			}),
			expectedEnds: []int{},
		},
		{
			testName: "data: []byte(\"ca\")",
			data:     []byte("ca"),
			findFunc: NewFindAlternatives([]FindFunc{
				NewFindByte('a'),
				NewFindByte('b'),
			}),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestFindValueRangeAlternatives(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, RangeStart: 'a', RangeEnd 'x'",
			data:         []byte{},
			findFunc:     NewFindValueRangeAlternatives('a', 'x'),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), RangeStart: 'a', RangeEnd 'x'",
			data:         []byte("a"),
			findFunc:     NewFindValueRangeAlternatives('a', 'x'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"x\"), RangeStart: 'a', RangeEnd 'x'",
			data:         []byte{'x'},
			findFunc:     NewFindValueRangeAlternatives('a', 'x'),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"{\"), RangeStart: 'a', RangeEnd 'x'",
			data:         []byte{'{'},
			findFunc:     NewFindValueRangeAlternatives('a', 'x'),
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestFindVariableRepetition(t *testing.T) {
	tests := []TestCase{
		//
		// NewFindVariableRepetitionMinMax
		//
		{
			testName:     "data: []byte{}, find \"0*1a\"",
			data:         []byte{},
			findFunc:     NewFindVariableRepetitionMinMax(0, 1, NewFindByte('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte{}, find \"1*2a\"",
			data:         []byte{},
			findFunc:     NewFindVariableRepetitionMinMax(1, 2, NewFindByte('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"1*2a\"",
			data:         []byte("a"),
			findFunc:     NewFindVariableRepetitionMinMax(1, 2, NewFindByte('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a\"), find \"2*3a\"",
			data:         []byte("a"),
			findFunc:     NewFindVariableRepetitionMinMax(2, 3, NewFindByte('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"2*3a\"",
			data:         []byte("aa"),
			findFunc:     NewFindVariableRepetitionMinMax(2, 3, NewFindByte('a')),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"aaaa\"), find \"2*3a\"",
			data:         []byte("aaaa"),
			findFunc:     NewFindVariableRepetitionMinMax(2, 3, NewFindByte('a')),
			expectedEnds: []int{2, 3},
		},
		//
		// NewFindVariableRepetitionMin
		//
		{
			testName:     "data: []byte{}, find \"0*a\"",
			data:         []byte{},
			findFunc:     NewFindVariableRepetitionMin(0, NewFindByte('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte{}, find \"1*a\"",
			data:         []byte{},
			findFunc:     NewFindVariableRepetitionMin(1, NewFindByte('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"1*a\"",
			data:         []byte("a"),
			findFunc:     NewFindVariableRepetitionMin(1, NewFindByte('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a\"), find \"2*a\"",
			data:         []byte("a"),
			findFunc:     NewFindVariableRepetitionMin(2, NewFindByte('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"2*a\"",
			data:         []byte("aa"),
			findFunc:     NewFindVariableRepetitionMin(2, NewFindByte('a')),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"2*a\"",
			data:         []byte("aaa"),
			findFunc:     NewFindVariableRepetitionMin(2, NewFindByte('a')),
			expectedEnds: []int{2, 3},
		},
		//
		// NewFindVariableRepetitionMax
		//
		{
			testName:     "data: []byte{}, find \"*1a\"",
			data:         []byte{},
			findFunc:     NewFindVariableRepetitionMax(1, NewFindByte('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte(\"a\"), find \"*2a\"",
			data:         []byte("a"),
			findFunc:     NewFindVariableRepetitionMax(2, NewFindByte('a')),
			expectedEnds: []int{0, 1},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"*2a\"",
			data:         []byte("aa"),
			findFunc:     NewFindVariableRepetitionMax(2, NewFindByte('a')),
			expectedEnds: []int{0, 1, 2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"*3a\"",
			data:         []byte("aaa"),
			findFunc:     NewFindVariableRepetitionMax(3, NewFindByte('a')),
			expectedEnds: []int{0, 1, 2, 3},
		},
		{
			testName:     "data: []byte(\"aaaa\"), find \"*3a\"",
			data:         []byte("aaaa"),
			findFunc:     NewFindVariableRepetitionMax(3, NewFindByte('a')),
			expectedEnds: []int{0, 1, 2, 3},
		},
		//
		// NewFindVariableRepetition
		//
		{
			testName:     "data: []byte{}, find \"*a\"",
			data:         []byte{},
			findFunc:     NewFindVariableRepetition(NewFindByte('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte(\"a\"), find \"*a\"",
			data:         []byte("a"),
			findFunc:     NewFindVariableRepetition(NewFindByte('a')),
			expectedEnds: []int{0, 1},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"*a\"",
			data:         []byte("aa"),
			findFunc:     NewFindVariableRepetition(NewFindByte('a')),
			expectedEnds: []int{0, 1, 2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"*a\"",
			data:         []byte("aaa"),
			findFunc:     NewFindVariableRepetition(NewFindByte('a')),
			expectedEnds: []int{0, 1, 2, 3},
		},
		//
		// combination with Concatenation
		//
		{
			testName: "data: []byte(\"a1b2c3\"), find \"*(ALPHA DIGIT)\"",
			data:     []byte("a1b2c3"),
			findFunc: NewFindVariableRepetition(NewFindConcatenation([]FindFunc{
				FindAlpha,
				FindDigit,
			})),
			expectedEnds: []int{
				0,
				2,
				4,
				6,
			},
		},
		//
		// combination with Concatenation and VariableRepetition
		//
		{
			testName: "data: []byte(\"a1bc3\"), find \"*(ALPHA *DIGIT)\"",
			data:     []byte("a1bc3"),
			findFunc: NewFindVariableRepetition(NewFindConcatenation([]FindFunc{
				FindAlpha,
				NewFindVariableRepetition(FindDigit),
			})),
			expectedEnds: []int{
				0, // ""
				1, // "a"
				2, // "a1"
				3, // "a1b"
				4, // "a1bc"
				5, // "a1bc3"
			},
		},
	}
	execTest(tests, t)
}

func TestFindSpecificRepetition(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"0a\"",
			data:         []byte{},
			findFunc:     NewFindSpecificRepetition(0, NewFindByte('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte{}, find \"1a\"",
			data:         []byte{},
			findFunc:     NewFindSpecificRepetition(1, NewFindByte('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find \"1a\"",
			data:         []byte("a"),
			findFunc:     NewFindSpecificRepetition(1, NewFindByte('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"1a\"",
			data:         []byte("ab"),
			findFunc:     NewFindSpecificRepetition(1, NewFindByte('a')),
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a\"), find \"2a\"",
			data:         []byte("a"),
			findFunc:     NewFindSpecificRepetition(2, NewFindByte('a')),
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"aa\"), find \"2a\"",
			data:         []byte("aa"),
			findFunc:     NewFindSpecificRepetition(2, NewFindByte('a')),
			expectedEnds: []int{2},
		},
		{
			testName:     "data: []byte(\"aaa\"), find \"2a\"",
			data:         []byte("aaa"),
			findFunc:     NewFindSpecificRepetition(2, NewFindByte('a')),
			expectedEnds: []int{2},
		},
	}
	execTest(tests, t)
}

func TestFindOptionalSequence(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find \"[ a ]\"",
			data:         []byte{},
			findFunc:     NewFindOptionalSequence(NewFindByte('a')),
			expectedEnds: []int{0},
		},
		{
			testName:     "data: []byte(\"a\"), find \"[ a ]\"",
			data:         []byte("a"),
			findFunc:     NewFindOptionalSequence(NewFindByte('a')),
			expectedEnds: []int{0, 1},
		},
		{
			testName:     "data: []byte(\"ab\"), find \"[ a ]\"",
			data:         []byte("ab"),
			findFunc:     NewFindOptionalSequence(NewFindByte('a')),
			expectedEnds: []int{0, 1},
		},
	}
	execTest(tests, t)
}

func TestFindAlpha(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find ALPHA",
			data:         []byte{},
			findFunc:     FindAlpha,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find ALPHA",
			data:         []byte("a"),
			findFunc:     FindAlpha,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"ab\"), find ALPHA",
			data:         []byte("ab"),
			findFunc:     FindAlpha,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1\"), find ALPHA",
			data:         []byte("1"),
			findFunc:     FindAlpha,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"12\"), find ALPHA",
			data:         []byte("12"),
			findFunc:     FindAlpha,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"A\"), find ALPHA",
			data:         []byte("A"),
			findFunc:     FindAlpha,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"AB\"), find ALPHA",
			data:         []byte("AB"),
			findFunc:     FindAlpha,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1A\"), find ALPHA",
			data:         []byte("1A"),
			findFunc:     FindAlpha,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a1\"), find ALPHA",
			data:         []byte("a1"),
			findFunc:     FindAlpha,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1A\"), find ALPHA",
			data:         []byte("1A"),
			findFunc:     FindAlpha,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"A1\"), find ALPHA",
			data:         []byte("A1"),
			findFunc:     FindAlpha,
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestFindDigit(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find DIGIT",
			data:         []byte{},
			findFunc:     FindDigit,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find DIGIT",
			data:         []byte("a"),
			findFunc:     FindDigit,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find DIGIT",
			data:         []byte("ab"),
			findFunc:     FindDigit,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1\"), find DIGIT",
			data:         []byte("1"),
			findFunc:     FindDigit,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"12\"), find DIGIT",
			data:         []byte("12"),
			findFunc:     FindDigit,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A\"), find DIGIT",
			data:         []byte("A"),
			findFunc:     FindDigit,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"AB\"), find DIGIT",
			data:         []byte("AB"),
			findFunc:     FindDigit,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1a\"), find DIGIT",
			data:         []byte("1a"),
			findFunc:     FindDigit,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a1\"), find DIGIT",
			data:         []byte("a1"),
			findFunc:     FindDigit,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1A\"), find DIGIT",
			data:         []byte("1A"),
			findFunc:     FindDigit,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A1\"), find DIGIT",
			data:         []byte("A1"),
			findFunc:     FindDigit,
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}

func TestFindDQuote(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find DQUOTE",
			data:         []byte{},
			findFunc:     FindDQuote,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find DQUOTE",
			data:         []byte("a"),
			findFunc:     FindDQuote,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"'\"), find DQUOTE",
			data:         []byte("'"),
			findFunc:     FindDQuote,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"\"\"), find DQUOTE",
			data:         []byte("\""),
			findFunc:     FindDQuote,
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestFindHexDig(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find HEXDIG",
			data:         []byte{},
			findFunc:     FindHexDig,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find HEXDIG",
			data:         []byte("a"),
			findFunc:     FindHexDig,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"ab\"), find HEXDIG",
			data:         []byte("ab"),
			findFunc:     FindHexDig,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1\"), find HEXDIG",
			data:         []byte("1"),
			findFunc:     FindHexDig,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"12\"), find HEXDIG",
			data:         []byte("12"),
			findFunc:     FindHexDig,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A\"), find HEXDIG",
			data:         []byte("A"),
			findFunc:     FindHexDig,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"AB\")",
			data:         []byte("AB"),
			findFunc:     FindHexDig,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"1a\"), find HEXDIG",
			data:         []byte("1a"),
			findFunc:     FindHexDig,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"a1\"), find HEXDIG",
			data:         []byte("a1"),
			findFunc:     FindHexDig,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"1A\")",
			data:         []byte("1A"),
			findFunc:     FindHexDig,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte(\"A1\")",
			data:         []byte("A1"),
			findFunc:     FindHexDig,
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestFindHTab(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find HTAB",
			data:         []byte{},
			findFunc:     FindHTab,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find HTAB",
			data:         []byte("a"),
			findFunc:     FindHTab,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x20}, find HTAB",
			data:         []byte{0x20},
			findFunc:     FindHTab,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x09}, find HTAB",
			data:         []byte{0x09},
			findFunc:     FindHTab,
			expectedEnds: []int{1},
		},
	}
	execTest(tests, t)
}

func TestFindOctet(t *testing.T) {
	tests := []TestCase{
		{
			testName:     "data: []byte{}, find Octet",
			data:         []byte{},
			findFunc:     FindOctet,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x00}, find Octet",
			data:         []byte{0x00},
			findFunc:     FindOctet,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0xff}, find Octet",
			data:         []byte{0xff},
			findFunc:     FindOctet,
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
			findFunc:     FindSp,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte(\"a\"), find SP",
			data:         []byte("a"),
			findFunc:     FindSp,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x20}, find SP",
			data:         []byte{0x20},
			findFunc:     FindSp,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0x09}, find SP",
			data:         []byte{0x09},
			findFunc:     FindSp,
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
			findFunc:     FindVChar,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x20}, find VCHAR",
			data:         []byte{0x20},
			findFunc:     FindVChar,
			expectedEnds: []int{},
		},
		{
			testName:     "data: []byte{0x21}, find VCHAR",
			data:         []byte{0x21},
			findFunc:     FindVChar,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0x7e}, find VCHAR",
			data:         []byte{0x7e},
			findFunc:     FindVChar,
			expectedEnds: []int{1},
		},
		{
			testName:     "data: []byte{0x7f}, find VCHAR",
			data:         []byte{0x7f},
			findFunc:     FindVChar,
			expectedEnds: []int{},
		},
	}
	execTest(tests, t)
}
