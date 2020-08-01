package mckmaps

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/kumasuke120/mockuma/internal/myjson"
	"github.com/kumasuke120/mockuma/internal/myos"
	"github.com/kumasuke120/mockuma/internal/types"
	"github.com/rs/cors"
)

type MockuMappings struct {
	Mappings  []*Mapping
	Filenames []string
	Config    *Config
}

func (m *MockuMappings) IsEmpty() bool {
	return len(m.Mappings) == 0 && len(m.Filenames) == 0
}

func (m *MockuMappings) GroupMethodsByURI() map[string][]myhttp.HTTPMethod {
	result := make(map[string][]myhttp.HTTPMethod)
	for _, m := range m.Mappings {
		mappingsOfURI := result[m.URI]
		mappingsOfURI = append(mappingsOfURI, m.Method)
		result[m.URI] = mappingsOfURI
	}
	return result
}

type Config struct {
	CORS               *CORSOptions
	MatchTrailingSlash bool
}

func defaultConfig() *Config {
	return &Config{
		CORS:               defaultDisabledCORS(),
		MatchTrailingSlash: false,
	}
}

type CORSOptions struct {
	Enabled          bool
	AllowCredentials bool
	MaxAge           int64
	AllowedOrigins   []string
	AllowedMethods   []myhttp.HTTPMethod
	AllowedHeaders   []string
	ExposedHeaders   []string
}

func defaultEnabledCORS() *CORSOptions {
	return &CORSOptions{
		Enabled:          true,
		AllowCredentials: true,
		MaxAge:           1800,
		AllowedOrigins:   []string{"*"},
		AllowedMethods: []myhttp.HTTPMethod{
			myhttp.MethodGet,
			myhttp.MethodPost,
			myhttp.MethodHead,
			myhttp.MethodOptions,
		},
		AllowedHeaders: []string{
			myhttp.HeaderOrigin,
			myhttp.HeaderAccept,
			myhttp.HeaderXRequestWith,
			myhttp.HeaderContentType,
			myhttp.HeaderAccessControlRequestMethod,
			myhttp.HeaderAccessControlRequestHeaders,
		},
		ExposedHeaders: nil,
	}
}

func defaultDisabledCORS() *CORSOptions {
	return &CORSOptions{Enabled: false}
}

var anyStrToTrue = func(string) bool { return true }

func (c *CORSOptions) ToCors() *cors.Cors {
	if c.Enabled {
		ac := c.AllowedMethods
		if !myhttp.MethodsAnyMatches(ac, myhttp.MethodOptions) { // always allows OPTIONS
			ac = append(ac, myhttp.MethodOptions)
		}

		// makes github.com/rs/cors returns the Origin of a request
		// as the value of response header Access-Control-Allow-Origin
		// when Access-Control-Allow-Credentials is 'true' and all
		// origins are allowed
		var ao []string
		var aof func(string) bool
		if c.allowsAllOrigins() {
			if c.AllowCredentials {
				ao = nil
				aof = anyStrToTrue
			} else {
				ao = []string{"*"}
				aof = nil
			}
		}

		return cors.New(cors.Options{
			AllowCredentials: c.AllowCredentials,
			MaxAge:           int(c.MaxAge),
			AllowedOrigins:   ao,
			AllowOriginFunc:  aof,
			AllowedMethods:   myhttp.MethodsToStringSlice(ac),
			AllowedHeaders:   c.AllowedHeaders,
			ExposedHeaders:   c.ExposedHeaders,
		})
	}

	return nil
}

func (c *CORSOptions) allowsAllOrigins() bool {
	if len(c.AllowedOrigins) == 0 {
		return true
	}
	for _, o := range c.AllowedOrigins {
		if "*" == o { // allows all origins
			return true
		}
	}
	return false
}

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

// preprocessors singletons
var (
	ppRemoveComment  = &dCommentProcessor{}
	ppRenderTemplate = makeDTemplateProcessor()
	ppLoadFile       = makeDFileProcessor()
	ppParseRegexp    = makeDRegexpProcessor()
	ppToJSONMatcher  = &dJSONProcessor{}
)

