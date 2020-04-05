package mckmaps

// types
const (
	tMain     = "main"
	tMappings = "mappings"
	tTemplate = "template"
	tVars     = "vars"
)

// directives
const (
	dFile     = "@file"
	dComment  = "@comment"
	dTemplate = "@template"
	dVars     = "@vars"
	dRegexp   = "@regexp"
	dJSON     = "@json"
)

// attributes
const (
	aType    = "type"
	aInclude = "include"

	aMapURI      = "uri"
	aMapMethod   = "method"
	aMapPolicies = "policies"
)

const (
	mapPolicyWhen      = "when"
	mapPolicyReturns   = "returns"
	mapPolicyForwards  = "forwards"
	mapPolicyRedirects = "redirects"
)

// commands of mappings policies
var mapPolicyCommands = []string{mapPolicyReturns, mapPolicyForwards, mapPolicyRedirects}

// attributes for mappings policies
const (
	pStatusCode = "statusCode"
	pHeaders    = "headers"
	pParams     = "params"
	pPathVars   = "pathVars"
	pBody       = "body"
	pLatency    = "latency"
	pPath       = "path"
)
