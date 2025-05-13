package mocks

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type ChatRequest struct {
	UserID string `json:"user_id"`
	Query  string `json:"query"`
}

func StartMockBackendServer() {
	port := os.Getenv("MOCK_BACKEND_PORT")
	if port == "" {
		port = "8091"
	}

	http.HandleFunc("/v1/chat/stream", streamHandler)

	log.Printf("Mock backend server started on :%s ...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}


func streamHandler(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Content-Type","text/event-stream")
	w.Header().Set("Cache-control","no-cache")
	w.Header().Set("Connection","keep-alive")

	//Decode incoming JSON request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err!= nil {
		http.Error(w, "Invalid JSON",http.StatusBadRequest)
		return;
	}

	//Simulate streaming a response word-by-word
	response := []string {
		"Hello", ",", req.UserID, "!", "You", "said:", req.Query,
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for i,word := range response {
		fmt.Fprintf(w, "id: %d\ndata: %s\n\n", i+1, word)
		flusher.Flush()

		//this is to simulate delay
		time.Sleep(500 * time.Millisecond)
	}
}