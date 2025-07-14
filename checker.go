package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type RequestBody struct {
	ToEmail string `json:"to_email"`
}

type ResponseBody struct {
	IsReachable string `json:"is_reachable"`
}

const (
	concurrency = 500 // Change this to increase number of concerancy
	apiURL      = "http://localhost:9100/v0/check_email"
	inputFile   = "emails.txt"
	validFile   = "valid_emails.txt"
	invalidFile = "invalid_emails.txt"
)

func main() {
	// Clear output files
	_ = os.WriteFile(validFile, []byte{}, 0644)
	_ = os.WriteFile(invalidFile, []byte{}, 0644)

	// Read emails
	emails, err := readLines(inputFile)
	if err != nil {
		fmt.Println("Failed to read input file:", err)
		return
	}

	emailChan := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex // Protect file writing

	// Start worker pool
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for email := range emailChan {
				status := checkEmail(email)

				// Print result
				fmt.Printf("%s: %s\n", email, status)

				// Write result safely
				mu.Lock()
				if status == "valid" {
					appendToFile(validFile, email+"\n")
				} else {
					appendToFile(invalidFile, email+"\n")
				}
				mu.Unlock()
			}
		}()
	}

	// Feed emails to workers
	for _, email := range emails {
		emailChan <- email
	}
	close(emailChan)

	// Wait for all workers to finish
	wg.Wait()
}

func readLines(filename string) ([]string, error) {
	var lines []string
	file, err := os.Open(filename)
	if err != nil {
		return lines, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
	}
	return lines, nil
}

func checkEmail(email string) string {
	body := RequestBody{ToEmail: email}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Request error for %s: %v\n", email, err)
		return "error"
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var response ResponseBody
	_ = json.Unmarshal(respBytes, &response)

	return response.IsReachable
}

func appendToFile(filename, text string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", filename, err)
		return
	}
	defer f.Close()
	f.WriteString(text)
}
