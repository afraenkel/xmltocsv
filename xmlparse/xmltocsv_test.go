package main

import "testing"

// 
func TestSafeAddKey(t *testing.T) {
	m := map[string]string{"key": "val"}
	m1 := map[string]string{"key": "val", "key.1": "val"}

	cases := []struct {
		in string
		inmap map[string]string 
		want string
	}{
		{"mykey", m, "mykey"},
		{"key", m,  "key.1"},
		{"key", m1, "key.1.1"},
	}
	for _, c := range cases {
		got := safeAddKey(c.in, c.inmap)
		if got != c.want {
			t.Errorf("safeAddKey(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

//
func TestAddOnes(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{".1", ".1" },
		{".1.1", ".2"},
		{".1.1.1", ".3"},
		{".1.1.1.1", ".4"},
	}
	for _, c := range cases {
		got := addOnes([]byte(c.in))
		if string(got) != c.want {
			t.Errorf("safeAddKey(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

//
func TestOneToNum(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"key.1", "key.1" },
		{"key.1.1", "key.2"},
		{"key.1.1.1", "key.3"},
		{"key", "key"},
	}
	for _, c := range cases {
		got := oneToNum(c.in)
		if got != c.want {
			t.Errorf("safeAddKey(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