type Parser struct {
	filename string
}

func NewParser(filename string) *Parser {
	return &Parser{filename: filename}
}

func (p *Parser) newJSONParseError(jsonPath *myjson.Path) *parserError {
	return &parserError{filename: p.filename, jsonPath: jsonPath}
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
			r, e = &MockuMappings{Mappings: mappings, Config: defaultConfig()}, _err
		} else {
			r, e = nil, _err
		}
	default:
		r, e = nil, p.newJSONParseError(nil)
	}

	if r != nil {
		relPaths, err := p.allRelative(loadedFilenames)
		if err != nil {
			return nil, err
		}
		r.Filenames = relPaths
	}

	p.reset()
	p.sortMappings(r)
	return
}

func (p *Parser) load(record bool, preprocessors ...types.Filter) (interface{}, error) {
	if err := checkFilepath(p.filename); err != nil {
		return nil, &loadError{filename: p.filename, err: err}
	}

	bytes, err := ioutil.ReadFile(p.filename)
	if err != nil {
		return nil, err
	}

	json, err := myjson.Unmarshal(bytes)
	if err != nil {
		return nil, p.newJSONParseError(nil)
	}

	v, err := types.DoFiltersOnV(json, preprocessors...) // runs given preprocessors
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

func (p *Parser) sortMappings(mappings *MockuMappings) {
	if mappings == nil {
		return
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

	mappings.Mappings = ms
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
	json     myjson.Object
	jsonPath *myjson.Path
	Parser
}

func (p *mainParser) parse() (*MockuMappings, error) {
	p.jsonPath = myjson.NewPath("")

	p.jsonPath.SetLast(aType)
	_type, err := p.json.GetString(aType)
	if err != nil || string(_type) != tMain {
		return nil, p.newJSONParseError(p.jsonPath)
	}

	p.jsonPath.SetLast(aInclude)
	mappings, err := p.parseInclude(err)
	if err != nil {
		return nil, err
	}

	p.jsonPath.SetLast(aConfig)
	rawConf := p.json.Get(aConfig)
	cc, err := p.parseConfig(rawConf)
	if err != nil {
		return nil, err
	}

	return &MockuMappings{Mappings: mappings, Config: cc}, nil
}

func (p *mainParser) parseInclude(err error) ([]*Mapping, error) {
	include, err := p.json.GetObject(aInclude)
	if err != nil {
		return nil, p.newJSONParseError(p.jsonPath)
	}

	p.jsonPath.Append(tMappings)
	filenamesOfMappings, err := include.GetArray(tMappings)
	if err != nil {
		return nil, p.newJSONParseError(p.jsonPath)
	}

	p.jsonPath.Append("")
	var mappings []*Mapping
	for idx, filename := range filenamesOfMappings {
		p.jsonPath.SetLast(idx)

		_filename, err := myjson.ToString(filename)
		if err != nil {
			return nil, p.newJSONParseError(p.jsonPath)
		}

		f := string(_filename)
		glob, err := filepath.Glob(f)
		if err != nil {
			return nil, p.newJSONParseError(myjson.NewPath(aInclude, tMappings, idx))
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
	p.jsonPath.RemoveLast()
	p.jsonPath.RemoveLast()

	return mappings, nil
}

func (p *mainParser) parseConfig(v interface{}) (c *Config, err error) {
	switch v.(type) {
	case nil:
		c = defaultConfig()
	case myjson.Object:
		vo := v.(myjson.Object)
		p.jsonPath.Append("")

		p.jsonPath.SetLast(aConfigCORS)
		var co *CORSOptions
		co, err = p.parseCORSOptions(vo)
		if err != nil {
			return
		}

		p.jsonPath.SetLast(aConfigMatchTrailingSlash)
		var mts bool
		if vo.Has(aConfigMatchTrailingSlash) {
			_mts, _err := vo.GetBoolean(aConfigMatchTrailingSlash)
			if _err != nil {
				err = p.newJSONParseError(p.jsonPath)
				return
			}
			mts = bool(_mts)
		} else {
			mts = false
		}

		c = &Config{CORS: co, MatchTrailingSlash: mts}
		p.jsonPath.RemoveLast()
	default:
		return nil, p.newJSONParseError(p.jsonPath)
	}

	return
}

func (p *mainParser) parseCORSOptions(v myjson.Object) (co *CORSOptions, err error) {
	_co := v.Get(aConfigCORS)
	switch _co.(type) {
	case nil:
		co = defaultDisabledCORS()
	case myjson.Boolean:
		if _co.(myjson.Boolean) {
			co = defaultEnabledCORS()
		} else {
			co = defaultDisabledCORS()
		}
	case myjson.Object:
		_corsV := _co.(myjson.Object)
		p.jsonPath.Append("")
		co, err = p.parseActualCORSOptions(_corsV)
		if err != nil {
			return
		}
		p.jsonPath.RemoveLast()
	default:
		err = p.newJSONParseError(p.jsonPath)
		return
	}
	return
}

func (p *mainParser) parseActualCORSOptions(v myjson.Object) (*CORSOptions, error) {
	p.jsonPath.SetLast(corsEnabled)
	enabled, err := v.GetBoolean(corsEnabled)
	if err != nil {
		return nil, p.newJSONParseError(p.jsonPath)
	}

	if enabled {
		cc := defaultEnabledCORS()

		p.jsonPath.SetLast(corsAllowCredentials)
		ac, err := v.GetBoolean(corsEnabled)
		if err != nil {
			return nil, p.newJSONParseError(p.jsonPath)
		}
		cc.AllowCredentials = bool(ac)

		if v.Has(corsMaxAge) {
			p.jsonPath.SetLast(corsMaxAge)
			ma, err := p.json.GetNumber(corsMaxAge)
			if err != nil {
				return nil, p.newJSONParseError(p.jsonPath)
			}
			cc.MaxAge = int64(ma)
		}

		if v.Has(corsAllowedOrigins) {
			p.jsonPath.SetLast(corsAllowedOrigins)
			ao, err := p.getAsStringSlice(v, corsAllowedOrigins)
			if err != nil {
				return nil, err
			}
			cc.AllowedOrigins = ao
		}

		if v.Has(corsAllowedMethods) {
			p.jsonPath.SetLast(corsAllowedMethods)
			_am, err := p.getAsStringSlice(v, corsAllowedMethods)
			if err != nil {
				return nil, err
			}
			am := make([]myhttp.HTTPMethod, len(_am))
			for idx, v := range _am {
				am[idx] = myhttp.ToHTTPMethod(v)
			}
			cc.AllowedMethods = am
		}

		if v.Has(corsAllowedHeaders) {
			p.jsonPath.SetLast(corsAllowedHeaders)
			ah, err := p.getAsStringSlice(v, corsAllowedHeaders)
			if err != nil {
				return nil, err
			}
			cc.AllowedHeaders = ah
		}

		if v.Has(corsExposedHeaders) {
			p.jsonPath.SetLast(corsExposedHeaders)
			eh, err := p.getAsStringSlice(v, corsExposedHeaders)
			if err != nil {
				return nil, err
			}
			cc.ExposedHeaders = eh
		}

		return cc, nil
	}

	return defaultDisabledCORS(), nil
}

func (p *mainParser) getAsStringSlice(v myjson.Object, name string) ([]string, error) {
	p.jsonPath.Append("")

	var result []string
	for idx, e := range ensureJSONArray(v.Get(name)) {
		p.jsonPath.SetLast(idx)

		s, err := myjson.ToString(e)
		if err != nil {
			return nil, p.newJSONParseError(p.jsonPath)
		}
		result = append(result, string(s))
	}
	p.jsonPath.RemoveLast()

	return result, nil
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
