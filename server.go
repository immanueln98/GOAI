package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strings"

	"golang.org/x/image/draw"
)

func main() {
	// Entry point of the server application
	app := App{}
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}

type App struct {
	// Application state and configurations

}

func (app App) Start() error {
	const serverAddr string = "0.0.0.0:3001"
	log.Printf("Starting server at %s\n", serverAddr)
	// Logic to start the server would go here
	http.HandleFunc("POST /api/image-search", app.imageSearch)
	return http.ListenAndServe(serverAddr, nil)
}

func (app App) imageSearch(w http.ResponseWriter, r *http.Request) {
	// Handler for image search requests
	log.Println("Image search handler invoked")
	type CelebMatchRequest struct {
		Image64 string `json:"img"`
	}
	// Deserialize request, process image, and respond
	var req CelebMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Printf("Received image: %s\n", req.Image64)
	// split image into metadata and data
	imgParts := strings.Split(req.Image64, ",")
	parts := len(imgParts)
	if parts != 2 {
		log.Printf("Invalid image format. Got %d parts", parts)
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	stdImage, err := standadizeImage(imgParts[1])
	if err != nil {
		log.Printf("Failed to standardize image: %v", err)
		http.Error(w, "Image processing failed", http.StatusInternalServerError)
		return
	}
	_ = stdImage // Use the standardized image for further processing

}

func standadizeImage(imageB64 string) (*string, error) {
	// Get the base64 decoder as an io.Reader
	b64Decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imageB64))
	// Decode the image
	origImg, _, err := image.Decode(b64Decoder)
	if err != nil {
		log.Printf("Failed to decode image: %v", err)
		return nil, err
	}
	// Resize to 800x600
	resizedImg := image.NewRGBA(image.Rect(0, 0, 800, 600))
	draw.NearestNeighbor.Scale(resizedImg, resizedImg.Bounds(), origImg, origImg.Bounds(), draw.Over, nil)

	var jpegToSend bytes.Buffer
	if err := jpeg.Encode(&jpegToSend, resizedImg, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("standardazing image failed: %w", err)
	}
	encodedStr := base64.StdEncoding.EncodeToString(jpegToSend.Bytes())
	return &encodedStr, nil
}
