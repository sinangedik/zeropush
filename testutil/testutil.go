package testutil

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"strings"
)

const (
	CORRECT_AUTH_TOKEN = "correct_auth_token"
	WRONG_AUTH_TOKEN   = "wrong_auth_token"
)

// test server
func authenticate(w http.ResponseWriter, r *http.Request) bool {
	auth_header := r.Header.Get("Authorization")
	if s := strings.SplitN(auth_header, " ", 2); len(s) < 2 || s[1] != `token="`+CORRECT_AUTH_TOKEN+`"` {
		http.Error(w, `{"error":"unauthorized"}`, 401)
		return false
	}
	return true
}

func add_quota_headers(w http.ResponseWriter) {
	w.Header().Add("X-Device-Quota", "your quota")
	w.Header().Add("X-Device-Quota-Remaining", "your quota remaining")
	w.Header().Add("X-Device-Quota-Overage", "your quota overage")
}

func check_required_fields(w http.ResponseWriter, fields ...string) bool {
	for _, field := range fields {
		if field == "" {
			http.Error(w, `{"error":"missing required field"}`, 400)
			return false
		}
	}
	return true
}

func verify_credentials(w http.ResponseWriter, r *http.Request) {
	if !authenticate(w, r) {
		return
	}
	w.Write([]byte(`{"message":"authenticated", "auth_token_type":"server_token"}`))
	w.WriteHeader(200)
}

func get_inactive_tokens(w http.ResponseWriter, r *http.Request) {
	if !authenticate(w, r) {
		return
	}
	w.Write([]byte(`
	[
	  {
			"device_token": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcedf",
			"marked_inactive_at": "2013-03-11T16:25:14-04:00"
		},
		{
		  "device_token": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcedf",
			"marked_inactive_at": "2013-03-11T16:25:14-04:00"
		}
	]`))
	w.WriteHeader(200)
}

func register_device(w http.ResponseWriter, r *http.Request) {
	if !authenticate(w, r) {
		return
	}
	params := mux.Vars(r)
	add_quota_headers(w)
	if !check_required_fields(w, params["device_token"]) {
		return
	}
	w.Write([]byte(`{"message":"ok"}`))
	w.WriteHeader(200)
}

func subscribe_to_channel(w http.ResponseWriter, r *http.Request) {
	subscribe(w, r, true)
}

func unsubscribe_from_channel(w http.ResponseWriter, r *http.Request) {
	subscribe(w, r, false)
}

func subscribe(w http.ResponseWriter, r *http.Request, sub bool) {
	authenticate(w, r)
	params := mux.Vars(r)
	add_quota_headers(w)
	if !(check_required_fields(w, params["device_token"])) {
		return
	}
	channels := `[]`
	if sub {
		channels = `["foo"]`
	}
	w.Write([]byte(`{
			  "device_token": "1236372819B36278G6783G21678321",
			  "channels":` + channels + `}`))
	w.WriteHeader(200)
}

func notify(w http.ResponseWriter, r *http.Request) {
	authenticate(w, r)
	params := mux.Vars(r)
	add_quota_headers(w)
	if !(check_required_fields(w, params["device_tokens"])) {
		return
	}
	w.Write([]byte(`{
		sent_count: 0,
		inactive_tokens:[],
	  unregistered_tokens:[
	"1234567891abcdef1234567890abcdef1234567890abcdef1234567890abcedf",
"1234567890abcdef1234567890abcdef1234567890abcdef1234567890abceee"
	]
	}`))
	w.WriteHeader(200)
}

func broadcast_to_channel(w http.ResponseWriter, r *http.Request) {
	authenticate(w, r)
	add_quota_headers(w)
	w.Write([]byte(`{
			  "sent_count": 100
	}`))
	w.WriteHeader(200)
}
func set_badge(w http.ResponseWriter, r *http.Request) {
	authenticate(w, r)
	params := mux.Vars(r)
	add_quota_headers(w)
	if !(check_required_fields(w, params["device_token"], params["badge"])) {
		return
	}
	w.Write([]byte(`{
		"message":"ok"
	}`))
	w.WriteHeader(200)
}
func get_device(w http.ResponseWriter, r *http.Request) {
	authenticate(w, r)
	params := mux.Vars(r)
	add_quota_headers(w)
	if !(check_required_fields(w, params["device_token"])) {
		return
	}
	w.Write([]byte(`{
			  "token": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcedf",
				  "active": true,
					  "marked_inactive_at": null,
						  "badge": 1,
							  "channels": [
								    "testflight",
										    "user@example.com"
												  ]
												}
	`))
	w.WriteHeader(200)
}
func NewZeroTestServer() *httptest.Server {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/verify_credentials", verify_credentials).Methods("GET")
	rtr.HandleFunc("/inactive_tokens", get_inactive_tokens).Methods("GET")
	rtr.HandleFunc("/register", register_device).Methods("POST").Queries("device_token", "{device_token}", "channel", "{channel}")
	rtr.HandleFunc("/register", register_device).Methods("POST").Queries("device_token", "{device_token}")
	rtr.HandleFunc("/register", register_device).Methods("POST").Queries("channel", "{channel}")
	rtr.HandleFunc("/register", register_device).Methods("POST")
	rtr.HandleFunc("/unregister", register_device).Methods("DELETE").Queries("device_token", "{device_token}", "channel", "{channel}")
	rtr.HandleFunc("/unregister", register_device).Methods("DELETE").Queries("device_token", "{device_token}")
	rtr.HandleFunc("/unregister", register_device).Methods("DELETE").Queries("channel", "{channel}")
	rtr.HandleFunc("/unregister", register_device).Methods("DELETE")
	rtr.HandleFunc("/subscribe/{channel}", subscribe_to_channel).Methods("POST").Queries("device_token", "{device_token}")
	rtr.HandleFunc("/subscribe/{channel}", subscribe_to_channel).Methods("POST")
	rtr.HandleFunc("/broadcast/{channel}", broadcast_to_channel).Methods("POST")
	rtr.HandleFunc("/subscribe/{channel}", unsubscribe_from_channel).Methods("DELETE").Queries("device_token", "{device_token}")
	rtr.HandleFunc("/subscribe/{channel}", unsubscribe_from_channel).Methods("DELETE")
	rtr.HandleFunc("/set_badge", set_badge).Methods("POST").Queries("device_token", "{device_token}", "badge", "{badge}")
	rtr.HandleFunc("/notify", register_device).Methods("POST").Queries("device_tokens[]", "{device_tokens}", "alert", "{alert}", "badge", "{badge}", "sound", "{sound}", "info", "{info}", "expiry", "{expiry}", "content_available", "{content_available}", "category", "{category}")
	rtr.HandleFunc("/devices/{device_token}", get_device).Methods("GET")
	return httptest.NewServer(rtr)
}
