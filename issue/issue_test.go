package issue

import (
	. "testing"
)

func assertEqual(t *T, e, a interface{}) {
	if e != a {
		t.Errorf("Expected '%s', got '%s'", e, a)
	}
}

func TestMapSprintf(t *T) {
	e := "hello great world"
	a := MapSprintf("%{foo} %{fee} %{fum}", H{"foo": "hello", "fee": "great", "fum": "world"})
	assertEqual(t, e, a)
}

func TestMapSprintfIgnoredFlags(t *T) {
	e := "234d, 23o, 23X"
	a := MapSprintf("%{foo}4d, %{foo}o, %{foo}X", H{"foo": 23})
	assertEqual(t, e, a)
}

func TestMapSprintfFlags(t *T) {
	e := "  23, 27, 17"
	a := MapSprintf("%<foo>4d, %<foo>o, %<foo>X", H{"foo": 23})
	assertEqual(t, e, a)
}

func TestMapSprintfDups(t *T) {
	e := "boys will be boys"
	a := MapSprintf("%{foo} %{fee} %{foo}", H{"foo": "boys", "fee": "will be"})
	assertEqual(t, e, a)
}

func TestMapSprintfMissingKey(t *T) {
	defer func() {
		if err := recover(); err != nil {
			assertEqual(t, err, "missing argument matching key {fee} in format string %{foo} %{fee} %{fum}")
		} else {
			t.Errorf("Expected missing key error but nothing was raised")
		}
	}()
	MapSprintf("%{foo} %{fee} %{fum}", H{"foo": "hello", "fum": "world"})
}
