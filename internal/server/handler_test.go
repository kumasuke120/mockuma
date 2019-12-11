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
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(http.StatusOK, rr1.Code)

	req2 := httptest.NewRequest("", "/hello", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(http.StatusMethodNotAllowed, rr2.Code)

	req3 := httptest.NewRequest("", "/notfound", nil)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	assert.Equal(http.StatusNotFound, rr3.Code)

	req4 := httptest.NewRequest("", "/m", nil)
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)
	assert.Equal(http.StatusBadRequest, rr4.Code)
}

func TestMockHandler_listAllMappings(t *testing.T) {
	handler := newMockHandler(mappings)
	handler.listAllMappings()
}
