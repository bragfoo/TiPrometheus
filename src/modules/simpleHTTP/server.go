package simpleHTTP

import (
	"net/http"
)

func Server() {
	http.HandleFunc("/write", RemoteWrtie)
	http.HandleFunc("/read", RemoteRead)
	http.ListenAndServe(":12350", nil)
}
