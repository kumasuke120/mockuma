package mapping

import (
	"net/http"
)

func (pr *PolicyReturns) Render(w *http.ResponseWriter) error {
	pr.Headers.render(w)

	(*w).WriteHeader(int(pr.StatusCode)) // must be called after (*w).Header() modifications

	var err error
	if pr.Body != nil && len(pr.Body) != 0 {
		_, err = (*w).Write(pr.Body)
	}

	return err
}

func (h *Headers) render(w *http.ResponseWriter) {
	outHeader := (*w).Header()

	for name, values := range h.headers {
		if _, ok := outHeader[name]; ok {
			outHeader.Del(name)
		}

		for _, value := range values {
			outHeader.Add(name, value)
		}
	}
}
