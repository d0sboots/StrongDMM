package dmmdata

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testReader struct {
	reader io.Reader
}

func (r *testReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *testReader) Name() string {
	return "TestReader"
}

func validateCommon(t *testing.T, input string) {
	validateBase(t, input, false, false)
}

func validateDMM(t *testing.T, input string) {
	validateBase(t, input, true, true)
}

// Test that all the parsing tests produce the same ultimate result.
func validateBase(t *testing.T, input string, isDmm bool, onlyLF bool) {
	assert := assert.New(t)
	// We want to use raw strings for multi-line convenience, but they
	// strip all carriage returns, so we need a stand-in.
	reader := strings.NewReader(strings.ReplaceAll(input, "^", "\r"))
	dmm, err := parse(&testReader{reader})
	require.Nil(t, err)
	assert.NotEqual(isDmm, dmm.IsTgm)
	assert.Equal(1, dmm.KeyLength)
	assert.Equal(5, dmm.MaxX)
	assert.Equal(6, dmm.MaxY)
	assert.Equal(1, dmm.MaxZ)
	expectedLineEnd := "\r\n"
	if onlyLF {
		expectedLineEnd = "\n"
	}
	assert.Equal(expectedLineEnd, dmm.LineBreak)
	for key, prefabs := range dmm.Dictionary {
		if len(prefabs) != 1 {
			t.Fatalf("Wrong number of prefabs for key %s: %d", key, len(prefabs))
		}
		assert.Empty(prefabs[0].Vars().Iterate(), key)
		// Magic number based on the hash of /obj/foo
		id := uint64(key[0]-'a') + 0x37781ed381b00f3
		assert.Equal(id, prefabs[0].Id(), key)

		expectedPath := "/obj/foo" + string(key[0]-'a'+'1')
		assert.Equal(expectedPath, prefabs[0].Path(), key)
	}
	for point, key := range dmm.Grid {
		assert.Equal(1, point.Z, key)

		expectedKey := string(rune((point.Y-point.X+6)%6 + 'a'))
		assert.Equal(Key(expectedKey), key, point)
	}
}

// Test "normal" parsing, in the TGM format generated by regular tools.
func TestParseTGMNormal(t *testing.T) {
	validateCommon(t, `// Comment line^
"a" = (
/obj/foo1)
"b" = (
/obj/foo2)
"c" = (
/obj/foo3)
"d" = (
/obj/foo4)
"e" = (
/obj/foo5)
"f" = (
/obj/foo6)

(1,1,1) = {"
f
e
d
c
b
a
"}
(2,1,1) = {"
e
d
c
b
a
f
"}
(3,1,1) = {"
d
c
b
a
f
e
"}
(4,1,1) = {"
c
b
a
f
e
d
"}
(5,1,1) = {"
b
a
f
e
d
c
"}
`)
}

// Test "normal" parsing, in the DMM format saved by DreamMaker.
func TestParseDMMNormal(t *testing.T) {
	validateDMM(t, `"a" = (/obj/foo1)
"b" = (/obj/foo2)
"c" = (/obj/foo3)
"d" = (/obj/foo4)
"e" = (/obj/foo5)
"f" = (/obj/foo6)

(1,1,1) = {"
fedcb
edcba
dcbaf
cbafe
bafed
afedc
"}
`)
}

// Test "horizontal" parsing, along the X axis instead of the Y axis.
// Also use a variety of different delimiter spacings, to test edge cases.
func TestParseHorizontal(t *testing.T) {
	validateCommon(t, `// Comment line
"a"	 = 	(/obj/foo1)	 ^
"b" = (^
	  /obj/foo2  	)
"c"=(  /obj/foo3  )
"d" =(	/obj/foo4	)
"e"   =   (/obj/foo5)
"f" = (/obj/foo6)  

(1,1,1)	 = 	{"fedcb"}	 
(1,2,1) =  {"
edcba
"}
(1,3,1)= {"dcbaf"}
(1,4,1) ={"cbafe"}
(1,5,1)={"bafed"}  
(1,6,1) = {"afedc"}
`)
}

// Test "block" parsing, with 2 3x3 blocks and 2 2x3 blocks.
func TestParseBlock(t *testing.T) {
	validateCommon(t, `// Comment line
"a" = (/obj/foo1)
"b" = (/obj/foo2)
"c" = (/obj/foo3)
"d" = (/obj/foo4)
"e" = (/obj/foo5)
"f" = (/obj/foo6)

(1,1,1) = {"
fed
edc
dcb
"}
(4,1,1) = {"^
cb
ba
af
"}
(1,4,1) = {"
cba
baf
afe"}
(4,4,1) = {"
fe
ed
dc"}
`)
}

