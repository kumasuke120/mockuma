package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockHandler_ServeHTTP(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	handler := newMockHandler(mappings)

	req1 := httptest.NewRequest("POST", "/hello", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req1)

	assert.Equal(http.StatusOK, rr.Code)
}
