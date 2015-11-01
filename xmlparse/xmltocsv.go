package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"bufio"
	"strings"
	"regexp"
	"strconv"
	"bytes"
)

// delim is the output delimiter (for the output csv file).
const delim string =  ","

// keysep is the separator used to join the keys of the XML.
const keysep string = "."


// safeAddKey checks if a string key is a key of a map m.
// If it is a key, then safeAddKey adds as many '.1' suffixes
// to key until it's not a key of map m, then returns that string.
// Otherwise, safeAddKey returns the original input key.
func safeAddKey(key string, m map[string]string) string {
	for {
		if _, iskey := m[key]; iskey {
			key = key + ".1"
		} else {
			break
		}
	}
	return key
}


// parseRecord returns a map of of key-values from an XML blob.
// The keys of the map are paths of the XML keys.
// The values of the returned map are the values in XML.
func parseRecord(record string) map[string]string {
	r := strings.NewReader(record)
	parser := xml.NewDecoder(r)

	depth := 0
	content := ""

	cfield := make([]string, 0)
	outmap := make(map[string]string)

	for {
		token, err := parser.Token()
		if err != nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			elmt := xml.StartElement(t)
			name := elmt.Name.Local
			cfield = append(cfield, name)
			depth++
		case xml.EndElement:
			name := strings.Join(cfield, keysep)

			m, _ := regexp.MatchString("[[:alnum:]]", content)
			if m {
				name = safeAddKey(name, outmap)
				outmap[name] = content
			}

			cfield = cfield[: len(cfield) - 1]
			content = ""
			depth--
		case xml.CharData:
			bytes := xml.CharData(t)
			content = "\"" + string([]byte(bytes)) + "\""
		case xml.ProcInst:
			continue
		case xml.Directive:
			continue
		default:
			fmt.Println("Unknown")
		}
	}
	return outmap
}


// parseLines writes parsed XML to a file object and returns
// a header array corresponding parsed XML.
func parseLines(scanner *bufio.Scanner, writer *bufio.Writer)[]string {
	header := make([]string, 0)

	for scanner.Scan() {
		parsed := parseRecord(scanner.Text())
		output := make([]string, 0)
		for _, field := range header {
			if val, iskey := parsed[field]; iskey {
				output = append(output, val)
				delete(parsed, field)
			} else {
				output = append(output, "")
			}
		}
		
		for field, value := range parsed {
			output = append(output, value)
			header = append(header, field)
		}
		output = append(output, strconv.Itoa(len(output)) )
		writer.WriteString(strings.Join(output, delim) + "\n")
	}
	return header
}


// cleanLines reads a file written by parseLines and writes 
// the header and adds the correct number of delimiters.
func cleanLines(n int, scanner *bufio.Scanner, writer *bufio.Writer) {
	for scanner.Scan() {
		txt := scanner.Text()
		ind := strings.LastIndex(txt, ",")
		numfields_str := txt[ind + 1: ]
		txt = txt[: ind]

		numfields, _ := strconv.Atoi(numfields_str)
		k := n - numfields
		if k > 0 {
			txt += strings.Repeat(delim,k)
		}
		writer.WriteString(txt + "\n")
	}
}


// cleanHeader returns the header with item.1.1.1.1
// replaced with item.4
func cleanHeader(header []string)[]string {
	for k, field := range header {
		header[k] = oneToNum(field)
	}
	return header
}

// oneToNum sends xxxx.1.1.1.1 to xxxx.4
func oneToNum(s string) string {
	c := regexp.MustCompile("(.1)+$")
	s1 := []byte(s)
	s2 := string(c.ReplaceAllFunc(s1,addOnes))
	return s2
}

func addOnes(reg []byte) []byte  {
	N := len(bytes.Split(reg,[]byte(".")))
	return []byte("." + strconv.Itoa(N-1))
}


//
//
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "file")
		os.Exit(1)
	}
	filepath := os.Args[1]

	file, err := os.Open(filepath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
	
	tmpfile, err := os.Create("tmp_output.txt")
    if err != nil {
        panic(err)
    }
	defer tmpfile.Close()

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(tmpfile)

	header := parseLines(scanner, writer)
	writer.Flush()
	tmpfile.Close()
	
	tmpfile, err = os.Open("tmp_output.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer tmpfile.Close()
	defer os.Remove("tmp_output.txt")
	
	outfile, err := os.Create("output.txt")
    if err != nil {
        panic(err)
    }
	defer outfile.Close()

	scanner = bufio.NewScanner(tmpfile)
	writer = bufio.NewWriter(outfile)
	header = cleanHeader(header)
	writer.WriteString(strings.Join(header, delim) + "\n")

	cleanLines(len(header), scanner, writer)
	
	writer.Flush()
}



