package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
	"github.com/stretchr/testify/assert"
)

func TestWaitBeforeReturns(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	i1 := &mckmaps.Interval{Min: 100, Max: 100}
	i1Start := time.Now()
	waitBeforeReturns(i1)
	i1Elapsed := time.Since(i1Start)
	assert.True(i1Elapsed.Milliseconds() >= 100)

	i2 := &mckmaps.Interval{Min: 10, Max: 30}
	i2Start := time.Now()
	for i := 0; i < 10; i++ {
		waitBeforeReturns(i2)
	}
	i2Elapsed := time.Since(i2Start)
	assert.True(i2Elapsed.Milliseconds() >= 100 && i2Elapsed.Milliseconds() <= 350)
}

func TestPolicyExecutor_executor(t *testing.T) {
	//noinspection GoImportUsedAsName
	assert := assert.New(t)

	rr1 := httptest.NewRecorder()
	var rw1 http.ResponseWriter = rr1
	exe1 := &policyExecutor{
		r:      httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w:      &rw1,
		policy: newStatusJSONPolicy(myhttp.StatusOk, "OK"),
	}
	e1 := exe1.execute()
	assert.Nil(e1)
	assert.Equal(rr1.Code, http.StatusOK)
	assert.Equal(rr1.Body.String(), fmt.Sprintf(`{"statusCode": %d, "message": "%s"}`, http.StatusOK, "OK"))

	rr2 := httptest.NewRecorder()
	var rw2 http.ResponseWriter = rr2
	exe2 := &policyExecutor{
		r:      httptest.NewRequest("GET", "/TestPolicyExecutor_executor", nil),
		w:      &rw2,
		policy: pNoPolicyMatched,
	}
	e2 := exe2.execute()
	assert.Nil(e2)
	assert.Equal(rr2.Code, http.StatusBadRequest)
	assert.Equal(rr2.Body.String(), string(pNoPolicyMatched.Returns.Body))
}
