package mckmaps

const (
	tMain     = "main"
	tMappings = "mappings"
	tTemplate = "template"
	tVars     = "vars"
)

const (
	dType     = "@type"
	dInclude  = "@include"
	dFile     = "@file"
	dComment  = "@comment"
	dTemplate = "@template"
	dVars     = "@vars"
	dRegexp   = "@regexp"
	dJSON     = "@json"
)

const (
	mapURI             = "uri"
	mapMethod          = "method"
	mapPolicies        = "policies"
	mapPolicyWhen      = "when"
	mapPolicyReturns   = "returns"
	mapPolicyForwards  = "forwards"
	mapPolicyRedirects = "redirects"
)

var mapPolicyCommands = []string{mapPolicyReturns, mapPolicyForwards, mapPolicyRedirects}

const (
	pStatusCode = "statusCode"
	pHeaders    = "headers"
	pParams     = "params"
	pPathVars   = "pathVars"
	pBody       = "body"
	pLatency    = "latency"
	pPath       = "path"
)
