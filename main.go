package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//  a simple note-taking API
// // the api allows:
// 1. POST /api/notes — create a note (learn reading request body)
// 2. GET /api/notes — list all notes
// 3. GET /api/notes/{id} — get one note (learn path params)
// 4. DELETE /api/notes/{id} — delete a note

// Have a struct for the post, i.e what are we expecting
type NoteInput struct {
	Message string `json:"message"`
}

// Data we're storing or sending back it's struct
type Note struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// the error messag and response wrapper
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

var notes = map[string]Note{}

func CollectNote(w http.ResponseWriter, r *http.Request) {
	var note NoteInput
	err := json.NewDecoder(r.Body).Decode(&note)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if note.Message == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "No content found in the Post",
		})
		return
	}

	var ter Note

	id := fmt.Sprintf("%d", time.Now().UnixNano())

	ter.ID = id
	ter.Message = note.Message
	ter.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	notes[ter.ID] = ter

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   ter,
	})
}

func GetAllNote(w http.ResponseWriter, r *http.Request) {
	// Have your response variable decleared, it should be a list of notes
	allnotes := make([]Note, 0)

	// loop through the notes and append the note to the variable declared above
	for _, v := range notes {
		allnotes = append(allnotes, v)
	}

	//set header
	w.Header().Set("Content-Type", "application/json")

	// Status Code
	w.WriteHeader(http.StatusOK)

	// write the respons back to our Client
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   allnotes,
	})

}

func GetSingleNote(w http.ResponseWriter, r *http.Request) {
	// get the ID
	id := r.PathValue("id")

	// check if it's in the note if no return Error
	val, ok := notes[id]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Note with that ID isn't available",
		})
		return
	}
	// set header
	w.Header().Set("Content-Type", "application/json")

	// set Status
	w.WriteHeader(http.StatusOK)

	// encode and send response back
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   val,
	})

}

func DeleteSingleNote(w http.ResponseWriter, r *http.Request) {
	// get the id
	id := r.PathValue("id")

	// check if that id is in the notes
	_, ok := notes[id]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Note with that ID isn't available",
		})
		return
	}

	// Delete id from notes
	delete(notes, id)

	// set header
	w.Header().Set("Content-Type", "application/json")

	// set Status
	w.WriteHeader(http.StatusOK)

	// encode and send response back
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   fmt.Sprintf("Note with id %v was deleted successfully", id),
	})

}

func main() {
	// Handler and Routing
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/notes/{id}", DeleteSingleNote)
	mux.HandleFunc("GET /api/notes/{id}", GetSingleNote)
	mux.HandleFunc("GET /api/notes", GetAllNote)
	mux.HandleFunc("POST /api/notes", CollectNote)

	// Server
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
