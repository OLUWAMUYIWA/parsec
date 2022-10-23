package parsec

import (
	"container/list"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

type TestInput struct {
	in []byte
}

func (i *TestInput) Car() byte {
	return (*i).in[0]
}

func (i *TestInput) Cdr() ParserInput {
	return &TestInput{
		in: (*i).in[1:],
	}
}

func (i *TestInput) Empty() bool {
	return len((*i).in) == 0
}

func (i *TestInput) String() string {
	var s strings.Builder

	for _, r := range i.in {
		s.WriteByte(r)
	}

	return s.String()
}

func TestTag(t *testing.T) {
	var inTable []*TestInput = []*TestInput{
		{[]byte{'a', 'b', 'c'}},
		{[]byte{'d', 'e', 'f'}},
	}

	expected := []byte{'a', 'd'}

	for i, r := range expected {
		res := Tag(r)(inTable[i])
		resR := res.Result.(byte)
		if resR != r {
			t.Errorf("Tag isn't popping the right byte \n")
		}
		inTable[i] = res.Rem.(*TestInput)
	}

	if !reflect.DeepEqual(*inTable[0], TestInput{[]byte{'b', 'c'}}) {
		t.Errorf("Remander is not correct: %s instead of: %s \n", inTable[0], &TestInput{[]byte{'b', 'c'}})
	}

	if !reflect.DeepEqual(*inTable[1], TestInput{[]byte{'e', 'f'}}) {
		t.Errorf("Remander is not correct: %s instead of %s \n", inTable[1], &TestInput{[]byte{'e', 'f'}})
	}

}

func TestIsNot(t *testing.T) {
	var inTable []*TestInput = []*TestInput{
		{[]byte{'a', 'b', 'c'}},
		{[]byte{'d', 'e', 'f'}},
	}

	bytes := []byte{'b', 'e'}

	for i, r := range bytes {
		res := IsNot(r)(inTable[i])
		resR := res.Result.(byte)
		if res.Err != nil {
			t.Errorf("Errored: %s", res.Err)
		}
		if resR == r {
			t.Errorf("IsNot matches the said byte: %v instead of not doing so %v\n", resR, r)
		}
	}
}

func TestCharUTF8(t *testing.T) {
	//these guys ar invalid utf-8s, they should fail
	nonUTF8s := &TestInput{
		in: []byte("pythön!"),
	}

	for !nonUTF8s.Empty() {
		res := CharUTF8()(nonUTF8s)
		if e, did := res.Errored(); did {
			if !errors.Is(e, UnmatchedErr()) {
				t.Errorf("Sould be unmatched error")
			}
		}
		nonUTF8s = res.Rem.(*TestInput)

	}

	nonUTF8s = &TestInput{
		in: []byte{0xe, 0xe},
	}

	res := CharUTF8()(nonUTF8s)
	if res.Result != nil {
		t.Errorf("should be nill, but is: %v", res.Result)
	}

	if !reflect.DeepEqual(*nonUTF8s, *(res.Rem.(*TestInput))) {
		t.Errorf("Should both be equal")
	}
}

func TestOneOf(t *testing.T) {
	in := TestInput{
		in: []byte{'d', 'e', 'f'},
	}

	any := []byte{'d', 'a', 'b', 'c'}

	res := OneOf(any)(&in)

	if e, didErr := res.Errored(); didErr {
		t.Errorf("Errored when it shouldn't: %s", e)
	}

	if reflect.DeepEqual(*(res.Rem.(*TestInput)), in) {
		t.Errorf("should not be equal beacuse it got reduced")
	}
	if res.Result.(byte) != byte('d') {
		t.Errorf("Should be%b, but is %b", byte('d'), res.Result.(byte))
	}

	if !reflect.DeepEqual(TestInput{in: []byte{'e', 'f'}}, *(res.Rem.(*TestInput))) {
		t.Errorf("should  be equal beacuse it got reduced")
	}
}

func TestDigit(t *testing.T) {
	in := &TestInput{
		in: []byte{'1', '2', '3', '4', '5', 'g'},
	}
	dig := Digit()
	expected := []int{1, 2, 3, 4, 5}

	for _, exp := range expected {
		res := dig(in)
		if res.Result.(int) != exp {
			t.Errorf("should be equal to the integers in expected. Should be: %d, but is %d", exp, res.Result.(int))
		}
		in = res.Rem.(*TestInput)
	}
}

func TestIsEmpty(t *testing.T) {
	in := &TestInput{
		in: []byte{},
	}

	res := IsNot('a')(in)

	if _, didErr := res.Errored(); !didErr {
		t.Errorf("Should have Errored but didn't")
	}

	// if e, did := res.Errored(); did {
	// 		if !errors.Is(e, IncompleteErr()) {
	// 			t.Logf("%v\n",e)
	// 			t.Logf("%v\n",IncompleteErr())
	// 			t.Fail()
	// 		}
	// }

}

func TestLetter(t *testing.T) {
	in := &TestInput{
		in: []byte{'a', 'b', 'c'},
	}

	let := Letter()

	res := let(in)
	if r := res.Result.(byte); !unicode.IsLetter(rune(r)) {
		t.Errorf("Wrong!")
	}
}

func TestTakeN(t *testing.T) {
	in := &TestInput{
		in: []byte{'a', 'b', 'c', 'd', 'e', 'f'},
	}

	take := TakeN(5)

	res := take(in)

	valRes := res.Result.(*list.List)

	if valRes.Len() != 5 {
		t.Errorf("Not all is taken, expected 5, got: %d", valRes.Len())
	}
	i := 0
	for e := valRes.Front(); e != nil; e = e.Next() {
		if val := e.Value.(byte); val != (*in).in[i] {
			t.Errorf("Expected: %d, got: %d", (*in).in[i], val)
		}
		i++
	}
}

func TestTakeTill(t *testing.T) {
	in := &TestInput{
		in: []byte{'a', 'b', 'c', 'd', 'e', 'f'},
	}

	var f Predicate = func(r byte) bool {
		return r == 'e'
	}

	take := TakeTill(f)

	res := take(in)

	resList := res.Result.(*list.List)

	if resList.Len() != 4 {
		t.Errorf("Should be 4")
	}
	i := 0
	for e := resList.Front(); e != nil; e = e.Next() {
		if v := e.Value.(byte); v != (*in).in[i] {
			t.Errorf("Not the bytes we expected")
		}
		i++
	}
}

func TestTakeWhile(t *testing.T) {
	in := &TestInput{
		in: []byte{'a', 'b', 'c', 'd', 'e', 'f', 'h', 'k'},
	}

	var f Predicate = func(r byte) bool {
		return r <= 'e'
	}

	take := TakeWhile(f)

	res := take(in)

	resList := res.Result.([]byte)

	if len(resList) != 5 {
		t.Errorf("Should be 5")
	}
	expexted := []byte{'a', 'b', 'c', 'd', 'e'}
	if !reflect.DeepEqual(resList, expexted) {
		t.Errorf("Not the bytes we expected: %v VS %v", resList, expexted)
	}

}

func TestTerminated(t *testing.T) {
	in := &TestInput{
		in: []byte{'c', 'a', 't', 'd', 'o', 'g', 'h', 'k'},
	}
	match := "cat"
	parser := Terminated(match, "dog")

	res := parser(in)
	if err, did := res.Errored(); did {
		t.Errorf("Errored when it shouldn't: %s", err)
	}

	if ret := res.Result.(string); ret != "cat" {
		t.Errorf("Should return: %s, but got: %s", match, ret)
	}

	in = &TestInput{
		in: []byte{'c', 'a', 't', 'd', 'o', 'g', 'h', 'k'},
	}

	match = "cats"
	parser = Terminated(match, "dog")
	res = parser(in)
	if _, did := res.Errored(); !did {
		t.Errorf("Should have errored")
	}

	if ret := res.Result; ret != nil {
		t.Errorf("Should return: nil, but got: %s", ret)
	}

}

func TestPreceded(t *testing.T) {
	in := &TestInput{
		in: []byte{'c', 'a', 't', 'd', 'o', 'g', 'h', 'k'},
	}
	match := "dog"
	pre := "cat"
	parser := Preceded(match, pre)

	res := parser(in)
	if err, did := res.Errored(); did {
		t.Errorf("Errored when it shouldn't: %s", err)
	}

	if ret := res.Result.(string); ret != match {
		t.Errorf("Should return: %s, but got: %s", match, ret)
	}

	in = &TestInput{
		in: []byte{'c', 'a', 't', 'd', 'o', 'g', 'h', 'k'},
	}

	match = "dogs"
	pre = "cat"
	parser = Preceded(match, pre)
	res = parser(in)
	if _, did := res.Errored(); !did {
		t.Errorf("Should have errored")
	}

	if ret := res.Result; ret != nil {
		t.Errorf("Should return: nil, but got: %s", ret)
	}
}

func TestNumber(t *testing.T) {
	in := &TestInput{
		in: []byte{'2', '5', '6', 'd', 'o', 'g', 'h', 'k'},
	}
	expexted := 256
	res := Number()(in)
	if ans := res.Result.(int); ans != expexted {
		t.Errorf("Expected %d, found %d", expexted, ans)
	}

	rem := res.Rem.(*TestInput)
	expectedRem := &TestInput{
		in: in.in[3:],
	}
	if len(rem.in) != len(expectedRem.in) {
		t.Errorf("expected : %d, foundL %d", len(expectedRem.in), len(rem.in))
	}

	in2 := &TestInput{
		in: []byte{'-', '2', '5', '6', 'd', 'o', 'g', 'h', 'k'},
	}
	expextedNeg := -256
	res2 := Number()(in2)
	if ans := res2.Result.(int); ans != expextedNeg {
		t.Errorf("Expected %d, found %d", expextedNeg, ans)
	}

	rem2 := res2.Rem.(*TestInput)
	expectedRem2 := &TestInput{
		in: in.in[4:],
	}
	if len(rem.in) != len(expectedRem.in) {
		t.Errorf("expected : %d, foundL %d", len(expectedRem2.in), len(rem2.in))
	}
}

func TestChars(t *testing.T) {
	in := &TestInput{
		in: []byte{'2', '5', '6', 'd', 'o', 'g', 'h', 'k'},
	}

	chars := Chars([]byte{'2', '5', '6', 'd'})

	res := chars(in)

	expected := []byte{'2', '5', '6', 'd'}
	result := res.Result.([]byte)
	if e, did := res.Errored(); did {
		t.Errorf("Error: %s", e)
	}

	for i, r := range expected {
		if r != result[i] {
			t.Errorf("Should be: %v, but is %v", r, result[i])
		}
	}
}

func TestStr(t *testing.T) {
	str := "abeg"

	in := &TestInput{
		in: []byte{'a', 'b', 'e', 'g', 'o', 'g', 'h', 'k'},
	}

	strParsec := Str(str)

	res := strParsec(in)

	if err, did := res.Errored(); did {
		t.Errorf("Errored: %s", err)
	}

	if s, ok := res.Result.(string); ok {
		if s != str {
			t.Errorf("Expected: %s, found: %s", str, s)
		}
	} else {
		t.Errorf("Could not convert to string")
	}

}

func TestMany0(t *testing.T) {

	in := &TestInput{
		in: []byte{'a', 'a', 'a', 'a', 'o', 'g', 'h', 'k'},
	}

	isA := Tag('a')
	many0_Tag := isA.Many0()
	res := many0_Tag(in)

	if err, did := res.Errored(); did {
		t.Errorf("Should never error. Error: %s", err)
	}

	lRes, ok := res.Result.(*list.List)

	if lRes.Len() != 4 {
		t.Errorf("list length should be 4")
	}

	// expected := []int32{'a', 'a', 'a', 'a'}
	for e := lRes.Front(); e != nil; e = e.Next() {
		if !reflect.DeepEqual(e.Value.(byte), byte('a')) {
			t.Errorf("Expected: %v, got %v", uint('a'), e.Value)
		}
	}
	// if !reflect.DeepEqual(lRes, expected) {
	// 	t.Errorf("Saw: %v", res.Result)
	// 	t.Errorf("Expected: %v, got %v", expected, lRes)
	// }

	in = &TestInput{
		in: []byte{'a', 'a', 'a', 'a', 'o', 'g', 'h', 'k'},
	}

	isA = Tag('b')
	many0_Tag = isA.Many0()
	res = many0_Tag(in)

	if err, did := res.Errored(); did {
		t.Errorf("Error: %s", err)
	}

	lRes2, ok := res.Result.(*list.List)

	if !ok {
		t.Errorf("SHould be a list but isn't")
	}

	if lRes2.Len() != 0 {
		t.Errorf("list length should be 0")
	}

}

func TestMany1(t *testing.T) {
	in := &TestInput{
		in: []byte{'a', 'a', 'a', 'a', 'o', 'g', 'h', 'k'},
	}

	isA := Tag('a')
	many1_Tag := isA.Many1()
	res := many1_Tag(in)

	if err, did := res.Errored(); did {
		t.Errorf("Error: %s", err)
	}

	lRes, ok := res.Result.(*list.List)

	if !ok {
		t.Errorf("SHould be a list but isn't")
	}

	if lRes.Len() != 4 {
		t.Errorf("list length should be 4")
	}

	for v := lRes.Front(); v != nil; v = v.Next() {
		r := v.Value.(byte)
		if r != 'a' {
			t.Errorf("Expected: %s, found: %s", "a", string(r))
		}
	}

	// part2
	in = &TestInput{
		in: []byte{'a', 'a', 'a', 'a', 'o', 'g', 'h', 'k'},
	}

	isA = Tag('b')
	many1_Tag = isA.Many1()
	res = many1_Tag(in)

	if err, did := res.Errored(); !did {
		t.Errorf("Should error but didn't. Error: %s", err)
	}

	lRes2, ok := res.Result.(*list.List)

	if !ok {
		t.Errorf("SHould be a list but isn't")
	}

	if lRes2.Len() != 0 {
		t.Errorf("list length should be 0")
	}

}

func TestCount(t *testing.T) {
	in := &TestInput{
		in: []byte{'a', 'a', 'a', 'a', 'a', 'o', 'g', 'h', 'k'},
	}

	isA := Tag('a')
	count := isA.Count(5)

	res := count(in)

	if err, did := res.Errored(); did {
		t.Errorf("Error unexpected: %s", err)
	}

	lRes, ok := res.Result.(*list.List)

	if !ok {
		t.Errorf("SHould be a list but isn't")
	}

	if lRes.Len() != 5 {
		t.Errorf("list length should be 5")
	}

	for v := lRes.Front(); v != nil; v = v.Next() {
		r := v.Value.(byte)
		if r != 'a' {
			t.Errorf("Expected: %s, found: %s", "a", string(r))
		}
	}

	// part 2

	count2 := isA.Count(3)

	res2 := count2(in)

	if err, did := res2.Errored(); did {
		t.Errorf("Error unexpected: %s", err)
	}

	lRes2, ok := res2.Result.(*list.List)

	if !ok {
		t.Errorf("SHould be a list but isn't")
	}

	if lRes2.Len() != 3 {
		t.Errorf("list length should be 5")
	}

	for v := lRes2.Front(); v != nil; v = v.Next() {
		r := v.Value.(byte)
		if r != 'a' {
			t.Errorf("Expected: %s, found: %s", "a", string(r))
		}
	}
}

func TestCount2(t *testing.T) {

	// pass 3
	in3 := &TestInput{
		in: []byte{'a', 'a', 'a', 'a', 'a', 'o', 'g', 'h', 'k'},
	}
	count3 := Tag('a').Count(10)

	res3 := count3(in3)

	if _, did := res3.Errored(); !did {
		t.Errorf("Error expected")
	}

	if res3.Result != nil {
		t.Errorf("result should be nil")
	}

	if !reflect.DeepEqual(res3.Rem.(*TestInput), in3) {
		t.Errorf("Cdr not correct: %v vs %v", res3.Rem.Cdr(), in3)
	}
}

// we use Table-driven tests ere
func TestThen(t *testing.T) {
	type test struct {
		input *TestInput
		want  int
	}
	pry := OneOf([]byte{'a', 'b', 'c', '4'})
	sec := Digit()
	tests := []test{
		{input: &TestInput{in: []byte{'b', '6'}}, want: 6},
		{input: &TestInput{in: []byte{'c', '9'}}, want: 9},
		{input: &TestInput{in: []byte{'x', '1'}}, want: 0},
	}
	parser := pry.Then(sec)

	for _, tt := range tests {
		res := parser(tt.input)
		if res.Err != nil { // failed either at first or second parser
			if !reflect.DeepEqual(tt.input, res.Rem.(*TestInput)) {
				t.Errorf("since we failed, we should get full input")
			}
		} else {
			result, ok := res.Result.(int)
			if !ok {
				t.Errorf("Value: %d", result)
				t.Errorf("should be an int but isn't. instead: %s", reflect.TypeOf(result))
			}
			if result != tt.want {
				t.Errorf(" wron value. expected %d, got %d", tt.want, result)
			}
		}
	}
}

func TestThenDiscard(t *testing.T) {
	type test struct {
		input *TestInput
		want  int
	}
	pry := Digit()
	sec := OneOf([]byte{'4', 'a', 'b', 'c'})

	tests := []test{
		{input: &TestInput{in: []byte{'6', 'b', '5', 'u'}}, want: 6},
		{input: &TestInput{in: []byte{'9', 'c'}}, want: 9},
		{input: &TestInput{in: []byte{'1', 'x'}}, want: 0},
	}
	parser := pry.ThenDiscard(sec)

	for _, tt := range tests {
		res := parser(tt.input)
		if res.Err != nil { // failed either at first or second parser
			if !reflect.DeepEqual(tt.input, res.Rem.(*TestInput)) {
				t.Errorf("since we failed, we should get full input")
			}
		} else {
			result, ok := res.Result.(int)
			if !ok {
				t.Errorf("Value: %d", result)
				t.Errorf("should be an int but isn't. instead: %s", reflect.TypeOf(result))
			}
			if result != tt.want {
				t.Errorf(" wrong value. expected %d, got %d", tt.want, result)
			}
		}
	}
}

func TestAndThen(t *testing.T) {
	type wanted struct {
		pre  string
		mid  byte
		post byte
	}

	type test struct {
		input *TestInput
		want  wanted
	}

	pry := Str("abc")
	mid := OneOf([]byte{'4', 'a', 'b', 'c', '6'})
	post := Tag('9')

	tests := []test{
		{input: &TestInput{in: []byte{'a', 'b', 'c', '6', '9', '5', 'u'}}, want: wanted{pre: "abc", mid: '6', post: '9'}},
		{input: &TestInput{in: []byte{'a', 'b', 'c', '4', '5'}}, want: wanted{pre: "abc", mid: '4', post: '5'}},
		{input: &TestInput{in: []byte{'1', 'x'}}, want: wanted{pre: "", mid: 0, post: 0}},
	}
	parser := pry.AndThen([]Parsec{mid, post})

	for _, tt := range tests {
		res := parser(tt.input)
		if res.Err != nil { // failed either at first or second parser
			if !reflect.DeepEqual(tt.input, res.Rem.(*TestInput)) {
				t.Errorf("since we failed, we should get full input")
			}
		} else {
			result, ok := res.Result.(*list.List)
			if !ok {
				t.Errorf("should be a list but isn't. instead: %s", reflect.TypeOf(result))
			}
			pre := result.Front()
			if s, ok := pre.Value.(string); !ok {
				t.Error("value shoule be string")
			} else {
				if s != tt.want.pre {
					t.Errorf("Expected: %s, found: %s", tt.want.pre, s)
				}
			}

			mid := pre.Next()
			if r, ok := mid.Value.(byte); !ok {
				t.Error("value shoule be byte")
			} else {
				if r != tt.want.mid {
					t.Errorf("Expected: %d, found: %d", tt.want.mid, r)
				}
			}

			post := mid.Next()
			if r, ok := post.Value.(byte); !ok {
				t.Error("value shoule be byte")
			} else {
				if r != tt.want.post {
					t.Errorf("Expected: %d, found: %d", tt.want.post, r)
				}
			}
		}
	}
}

func TestAlt(t *testing.T) {
	first := Str("abc")
	second := OneOf([]byte{'e', 'a', 'o'})
	third := Tag('9')

	type test[T any] struct {
		input *TestInput
		want  T
	}

	test1 := test[string]{input: &TestInput{in: []byte{'a', 'b', 'c', '6', '9', '5', 'u'}}, want: "abc"}
	test2 := test[byte]{input: &TestInput{in: []byte{'a', 'b', 'c', '4', '5'}}, want: 'a'}
	test3 := test[byte]{input: &TestInput{in: []byte{'1', 'x'}}}

	parser := Alt(first, second, third)

	res1 := parser(test1.input)
	result1, ok := res1.Result.(string)
	if !ok {
		t.Errorf("Result should be string. the string parser should be chosen, but i isnt: %s", reflect.TypeOf(res1.Result))
	}
	if result1 != test1.want {
		t.Errorf("should be abc but is '%s'", result1)
	}
	exptdrem1 := &TestInput{in: []byte{'6', '9', '5', 'u'}}
	if !reflect.DeepEqual(exptdrem1, res1.Rem.(*TestInput)) {
		t.Errorf("expected %s to remain but remained %s", exptdrem1, res1.Rem.(*TestInput))
	}

	// changed the order to allow the second to go first
	parser2 := Alt(second, first, third)
	res2 := parser2(test2.input)
	result2, ok := res2.Result.(byte)
	if !ok {
		t.Errorf("Result should be byte. the string parser should be chosen, but i isnt: %s", reflect.TypeOf(res2.Result))
	}
	if result2 != test2.want {
		t.Errorf("should be %d but is '%d'", test2.want, result2)
	}

	res3 := parser(test3.input)
	if _, did := res3.Errored(); !did {
		t.Errorf("should have errored but didnt")
	}
	if res3.Result != nil {
		t.Errorf("parser should have nil result as none of the parsers matched")
	}
	if !reflect.DeepEqual(test3.input, res3.Rem.(*TestInput)) {
		t.Errorf("full remainder should remain")
	}
}

func TestGuarded(t *testing.T) {
	l, r := byte('a'), byte('z')
	g := Guarded(l, r)

	type test struct {
		input  *TestInput
		wanted *list.List
		rem    *TestInput
	}

	l1 := list.New()
	l1.PushBack('b')
	l1.PushBack('c')
	l1.PushBack('d')
	l2 := list.New()
	l2.PushBack('2')
	tests := []test{
		{input: &TestInput{in: []byte{'a', 'b', 'c', 'd', 'z'}}, wanted: l1, rem: &TestInput{in: []byte{}}},
		{input: &TestInput{in: []byte{'a', '2', 'z', 'z', 'z'}}, wanted: l2, rem: &TestInput{in: []byte{'z', 'z'}}},
	}

	for _, tt := range tests {
		res := g(tt.input)
		if err, did := res.Errored(); did {
			t.Errorf("should not have errored but did: %s", err)
		}
		if !reflect.DeepEqual(res.Rem.(*TestInput), tt.rem) {
			t.Errorf("remainder not correct")
		}
		if !reflect.DeepEqual(tt.wanted, res.Result.(*list.List)) {
			t.Errorf("Wrong result")
		}
	}

	sectest := test{
		input: &TestInput{in: []byte{'f', 'c'}}, rem: &TestInput{[]byte{'f', 'c'}}, wanted: nil,
	}
	ressec := g(sectest.input)
	if err, did := ressec.Errored(); !did { // redundant
		t.Errorf("should have errored")
		if !errors.Is(UnmatchedErr(), err) {
			t.Errorf("error should be unmatched but is %s", err.Error())
		}
	}

	if ressec.Result != nil {
		t.Errorf("result should be nil")
	}

	terttest := test{
		input: &TestInput{in: []byte{'a', 'f', 'c'}}, rem: &TestInput{[]byte{'a', 'f', 'c'}}, wanted: nil,
	}
	restert := g(terttest.input)
	if err, did := restert.Errored(); !did { // redundant
		t.Errorf("should have errored")
		if !errors.Is(UnmatchedErr(), err) {
			t.Errorf("error should be unmatched but is %s", err.Error())
		}
	}

	if restert.Result != nil {
		t.Errorf("result should be nil")
	}

	if !reflect.DeepEqual(restert.Rem.(*TestInput), terttest.rem) {
		t.Errorf("stream should not have been consumed")

	}
}

func TestGuardedWhile(t *testing.T) {
	l, r := byte('i'), byte('e')
	p := GuardedWhile(l, r, func(r byte) bool { return unicode.IsDigit(rune(r)) })

	type test struct {
		input  *TestInput
		wanted []byte
		rem    *TestInput
	}
	s1 := []byte{'1', '2', '3'}
	tests := []test{
		{input: &TestInput{in: []byte{'i', '1', '2', '3', 'e'}}, wanted: s1, rem: &TestInput{in: []byte{}}},
		{input: &TestInput{in: []byte{'i', '2', '2', 'c', 'c'}}, wanted: nil, rem: &TestInput{in: []byte{'i', '2', '2', 'c', 'c'}}},
		{input: &TestInput{in: []byte{'a', '2', '2', 'c', 'c'}}, wanted: nil, rem: &TestInput{in: []byte{'a', '2', '2', 'c', 'c'}}},
	}
	res := p(tests[0].input)
	if err, did := res.Errored(); did {
		t.Errorf("should not have errored but did: %s", err)
	}
	if !reflect.DeepEqual(tests[0].wanted, res.Result.([]byte)) {
		t.Errorf("Wrong result")
	}
	if !reflect.DeepEqual(res.Rem.(*TestInput), tests[0].rem) {
		t.Errorf("remainder not correct. expected: %v. Got %v", tests[0].rem, res.Rem.(*TestInput))
	}

	// li := res.Result.(*list.List)
	// var digits []byte
	// for e := li.Front(); e != nil; e = e.Next() {
	// 	digits = append(digits, e.Value.(byte))
	// }
	digits := res.Result.([]byte)
	num, _ := strconv.ParseInt(string(digits), 10, 0)
	if num != 123 {
		t.Errorf("Should be %d, but is %d\n", 123, num)
	}

	// second part. the failures
	for i := 1; i < len(tests); i++ {
		tt := tests[i]
		res := p(tt.input)
		if _, did := res.Errored(); !did {
			t.Errorf("should have errored but did")
		}
		if !reflect.DeepEqual(res.Rem.(*TestInput), tt.rem) {
			t.Errorf("remainder not correct. should have full remainder: %v, but has: %v", tt.rem, res.Rem.(*TestInput))
		}
		if !reflect.DeepEqual(nil, res.Result) {
			t.Errorf("Wrong result. result should be empty but is: %v", res.Result.([]byte))
		}
	}
}

func TestStrN(t *testing.T) {

	in := &TestInput{
		in: []byte{'a', 'b', 'e', 'g', 'o', 'g', 'h', 'k'},
	}

	res := StrN(5)(in)
	rem := res.Rem

	if res.Result.(string) != "abego" {
		t.Errorf("Should be %s, but i: %s", "abego", res.Result.(string))
	}

	remExp := &TestInput{
		in: []byte{'g', 'h', 'k'},
	}

	for i := 0; i < 3; i++ {
		a := rem.Car()
		rem = rem.Cdr().(*TestInput)
		b := remExp.Car()
		remExp = remExp.Cdr().(*TestInput)
		if a != b {
			t.Errorf("Should be: %s, but is %s", string(b), string(a))
		}
	}

	// failure
	in = &TestInput{
		in: []byte{'a', 'b', 'e', 'g', 'o', 'g', 'h', 'k'},
	}

	em := '😀'
	s := string(em)
	emoji := []byte(s)
	r := []byte{'a', 'b', 'e', 'g'}
	r = append(r, emoji...)

	// exp := &TestInput{
	// 	in: []byte{'a', 'b', 'e', 'g', 240, 159, 152, 128}, // 240 159 152 128 is = '😀'
	// }
	expStr := "abeg😀"
	in = &TestInput{
		in: r,
	}
	strN := StrN(8)
	res = strN(in)
	resStr := res.Result.(string)

	if resStr != expStr {
		t.Errorf("Expected: %s, got: %s", expStr, resStr)
	}

}

// to be tested later

func TestFoldMany0(t *testing.T) {

}

func TestFoldMany1(t *testing.T) {

}
