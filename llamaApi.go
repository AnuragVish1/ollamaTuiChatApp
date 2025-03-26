package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// struct for request

type request struct {
	Model    string     `json:"model"`
	Messages []messages `json:"messages"`
	Stream   bool       `json:"stream"`
}

type messages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// struct for response

type response struct {
	Model      string `json:"model"`
	Created_at string `json:"created_at"`
	Message    struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done_reason          string `json:"done_reason"`
	Done                 bool   `json:"done"`
	Total_duration       int64  `json:"total_duration"`
	Load_duration        int64  `json:"load_duration"`
	Prompt_eval_count    int    `json:"prompt_eval_count"`
	Prompt_eval_duration int64  `json:"prompt_eval_duration"`
	Eval_count           int    `json:"eval_count"`
	Eval_duration        int64  `json:"eval_duration"`
}

// sending and receiving message

func sendingMessage(r request) (response, error) {
	apiEndPoint := "http://127.0.0.1:11434/api/chat"

	// making the request to json format

	jsonData, err := json.Marshal(r)

	if err != nil {
		return response{}, err
	}

	req, err := http.NewRequest("POST", apiEndPoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return response{}, err
	}

	// setting headers

	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		print("Err")
		return response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		print("errpr")
		return response{}, err
	}

	// Parse JSON response
	var result response
	err = json.Unmarshal(body, &result)
	if err != nil {
		print("Eror here")
		return response{}, err
	}

	return result, nil

}
