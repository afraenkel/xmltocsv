package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
)

// read in a simple json file (output files)
func readJSON(filepath string) map[string]string {
	var x map[string]string
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(file, &x)
	return x
}

//
func TestSafeAddKey(t *testing.T) {
	m := map[string]string{"key": "val"}
	m1 := map[string]string{"key": "val", "key.1": "val"}

	cases := []struct {
		in    string
		inmap map[string]string
		want  string
	}{
		{"mykey", m, "mykey"},
		{"key", m, "key.1"},
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
func TestOneToNum(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"key.1", "key.1"},
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

//
func TestCleanHeader(t *testing.T) {
	cases := []struct {
		in, want []string
	}{
		{[]string{"hi", "hi.1", "hi.1.1"}, []string{"hi", "hi.1", "hi.2"}},
	}
	for _, c := range cases {
		got := cleanHeader(c.in)
		for k, v := range got {
			if v != c.want[k] {
				t.Errorf("safeAddKey(%q) == %q, want %q", c.in, got, c.want)
			}
		}
	}
}

//
func TestParseRecord(t *testing.T) {
	cases := map[string][]string{}

	basedir := "./test_data/parseRecord"
	xmlfiles, _ := ioutil.ReadDir(basedir)
	for _, xmlfile := range xmlfiles {
		fpath := filepath.Join(basedir, xmlfile.Name())
		base := fpath[:len(fpath)-4]
		ext := fpath[len(fpath)-4:]
		if ext == ".xml" {
			cases[base] = append(cases[base], fpath)
		} else {
			cases[base] = append(cases[base], fpath)
		}
	}
	for base, path := range cases {
		want := readJSON(path[1])
		file, err := ioutil.ReadFile(path[0])
		if err != nil {
			log.Fatal(err)
		}
		got := parseRecord(string(file[:]))
		for key, val := range want {
			if val != got[key] {
				t.Errorf("Error in case %q: got %q, want %q", base, got, want)
			}
		}
	}
}

//

//
func TestParseLines(t *testing.T) {
}

//
func TestMain(t *testing.T) {

}
