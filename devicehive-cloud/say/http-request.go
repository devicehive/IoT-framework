package say

import (
	"fmt"
	"net/http"
)

func RequestStr(r *http.Request) (s string) {
	s = fmt.Sprintf("%s %s", r.Method, r.URL)
	if r.Header != nil || len(r.Header) != 0 {
		s = s + "("
		first := true
		for k, v := range r.Header {
			if !first {
				s = s + ","
			}
			s = s + k + fmt.Sprintf(":%v", v)
			first = false
		}
		s = s + ")"
	}
	return
}
