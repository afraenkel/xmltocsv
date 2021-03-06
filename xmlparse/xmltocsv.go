package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// FileProcessSpecs contains information on reading/writing
// file buffers for processing.
type FileProcessSpecs struct {
	inpath  string
	outpath string

	outpathtmp bool
}

// ParseInfo contains parsing info for the csv formatting
type ParseInfo struct {
	delim  string
	keysep string
}

// Defaults for processing
var filespec FileProcessSpecs
var parse = ParseInfo{delim: ",", keysep: "."}

// safeAddKey checks if a string key is a key of a map m.
// If it is a key, then safeAddKey adds as many '.1' suffixes
// to key until it's not a key of map m, then returns that string.
// Otherwise, safeAddKey returns the original input key.
func safeAddKey(key string, m map[string]string) string {
	for {
		if _, iskey := m[key]; iskey {
			key = key + parse.keysep + "1"
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
			name := strings.Join(cfield, parse.keysep)

			m, _ := regexp.MatchString("[[:alnum:]]", content)
			if m {
				name = safeAddKey(name, outmap)
				outmap[name] = content
			}

			cfield = cfield[:len(cfield)-1]
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
func parseLines(scanner *bufio.Scanner, writer *bufio.Writer) []string {
	defer writer.Flush()
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
		output = append(output, strconv.Itoa(len(output)))
		writer.WriteString(strings.Join(output, parse.delim) + "\n")
	}
	return header
}

// cleanLines reads a file written by parseLines and writes
// the header and adds the correct number of delimiters.
func cleanLines(n int, scanner *bufio.Scanner, writer *bufio.Writer) {
	if n == 0 {
		txt := scanner.Text()
		writer.WriteString(txt)
	} else {
		for scanner.Scan() {
			txt := scanner.Text()
			ind := strings.LastIndex(txt, parse.delim)
			numfields_str := txt[ind+1:]
			txt = txt[:ind]

			numfields, _ := strconv.Atoi(numfields_str)
			k := n - numfields
			if k > 0 {
				txt += strings.Repeat(parse.delim, k)
			}
			writer.WriteString(txt + "\n")
		}
	}
}

// cleanHeader returns the header with item.1.1.1.1
// replaced with item.4  (where "." is the keysep)
func cleanHeader(header []string) []string {
	for k, field := range header {
		header[k] = oneToNum(field)
	}
	return header
}

// oneToNum sends xxxx.1.1.1.1 to xxxx.4
func oneToNum(s string) string {
	addOnes := func(reg []byte) []byte {
		N := len(bytes.Split(reg, []byte(parse.keysep)))
		return []byte(parse.keysep + strconv.Itoa(N-1))
	}

	regex := strings.Replace("(%s1)+$", "%s", parse.keysep, 1)
	c := regexp.MustCompile(regex)
	s1 := []byte(s)
	s2 := string(c.ReplaceAllFunc(s1, addOnes))
	return s2
}

// openToProcess is a convenience function for reading from one file and
// writing to another file.  You must close/flush the output.
func openToProcess(filespec *FileProcessSpecs) (*os.File, *os.File, func()) {

	files := map[string]*os.File{
		"infile":  os.Stdin,
		"outfile": os.Stdout,
	}

	if filespec.inpath != "" {
		infile, err := os.Open(filespec.inpath)
		files["infile"] = infile
		if err != nil {
			log.Fatal(err)
		}
	}

	if filespec.outpathtmp {
		outfile, err := ioutil.TempFile("./", filespec.outpath)
		files["outfile"] = outfile
		if err != nil {
			log.Fatal(err)
		}
		filespec.outpath = outfile.Name()

	} else if filespec.outpath != "" {
		outfile, err := os.Create(filespec.outpath)
		files["outfile"] = outfile
		if err != nil {
			log.Fatal(err)
		}
	}

	closeFunc := func() {
		files["infile"].Close()
		files["outfile"].Close()
	}
	return files["infile"], files["outfile"], closeFunc
}

func processToTemp(filespec *FileProcessSpecs) ([]string, func()) {
	infile, outfile, closeFunc := openToProcess(filespec)
	defer closeFunc()
	scanner := bufio.NewScanner(infile)
	writer := bufio.NewWriter(outfile)
	defer writer.Flush()

	return parseLines(scanner, writer), func() { os.Remove(outfile.Name()) }
}

//
func processToFinal(filespec *FileProcessSpecs, header []string) {
	infile, outfile, closeFunc := openToProcess(filespec)
	defer closeFunc()

	scanner := bufio.NewScanner(infile)
	writer := bufio.NewWriter(outfile)
	defer writer.Flush()

	header = cleanHeader(header)
	writer.WriteString(strings.Join(header, parse.delim) + "\n")

	cleanLines(len(header), scanner, writer)
}

//
func main() {
	flag.StringVar(&parse.delim, "d", ",", "Output delimiter")
	flag.StringVar(&parse.keysep, "k", ".", "Output header key sepearator")
	argsIn := flag.String("i", "", "Input file path -- defaults to Stdin")
	argsOut := flag.String("o", "", "Output file path -- defaults to Stdout")

	flag.Parse()

	filespec.inpath = *argsIn
	filespec.outpathtmp = true

	header, rmtemp := processToTemp(&filespec)
	defer rmtemp()

	filespec.inpath = filespec.outpath
	filespec.outpath = *argsOut
	filespec.outpathtmp = false

	processToFinal(&filespec, header)

}
