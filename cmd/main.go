package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"videoUploadAndProcessing/pkg/upload"

	"github.com/joho/godotenv"
)

func main() {
	// load .env file from enviroments variable
	envPath := os.Getenv("MYAPP_ENV_PATH")
	if envPath == "" {
		envPath = ".env" // Default to the local .env file
	}

	// load .env file
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file from path %s: %v", envPath, err)
	}

	logfilepath := os.Getenv("VIDEO_PROCESSING_LOG_PATH")

	// Extract the directory path from the log file path
	logDir, _ := filepath.Split(logfilepath)

	// Create the directory (including any necessary parent directories) if it doesn't exist
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create log directory: %v\n", err)
	}

	// Create the log file with necessary permissions
	logFile, err := os.OpenFile(logfilepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open or create log file: %v\n", err)
	}
	defer logFile.Close()

	// Redirect log output to file
	log.SetOutput(logFile)
	log.Printf("Writing log in %s", logfilepath)
	// Create a tmp directory
	tmpDir := "tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		log.Fatalf("Failed to create tmp directory: %v\n", err)
	}

	// Create a new job queue with 100 slots.
	jobQueue := make(chan upload.Job, 100)

	// Initialize a slice of workers based on the defined number of workers in the upload package.
	workers := make([]upload.Worker, upload.NumWorkers)

	// Initialize and start all the workers.
	for i := 0; i < upload.NumWorkers; i++ {
		workers[i] = upload.Worker{
			ID:       i + 1,    // Assign a unique ID to each worker starting from 1.
			JobQueue: jobQueue, // All workers share the same job queue.
		}
		workers[i].Start() // Start the worker.
	}

	// Create a new HTTP ServeMux.
	mux := http.NewServeMux()

	// Register a new route that handles video uploads.
	mux.HandleFunc("/video-processing-trigger", func(w http.ResponseWriter, r *http.Request) {
		// Use the first worker to handle the upload.
		// In reality, any worker could handle this since they all share the same job queue.
		upload.HandleUpload(w, r, workers[0])
	})

	// Define the port for the server.
	port := os.Getenv("VIDEO_PROCESSING_PORT")
	log.Printf("Starting server on port %s\n", port)

	// Start the HTTP server.
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}
