package server

import (
	"fmt"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
)

// predefined policies
var (
	pNotFound            = newJSONPolicy(myhttp.StatusNotFound, "Not Found")
	pNoPolicyMatched     = newJSONPolicy(myhttp.StatusBadRequest, "No policy matched")
	pMethodNotAllowed    = newJSONPolicy(myhttp.StatusMethodNotAllowed, "Method Not Allowed")
	pInternalServerError = newJSONPolicy(myhttp.StatusInternalServerError, "Internal Server Error")
	pBadGateway          = newJSONPolicy(myhttp.StatusBadGateway, "Bad Gateway")
	pEmptyOK             = &mckmaps.Policy{
		CmdType: mckmaps.CmdTypeReturns,
		Returns: &mckmaps.Returns{StatusCode: myhttp.StatusOK},
	}
)

func newJSONPolicy(statusCode myhttp.StatusCode, message string) *mckmaps.Policy {
	return &mckmaps.Policy{
		CmdType: mckmaps.CmdTypeReturns,
		Returns: &mckmaps.Returns{
			StatusCode: statusCode,
			Headers: []*mckmaps.NameValuesPair{
				{Name: myhttp.HeaderContentType, Values: []string{myhttp.ContentTypeJSON}},
			},
			Body: []byte(fmt.Sprintf(`{"statusCode": %d, "message": "%s"}`, statusCode, message)),
		},
	}
}

func newForwardPolicy(uri string) *mckmaps.Policy {
	return &mckmaps.Policy{
		CmdType: mckmaps.CmdTypeForwards,
		Forwards: &mckmaps.Forwards{
			Path: uri,
		},
	}
}
