package utils

import "testing"

type A struct {
	a string
	b *B
}

type B struct {
	a int
	b string
	c string
}

func TestSpoil(t *testing.T) {
	str := "This is a spooky spoiler!"
	exp := "||This is a spooky spoiler!||"
	got := Spoil(str)
	if got != exp {
		t.Errorf("Spoil(\"%s\") = %s; want %s", str, got, exp)
	}
}

func TestStrlen(t *testing.T) {
	str := "this is an example string"
	got := Strlen(str)
	exp := len(str)
	if got != exp {
		t.Errorf("Strlen(\"%s\") = %d; want %d", str, got, exp)
	}

	got = Strlen(&str)
	exp = len(str)
	if got != exp {
		t.Errorf("Strlen(\"%p\") = %d; want %d", &str, got, exp)
	}

	s := struct {
		a string
		b int
		c interface{}
	}{str, 42, nil}
	got = Strlen(s)
	exp = len(str)
	if got != exp {
		t.Errorf("Strlen(%v) = %d; want %d", s, got, exp)
	}

	as := &A{
		str,
		&B{
			42,
			str,
			str + str,
		},
	}
	got = Strlen(as)
	exp = 4 * len(str)
	if got != exp {
		t.Errorf("Strlen(%v) = %d; want %d", as, got, exp)
	}
}
