package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	// "log"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

type user struct {
	Name     string
	Email    string
	Browsers []string
}

func easyjson9e1087fdDecodeFakeCom(in *jlexer.Lexer, out *user) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 100)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *user) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeFakeCom(&r, v)
	return r.Error()
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintln(out, "found users:")

	scanner := bufio.NewScanner(file)
	user := user{}
	uniqueBrowsers := make(map[string]int)

	hasIE := false
	hasAndroid := false

	idx := 0
	for scanner.Scan() {
		hasIE = false
		hasAndroid = false

		if err := user.UnmarshalJSON(scanner.Bytes()); err != nil {
			panic("failed to unmarshal")
		}

	LOOP:
		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				hasAndroid = true
				uniqueBrowsers[browser] = 0
			}
			if strings.Contains(browser, "MSIE") {
				hasIE = true
				uniqueBrowsers[browser] = 0
			}
			if hasIE && hasAndroid {
				break LOOP
			}
		}

		if hasIE && hasAndroid {
			email := strings.Replace(user.Email, "@", " [at] ", 100)
			fmt.Fprintf(out, "[%d] %s <%s>\n", idx, user.Name, email)
		}

		idx++
	}

	if err := scanner.Err(); err != nil {
		panic("error reading file")
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(uniqueBrowsers))
}
