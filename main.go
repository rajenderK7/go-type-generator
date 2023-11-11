package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

const generatedFilename = "gen.txt"

func makePublicMember(key string) string {
	if len(key) == 0 {
		return key
	}
	// Since strings are immutable in Go we cannot directly
	// modify the characters of a string. Instead we first convert
	// them to a slice of runes ([]int32) do the modification and
	// then convert back to a string
	runes := []rune(key)
	// Capitalize the first character of the key
	// to make it a public member.
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// GenerateGoStruct generates a valid Go struct type based on the input JSON
// and writes to the "gen.txt" file
func GenerateGoStruct(jsonData map[string]interface{}, sb *strings.Builder, tabSpaces int) {
	sb.WriteString("struct {")
	sb.WriteByte('\n')
	for key, val := range jsonData {
		// Every new line is indented with tab space(s).
		for i := 0; i < tabSpaces; i++ {
			sb.WriteByte('\t')
		}
		sb.WriteString(makePublicMember(key))
		sb.WriteByte('\t')
		var valType string
		// This check is for nested JSON objects.
		if nestedJSON, ok := val.(map[string]interface{}); ok {
			GenerateGoStruct(nestedJSON, sb, tabSpaces+1)
			sb.WriteString(fmt.Sprintf("\t`json:\"%s\"`\n", key))
			continue
		}
		valType = fmt.Sprintf("%T", val)
		// TODO: Appropriate handling of "null" from JSON is required.
		if valType == "<nil>" {
			valType = "bool"
		}
		// It is for now expected that the JSON contains only
		// homogenous arrays i.e. [1, 2, 3] or ["a", "b", "c"]
		// and not ["a", 1.32, "b"].
		if array, ok := val.([]interface{}); ok {
			// This check expects the array to contain homogenous nested
			// JSON objects.
			if nestedJSON, ok := array[0].(map[string]interface{}); ok {
				sb.WriteString("[]")
				GenerateGoStruct(nestedJSON, sb, tabSpaces+1)
				sb.WriteString(fmt.Sprintf("\t`json:\"%s\"`\n", key))
				continue
			}
			valType = fmt.Sprintf("[]%T", array[0])
		}
		sb.WriteString(valType)
		sb.WriteString(fmt.Sprintf("\t`json:\"%s\"`", key))
		sb.WriteByte('\n')
	}
	for i := 0; i < tabSpaces-1; i++ {
		sb.WriteByte('\t')
	}
	sb.WriteByte('}')
}

func main() {
	f, err := os.Open("f.json")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		panic(err.Error())
	}

	var data map[string]interface{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		panic(err.Error())
	}

	sb := strings.Builder{}
	sb.WriteString("type Generated ")
	GenerateGoStruct(data, &sb, 1)

	// Write the generated data to a file.
	file, err := os.Create(generatedFilename)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	_, err = file.WriteString(strings.TrimRight(sb.String(), "\n"))
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Open %s for the generated Go struct", generatedFilename)
}
