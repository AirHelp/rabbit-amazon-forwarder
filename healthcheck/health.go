package health

import "net/http"

const (
	success = "success"
)

// Check verifies if application is working properly
func Check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(success))
}