// Test variable reading. Also checks the detection of non-TGM files.
// This tests a lot of whitespace and variable-related edge-case behavior.
// This test (like the others) should parse correctly in BYOND!
func TestParseVars(t *testing.T) {
	assert := assert.New(t)
	dmm, err := parse(&testReader{strings.NewReader(`"aaa" = (
 	/obj/foo1 	{ 	
no_ws=1;
  space  =  "\"	2 \\"  ;  
	tab	=	3	;	
} 	, 	/obj/foo2 	, 	
 	/obj/foo1 	{no_ws=1} 	) 	

// Comment line that shouldn't flag TGM format

(1,1,1) = {"aaa"}
`)})
	require.Nil(t, err)
	assert.False(dmm.IsTgm)
	assert.Equal(3, dmm.KeyLength)
	assert.Equal(1, dmm.MaxX)
	assert.Equal(1, dmm.MaxY)
	assert.Equal(1, dmm.MaxZ)

	require.Len(t, dmm.Dictionary, 1)
	prefabs := dmm.Dictionary["aaa"]
	require.Len(t, prefabs, 3)

	assert.Equal("/obj/foo1", prefabs[0].Path())
	assert.ElementsMatch(prefabs[0].Vars().Iterate(), []string{"no_ws", "space", "tab"})
	assert.Equal("1", prefabs[0].Vars().ValueV("no_ws", ""))
	assert.Equal(`"\"\t2 \\"`, prefabs[0].Vars().ValueV("space", ""))
	assert.Equal("3", prefabs[0].Vars().ValueV("tab", ""))

	assert.Equal("/obj/foo2", prefabs[1].Path())
	assert.Empty(prefabs[1].Vars().Iterate())

	assert.Equal("/obj/foo1", prefabs[2].Path())
	assert.ElementsMatch(prefabs[2].Vars().Iterate(), []string{"no_ws"})
	assert.Equal("1", prefabs[2].Vars().ValueV("no_ws", ""))
}

// Table-based test to check failure edge cases
func TestFailure(t *testing.T) {
	tests := []struct {
		input string
		err   string
	}{
		{input: "/	/", err: "expected comment"},
		{input: `"a"=(/1) "bb"=(/2)`, err: "key length: 1 vs 2"},
		{input: `"a"=/1 "b"=/2`, err: "failed to start a data block"},
		{input: `(1,1,1)={"ab"}`, err: "extra characters at EOL [ab]"},
		// Cover the other branch, where we have a newline first
		{input: "(1,1,1)={\"ab\n\"}", err: "extra characters at EOL [ab]"},
		{input: `(1,1,1,1)`, err: "incorrect number of axis"},
		{input: `(1,1)`, err: "incorrect reading axis [1] (expected 2)"},
		{input: "\"a\"=(/1)\n(\"b\"=/2)", err: "at line 2: strconv.ParseInt"},
	}

	for _, tc := range tests {
		dmm, err := parse(&testReader{strings.NewReader(tc.input)})
		require.Nil(t, dmm, tc.input)
		require.NotNil(t, err, tc.input)
		if !strings.Contains(err.Error(), tc.err) {
			t.Errorf("Error [%s] does not contain [%s] for input [%s]", err.Error(), tc.err, tc.input)
		}
	}
}

func TestEmpty(t *testing.T) {
	assert := assert.New(t)
	dmm, err := parse(&testReader{strings.NewReader("")})
	require.Nil(t, err)
	assert.False(dmm.IsTgm)
	assert.Equal(0, dmm.KeyLength)
	assert.Equal(0, dmm.MaxX)
	assert.Equal(0, dmm.MaxY)
	assert.Equal(0, dmm.MaxZ)
	assert.Equal("\n", dmm.LineBreak)
}

func TestReadFailure(t *testing.T) {
	tests := []struct {
		reader io.Reader
		err    string
	}{
		{reader: iotest.ErrReader(fmt.Errorf("immediate fail")), err: "immediate fail"},
		// Checking the other branch, where we error while parsing the data
		{reader: iotest.TimeoutReader(strings.NewReader(`(1,1,1)={""}`)), err: "timeout"},
	}

	for _, tc := range tests {
		dmm, err := parse(&testReader{tc.reader})
		require.Nil(t, dmm, tc.err)
		require.NotNil(t, err, tc.err)
		if !strings.Contains(err.Error(), tc.err) {
			t.Errorf("Error [%s] does not contain [%s]", err.Error(), tc.err)
		}
	}
}
