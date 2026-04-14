package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GenderAPIResponse struct {
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
	Count       int     `json:"count"`
}

// Our Server Response
type ServerResponse struct {
	Name          string  `json:"name"`
	Gender        string  `json:"gender"`
	Probability   float64 `json:"probability"`
	Count         int     `json:"sample_size"`
	Confidence    bool    `json:"is_confident"`
	ProcessedTime string  `json:"processed_at"`
}

type SuccessResponse struct {
	Status string         `json:"status"`
	Data   ServerResponse `json:"data"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func classify(w http.ResponseWriter, r *http.Request) {
	// Get the request from our client
	name := r.URL.Query().Get("name")

	// validate that name isn't empty
	if name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "name query parameter is required",
		})
		return
	}

	// Create the Client Object
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// make the request
	url := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
	resp, err := client.Get(url)
	if err != nil {
		// This catches network errors (e.g., server is down)
		fmt.Printf("Failed to reach server: %v\n", err)
		return
	}

	// close the connection
	defer resp.Body.Close()

	// Additional check
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server returned an error: %d\n", resp.StatusCode)
		return
	}

	//read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	// 6. Unmarshal (The Go version of json.loads)
	var result GenderAPIResponse

	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		return
	}

	// OUR LOGIC
	if result.Gender == "" || result.Count == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "No prediction available for the provided name",
		})
		return
	}

	// the confidences score
	response := ServerResponse{
		Name:          name,
		Gender:        result.Gender,
		Probability:   result.Probability,
		Count:         result.Count,
		ProcessedTime: time.Now().UTC().Format(time.RFC3339),
	}

	if response.Probability >= 0.7 && response.Count >= 100 {
		response.Confidence = true
	}

	// Set header for our response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Status Code
	w.WriteHeader(http.StatusOK)

	// write the respons back to our Client
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   response,
	})

}

func main() {
	// Handler and Routing
	mux := http.NewServeMux()
	mux.HandleFunc("/api/classify", classify)

	// Server
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
