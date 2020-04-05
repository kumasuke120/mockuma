package mckmaps

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/myos"
)

var loadedFilenames []string

func recordLoadedFile(name string) {
	loadedFilenames = append(loadedFilenames, name)
}

type loadError struct {
	filename string
	err      error
}

func indentErrorMsg(err error) string {
	errMsg := err.Error()
	errMsg = strings.ReplaceAll(errMsg, "\n", "\n\t")
	return errMsg
}

func (e *loadError) Error() string {
	return fmt.Sprintf("cannot load the file '%s': \n\t%s",
		e.filename, indentErrorMsg(e.err))
}

type parserError struct {
	filename string
	jsonPath *myjson.Path
	err      error
}

func (e *parserError) Error() string {
	result := ""
	if e.jsonPath == nil {
		result += "cannot parse json data"
	} else {
		result += fmt.Sprintf("cannot parse the value on json-path \"%v\"", e.jsonPath)
	}

	if e.filename != "" {
		result += fmt.Sprintf(" in the file '%s'", e.filename)
	}

	if e.err != nil {
		result += ": \n\t" + indentErrorMsg(e.err)
	}

	return result
}

type Parser struct {
	filename string
}

func NewParser(filename string) *Parser {
	return &Parser{filename: filename}
}

func (p *Parser) Parse() (r *MockuMappings, e error) {
	var json interface{}
	if json, e = p.load(true, ppRemoveComment, ppRenderTemplate); e != nil {
		return
	}

	switch json.(type) {
	case myjson.Object: // parses in multi-file mode
		parser := &mainParser{json: json.(myjson.Object), Parser: *p}
		r, e = parser.parse()
	case myjson.Array: // parses in single-file mode
		parser := &mappingsParser{json: json, Parser: *p}
		mappings, _err := parser.parse()
		if _err == nil {
			r, e = &MockuMappings{Mappings: mappings}, _err
		} else {
			r, e = nil, _err
		}
	default:
		r, e = nil, newParserError(p.filename, nil)
	}

	if r != nil {
		relPaths, err := p.allRelative(loadedFilenames)
		if err != nil {
			return nil, err
		}
		r.Filenames = relPaths
	}

	p.reset()
	r = p.sortMappings(r)
	return
}

func (p *Parser) load(record bool, preprocessors ...filter) (interface{}, error) {
	if err := checkFilepath(p.filename); err != nil {
		return nil, &loadError{filename: p.filename, err: err}
	}

	bytes, err := ioutil.ReadFile(p.filename)
	if err != nil {
		return nil, err
	}

	json, err := myjson.Unmarshal(bytes)
	if err != nil {
		return nil, newParserError(p.filename, nil)
	}

	v, err := doFiltersOnV(json, preprocessors...) // runs given preprocessors
	if err != nil {
		return nil, &loadError{filename: p.filename, err: err}
	}

	if record {
		recordLoadedFile(p.filename)
	}
	return v, nil
}

func (p *Parser) allRelative(filenames []string) (ret []string, err error) {
	wd := myos.GetWd()

	ret = make([]string, len(filenames))
	for i, p := range filenames {
		rp := p
		if filepath.IsAbs(p) {
			rp, err = filepath.Rel(wd, p)
			if err != nil {
				ret = nil
				return
			}
		}

		ret[i] = rp
	}
	return
}

func (p *Parser) reset() {
	ppRenderTemplate.reset()
	ppLoadFile.reset()
	ppParseRegexp.reset()

	loadedFilenames = nil
}

func (p *Parser) sortMappings(mappings *MockuMappings) *MockuMappings {
	if mappings == nil {
		return nil
	}

	uri2mappings := make(map[string][]*Mapping)

	var uriOrder []string
	uriOrderContains := make(map[string]bool)
	for _, m := range mappings.Mappings {
		mappingsOfURI := uri2mappings[m.URI]

		mappingsOfURI = appendToMappingsOfURI(mappingsOfURI, m)
		uri2mappings[m.URI] = mappingsOfURI
		if _, ok := uriOrderContains[m.URI]; !ok {
			uriOrderContains[m.URI] = true
			uriOrder = append(uriOrder, m.URI)
		}
	}

	ms := make([]*Mapping, 0, len(mappings.Mappings))
	for _, uri := range uriOrder {
		mappingsOfURI := uri2mappings[uri]
		ms = append(ms, mappingsOfURI...)
	}
	return &MockuMappings{Mappings: ms, Filenames: mappings.Filenames}
}

func appendToMappingsOfURI(dst []*Mapping, m *Mapping) []*Mapping {
	merged := false
	for _, dm := range dst {
		if dm.URI == m.URI && dm.Method == m.Method {
			dm.Policies = append(dm.Policies, m.Policies...)
			merged = true
		}
	}

	if !merged {
		dst = append(dst, m)
	}
	return dst
}

type mainParser struct {
	json myjson.Object
	Parser
}

func (p *mainParser) parse() (*MockuMappings, error) {
	_type, err := p.json.GetString(aType)
	if err != nil || string(_type) != tMain {
		return nil, newParserError(p.filename, myjson.NewPath(aType))
	}

	include, err := p.json.GetObject(aInclude)
	if err != nil {
		return nil, newParserError(p.filename, myjson.NewPath(aInclude))
	}

	filenamesOfMappings, err := include.GetArray(tMappings)
	if err != nil {
		return nil, newParserError(p.filename, myjson.NewPath(aInclude, tMappings))
	}

	var mappings []*Mapping
	for idx, filename := range filenamesOfMappings {
		_filename, err := myjson.ToString(filename)
		if err != nil {
			return nil, newParserError(p.filename, myjson.NewPath(aInclude, tMappings, idx))
		}

		f := string(_filename)
		glob, err := filepath.Glob(f)
		if err != nil {
			return nil, newParserError(p.filename, myjson.NewPath(aInclude, tMappings, idx))
		}

		for _, g := range glob {
			parser := &mappingsParser{Parser: Parser{filename: g}}
			partOfMappings, err := parser.parse() // parses mappings for each included file
			if err != nil {
				return nil, err
			}

			mappings = append(mappings, partOfMappings...)
		}

		recordLoadedFile(f)
	}

	return &MockuMappings{Mappings: mappings}, nil
}

func newParserError(filename string, jsonPath *myjson.Path) *parserError {
	return &parserError{filename: filename, jsonPath: jsonPath}
}

func ensureJSONArray(v interface{}) myjson.Array {
	switch v.(type) {
	case myjson.Array:
		return v.(myjson.Array)
	default:
		return myjson.NewArray(v)
	}
}

func checkFilepath(path string) (err error) {
	wd := myos.GetWd()

	relPath := path
	if filepath.IsAbs(path) {
		relPath, err = filepath.Rel(wd, path)
		if err != nil {
			return
		}
	}

	if strings.HasPrefix(relPath, "..") { // paths should be under the current working directory
		return errors.New("included file isn't in the current working directory")
	}
	return
}
