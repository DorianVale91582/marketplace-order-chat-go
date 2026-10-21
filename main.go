package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func main() {
	c, err := newInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in roomRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		data, err := openRoom(r.Context(), c, in)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
	http.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/handoff") {
			http.Error(w, "not found", 404)
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "not found", 404)
			return
		}
		var body struct{ RoomID, SellerID, BuyerID, Note string }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		if err := publishHandoff(r.Context(), c, roomRequest{RoomID: body.RoomID, SellerID: body.SellerID, BuyerID: body.BuyerID}, body.Note); err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	log.Println("chat rooms listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
