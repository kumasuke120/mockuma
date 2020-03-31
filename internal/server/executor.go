package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
	"github.com/kumasuke120/mockuma/internal/myhttp"
)

type policyExecutor struct {
	h      *mockHandler
	r      *http.Request
	w      *http.ResponseWriter
	policy *mckmaps.Policy

	fromForwards bool
	statusCode   int
}

type forwardError struct {
	err error
}

func (e *forwardError) Error() string {
	return "fail to forward: " + e.err.Error()
}

func (e *policyExecutor) execute() error {
	cmdType := e.policy.CmdType
	switch cmdType {
	case mckmaps.CmdTypeReturns:
		fallthrough
	case mckmaps.CmdTypeRedirects:
		return e.executeReturns()
	case mckmaps.CmdTypeForwards:
		return e.executeForwards()
	}

	log.Printf("[executor] %-9s: unsupported command type\n", cmdType)
	return errors.New("unsupported command type: " + string(cmdType))
}

func (e *policyExecutor) executeReturns() error {
	returns := e.policy.Returns

	if returns.Latency != nil {
		waitBeforeReturns(returns.Latency)
	}

	err := e.writeResponseForReturns(returns)
	if err != nil {
		return err
	}

	e.statusCode = int(returns.StatusCode) // records statusCode for forwards
	if !e.fromForwards {                   // forwards prints its log by itself
		log.Printf("[executor] %-9s: (%d) %s %s\n", e.policy.CmdType,
			e.statusCode, e.r.Method, e.r.URL)
	}
	return nil
}

func (e *policyExecutor) writeResponseForReturns(returns *mckmaps.Returns) error {
	e.writeHeaders(returns.Headers)
	(*e.w).WriteHeader(int(returns.StatusCode)) // statusCode must be written after headers
	err := e.writeBody(returns.Body)
	if err != nil {
		return err
	}
	return nil
}

func (e *policyExecutor) writeHeaders(headers []*mckmaps.NameValuesPair) {
	outHeader := (*e.w).Header()

	// new headers overrides old ones
	for _, pair := range headers {
		if _, ok := outHeader[pair.Name]; ok {
			outHeader.Del(pair.Name)
		}

		for _, value := range pair.Values {
			outHeader.Add(pair.Name, value)
		}
	}
}

func (e *policyExecutor) writeBody(body []byte) error {
	var err error
	if body != nil && len(body) != 0 {
		_, err = (*e.w).Write(body)
	}
	return err
}

func (e *policyExecutor) executeForwards() error {
	forwards := e.policy.Forwards

	if forwards.Latency != nil {
		waitBeforeReturns(forwards.Latency)
	}

	fPath := forwards.Path
	if strings.HasPrefix(fPath, "http://") || strings.HasPrefix(fPath, "https://") {
		return e.forwardsRemote(fPath)
	} else {
		return e.forwardsLocal(fPath)
	}
}

func (e *policyExecutor) forwardsRemote(fPath string) error {
	newRequest, err := e.newForwardRemoteRequest(fPath)
	if err != nil {
		return err
	}

	httpClient := http.Client{}
	resp, err := httpClient.Do(newRequest)
	if err != nil {
		return &forwardError{err: err}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("[executor] error    : error encountered when forwarding: " + err.Error())
		}
	}()

	err = e.writeResponseForForwardsRemote(resp)
	if err != nil {
		return err
	}

	e.statusCode = resp.StatusCode // records statusCode for forwards
	log.Printf("[executor] %-9s: (%d) %s %s => %s\n", e.policy.CmdType,
		resp.StatusCode, e.r.Method, e.r.URL, newRequest.URL)
	return nil
}

func (e *policyExecutor) newForwardRemoteRequest(fPath string) (*http.Request, error) {
	reqURL := e.r.URL

	_url, err := url.Parse(fPath)
	if err != nil {
		return nil, &forwardError{err: err}
	}
	_url.RawQuery = reqURL.RawQuery

	newRequest, err := e.newForwardRequest(_url.String())
	if err != nil {
		return nil, err
	}
	newRequest.Header.Set(myhttp.HeaderXForwardedFor, e.r.RemoteAddr)

	return newRequest, err
}

func (e *policyExecutor) writeResponseForForwardsRemote(resp *http.Response) error {
	for key, values := range resp.Header {
		(*e.w).Header()[key] = values
	}
	(*e.w).Header().Set(myhttp.HeaderXForwardedServer, HeaderValueServer)
	(*e.w).WriteHeader(resp.StatusCode) // statusCode must be written after headers
	_, err := io.Copy(*e.w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (e *policyExecutor) forwardsLocal(fPath string) error {
	newRequest, err := e.newForwardLocalRequest(fPath)
	if err != nil {
		return err
	}

	fe := e.h.matchNewExecutor(newRequest, *e.w)
	fe.fromForwards = true
	err = fe.execute() // executor writes response for forwards

	if err == nil {
		e.statusCode = fe.statusCode // records statusCode for forwards
		log.Printf("[executor] %-9s: (%d) %s %s => %s\n", e.policy.CmdType,
			fe.statusCode, e.r.Method, e.r.URL, newRequest.URL)
	}
	return err
}

func (e *policyExecutor) newForwardLocalRequest(fPath string) (*http.Request, error) {
	reqURL := e.r.URL

	if !strings.HasPrefix(fPath, "/") {
		uri := reqURL.Path
		fPath = path.Join(uri, "../"+fPath)
	}

	rawQuery := reqURL.RawQuery
	var requestURI string
	if len(rawQuery) == 0 {
		requestURI = fPath
	} else {
		requestURI = fmt.Sprintf("%s?%s", fPath, rawQuery)
	}
	newRequest, err := e.newForwardRequest(requestURI)
	if err != nil {
		return nil, err
	}
	return newRequest, err
}

func (e *policyExecutor) newForwardRequest(url string) (*http.Request, error) {
	method := e.r.Method
	body, err := ioutil.ReadAll(e.r.Body)
	if err != nil {
		return nil, &forwardError{err: err}
	}
	newRequest, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, &forwardError{err: err}
	}
	newRequest.Header = e.r.Header
	return newRequest, nil
}

func waitBeforeReturns(latency *mckmaps.Interval) {
	diff := latency.Max - latency.Min
	if diff > 0 {
		d := rand.Int63n(diff) + latency.Min
		time.Sleep(time.Duration(d * int64(time.Millisecond)))
	} else if latency.Min > 0 {
		time.Sleep(time.Duration(latency.Min * int64(time.Millisecond)))
	}
}
