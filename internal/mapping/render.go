package mapping

import (
	"net/http"
)

func (pr *PolicyReturns) Render(w http.ResponseWriter) error {
	w.WriteHeader(int(pr.StatusCode))
	pr.Headers.render(w)

	var err error
	if pr.Body != "" {
		_, err = w.Write([]byte(pr.Body))
	}

	return err
}

func (h *Headers) render(w http.ResponseWriter) {
	outHeader := w.Header()

	for name, values := range h.headers {
		if _, ok := outHeader[name]; ok {
			outHeader.Del(name)
		}

		for _, value := range values {
			outHeader.Add(name, value)
		}
	}
}
