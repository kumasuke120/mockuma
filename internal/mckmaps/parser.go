package mckmaps

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/kumasuke120/mockuma/internal/myjson"
)

const (
	tMain     = "main"
	tMappings = "mappings"
)

const (
	dType    = "@type"
	dInclude = "@include"
)

type JsonParseError struct {
	JsonPath string
}

func (e *JsonParseError) Error() string {
	if e.JsonPath == "" {
		return "Cannot parse json data"
	} else {
		return fmt.Sprintf("Cannot parse value on json-path '%s", e.JsonPath)
	}
}

type ParserError struct {
	Filename string
	JsonPath string
}

func (e *ParserError) Error() string {
	result := ""
	if e.JsonPath == "" {
		result += "Cannot parse json data"
	} else {
		result += fmt.Sprintf("Cannot parse value on json-path '%s'", e.JsonPath)
	}

	if e.Filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.Filename)
	}

	return result
}

func newParserError(filename string, jsonPath string) *ParserError {
	return &ParserError{Filename: filename, JsonPath: jsonPath}
}

type Parser struct {
	filename string
	chdir    bool
}

type mainParser struct {
	json myjson.Object
	Parser
}

type mappingsParser struct {
	Parser
}

func (p *Parser) Parse() (*MockuMappings, error) {
	bytes, err := ioutil.ReadFile(p.filename)
	if err != nil {
		return nil, err
	}

	json, err := myjson.Unmarshal(bytes)
	if err != nil {
		return nil, newParserError(p.filename, "")
	}

	if p.chdir {
		p.chdirBasedOnFilename()
	}

	switch json.(type) {
	case myjson.Object:
		parser := &mainParser{Parser: *p, json: json.(myjson.Object)}
		return parser.Parse()
	}

	return nil, newParserError(p.filename, "")
}

func (p *Parser) chdirBasedOnFilename() {
	abs, err := filepath.Abs(p.filename)
	if err != nil {
		log.Fatal("Cannot acquire the absolute path of mockuMappings:", err)
	}

	dir := filepath.Dir(abs)
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal("Cannot change the working directory:", err)
	}

	log.Println("[loader] working directory has been changed to:", dir)
}

func (p *mainParser) Parse() (*MockuMappings, error) {
	_type, err := p.json.GetString(dType)
	if err != nil || _type != tMain {
		return nil, newParserError(p.filename, "$."+dType)
	}

	include, err := p.json.GetObject(dInclude)
	if err != nil {
		return nil, newParserError(p.filename, "$."+dInclude)
	}

	filenamesOfMappings, err := include.GetArray(tMappings)
	if err != nil {
		return nil, newParserError(p.filename, fmt.Sprintf("$.%s.%s", dInclude, tMappings))
	}

	var mappings = make([]*Mapping, len(filenamesOfMappings))
	for idx, filename := range filenamesOfMappings {
		_filename, err := myjson.ToString(filename)
		if err != nil {
			return nil, newParserError(p.filename, fmt.Sprintf("$.%s.%s[%d]", dInclude, tMappings, idx))
		}

		parser := &mappingsParser{Parser{filename: string(_filename)}}
		mapping, err := parser.Parse()
		if err != nil {
			return nil, err
		}

		mappings = append(mappings, mapping)
	}

	return &MockuMappings{Mappings: mappings}, nil
}

func (p *mappingsParser) Parse() (*Mapping, error) {
	// TODO: Not implemented yet
	panic("Not implemented yet")
}
