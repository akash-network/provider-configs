package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// PCIeDevice represents a PCIe device with its properties.
type PCIeDevice struct {
	Name       string `json:"name"`
	Interface  string `json:"interface"`
	MemorySize string `json:"memory_size"`
}

// VendorDevices represents a collection of PCIe devices under a specific vendor.
type VendorDevices struct {
	Name    string                `json:"name"`
	Devices map[string]PCIeDevice `json:"devices"`
}

// DeviceData holds the latest PCIe device data from the API and a mutex for synchronization.
type DeviceData struct {
	mu          sync.RWMutex
	devices     map[string]VendorDevices
	lastUpdated time.Time
	updateCount int64
	errorCount  int64
}

// validateJSON performs comprehensive validation of the JSON data
func (d *DeviceData) validateJSON(data []byte) error {
	// First, check if it's valid JSON
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON syntax: %w", err)
	}

	// Then, try to unmarshal into our expected structure
	var deviceData map[string]VendorDevices
	if err := json.Unmarshal(data, &deviceData); err != nil {
		return fmt.Errorf("JSON structure validation failed: %w", err)
	}

	// Additional business logic validation
	if len(deviceData) == 0 {
		return fmt.Errorf("validation failed: empty device data")
	}

	// Validate each vendor has required fields
	for vendorID, vendor := range deviceData {
		if vendor.Name == "" {
			return fmt.Errorf("validation failed: vendor %s missing name", vendorID)
		}

		if vendor.Devices == nil {
			return fmt.Errorf("validation failed: vendor %s has nil devices map", vendorID)
		}

		// Validate each device has required fields
		for deviceID, device := range vendor.Devices {
			if device.Name == "" {
				return fmt.Errorf("validation failed: device %s under vendor %s missing name", deviceID, vendorID)
			}
			if device.Interface == "" {
				return fmt.Errorf("validation failed: device %s under vendor %s missing interface", deviceID, vendorID)
			}
			// Note: MemorySize can be empty for some devices, so we don't validate it as required
		}
	}

	return nil
}

// Update fetches the latest data from the GitHub repository with validation.
func (d *DeviceData) Update() {
	log.Println("Attempting to update device data...")

	resp, err := http.Get("https://raw.githubusercontent.com/akash-network/provider-configs/main/devices/pcie/gpus.json")
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		d.incrementErrorCount()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP error: received status code %d", resp.StatusCode)
		d.incrementErrorCount()
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		d.incrementErrorCount()
		return
	}

	// Validate JSON before updating
	if err := d.validateJSON(body); err != nil {
		log.Printf("JSON validation failed, keeping previous data: %v", err)
		d.incrementErrorCount()
		return
	}

	// If validation passes, unmarshal the data
	var newData map[string]VendorDevices
	if err := json.Unmarshal(body, &newData); err != nil {
		// This shouldn't happen since we already validated, but let's be safe
		log.Printf("Unexpected error parsing validated JSON: %v", err)
		d.incrementErrorCount()
		return
	}

	// Update the data atomically
	d.mu.Lock()
	defer d.mu.Unlock()

	oldCount := len(d.devices)
	d.devices = newData
	d.lastUpdated = time.Now()
	d.updateCount++

	newCount := len(d.devices)
	log.Printf("Successfully updated device data: %d vendors (was %d), update #%d",
		newCount, oldCount, d.updateCount)
}

// GetStats returns current statistics about the data
func (d *DeviceData) GetStats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"vendor_count": len(d.devices),
		"last_updated": d.lastUpdated,
		"update_count": d.updateCount,
		"error_count":  d.errorCount,
		"has_data":     len(d.devices) > 0,
	}
}

// incrementErrorCount safely increments the error counter
func (d *DeviceData) incrementErrorCount() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.errorCount++
}

// ServeHTTP responds with the latest PCIe device data.
func (d *DeviceData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request:", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.devices) == 0 {
		http.Error(w, `{"error": "No device data available"}`, http.StatusServiceUnavailable)
		return
	}

	if err := json.NewEncoder(w).Encode(d.devices); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
}

// handleStats serves statistics about the service
func (d *DeviceData) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := d.GetStats()
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding stats response: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	}
}

// handleWebhook processes webhook requests and triggers an immediate update
func (d *DeviceData) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the body of the webhook request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Log the webhook request body for inspection
	log.Printf("Received webhook: %s", string(body))

	// Trigger an immediate update
	log.Println("Webhook received, triggering immediate update...")
	go d.Update()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "webhook received", "action": "update triggered"}`))
}

// healthCheck provides a simple health check endpoint
func (d *DeviceData) healthCheck(w http.ResponseWriter, r *http.Request) {
	stats := d.GetStats()

	status := "healthy"
	statusCode := http.StatusOK

	if !stats["has_data"].(bool) {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"status": status,
		"stats":  stats,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	data := &DeviceData{
		devices: make(map[string]VendorDevices), // Initialize with empty map
	}

	// Initialize the data with the first fetch
	log.Println("Starting server and performing initial data fetch...")
	data.Update()

	// Check if initial fetch was successful
	stats := data.GetStats()
	if !stats["has_data"].(bool) {
		log.Println("WARNING: Failed to fetch initial data. Server will start but with no device data.")
	} else {
		log.Printf("Initial fetch successful: loaded %d vendors", stats["vendor_count"])
	}

	// Start a goroutine to periodically update the data
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data.Update()
			}
		}
	}()

	// Set up the webhook handler
	http.HandleFunc("/devices/gpus/webhook", data.handleWebhook)

	// Set up the stats endpoint
	http.HandleFunc("/devices/gpus/stats", data.handleStats)

	// Set up the health check endpoint
	http.HandleFunc("/health", data.healthCheck)

	// Set up the main device data endpoint
	http.Handle("/devices/gpus", data)

	srv := &http.Server{
		Addr:         ":443",
		Handler:      nil, // nil uses http.DefaultServeMux
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Paths to the certificate and key files at the root of the Docker container
	certFile := "/cert.pem"
	keyFile := "/key.pem"

	// Start the HTTPS server in a goroutine
	go func() {
		log.Printf("Starting HTTPS server on %s", srv.Addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
			log.Fatalf("HTTPS server ListenAndServeTLS: %v", err)
		}
	}()

	// Set up channel to listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit // Wait for signal

	log.Println("Server is shutting down...")

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
