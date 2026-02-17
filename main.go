package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// GlobalSettings holds arbitrary global settings from configuration
type GlobalSettings map[string]interface{}

// ChannelConfig represents configuration for a single WebSocket channel
type ChannelConfig struct {
	Path    string `json:"path"`
	Logfile string `json:"logfile,omitempty"`
	// other per-channel options may be added later
}

// AppConfig represents the full JSON configuration file
type AppConfig struct {
	GlobalSettings GlobalSettings  `json:"global_settings"`
	Channels       []ChannelConfig `json:"channels"`
}

// ChannelManager manages all WebSocket channels dynamically
type ChannelManager struct {
	clients   map[string]map[*websocket.Conn]bool // endpoint -> clients map
	broadcast map[string]chan []byte              // endpoint -> broadcast channel
	mutex     sync.RWMutex                        // protects clients map
}

// NewChannelManager creates a new channel manager
func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		clients:   make(map[string]map[*websocket.Conn]bool),
		broadcast: make(map[string]chan []byte),
	}
}

// RegisterEndpoint registers a new endpoint
func (cm *ChannelManager) RegisterEndpoint(endpoint string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if _, exists := cm.clients[endpoint]; !exists {
		cm.clients[endpoint] = make(map[*websocket.Conn]bool)
		// Buffered to prevent a slow/broken writer path from blocking readers.
		cm.broadcast[endpoint] = make(chan []byte, 256)
		log.Printf("Registered endpoint: %s", endpoint)
	}
}

// AddClient adds a client to an endpoint
func (cm *ChannelManager) AddClient(endpoint string, conn *websocket.Conn) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if clients, exists := cm.clients[endpoint]; exists {
		clients[conn] = true
		return true
	}
	return false
}

// RemoveClient removes a client from an endpoint
func (cm *ChannelManager) RemoveClient(endpoint string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if clients, exists := cm.clients[endpoint]; exists {
		delete(clients, conn)
	}
}

// Broadcast sends a message to all clients of an endpoint
func (cm *ChannelManager) Broadcast(endpoint string, message []byte) {
	cm.mutex.RLock()
	broadcastChan, exists := cm.broadcast[endpoint]
	cm.mutex.RUnlock()
	
	if exists {
		broadcastChan <- message
	}
}

// ResolveEndpoint maps an incoming request path to a registered endpoint.
// It prefers exact matches, but also supports prefix matches (longest wins)
// because net/http mux patterns ending with '/' match subpaths.
func (cm *ChannelManager) ResolveEndpoint(requestPath string) (string, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if requestPath == "" {
		return "", false
	}

	// Fast-path: exact match.
	if _, ok := cm.clients[requestPath]; ok {
		return requestPath, true
	}

	// Also try with enforced trailing slash.
	normalized := requestPath
	if !strings.HasSuffix(normalized, "/") {
		normalized += "/"
	}
	if _, ok := cm.clients[normalized]; ok {
		return normalized, true
	}

	// Prefix match: choose the longest registered endpoint that prefixes requestPath.
	var best string
	for endpoint := range cm.clients {
		if strings.HasPrefix(requestPath, endpoint) || strings.HasPrefix(normalized, endpoint) {
			if len(endpoint) > len(best) {
				best = endpoint
			}
		}
	}
	if best == "" {
		return "", false
	}
	return best, true
}

// GetBroadcastChannel returns the broadcast channel for an endpoint
func (cm *ChannelManager) GetBroadcastChannel(endpoint string) chan []byte {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.broadcast[endpoint]
}

// SnapshotClients returns a stable snapshot of currently connected clients
// for an endpoint. This avoids data races with concurrent Add/Remove.
func (cm *ChannelManager) SnapshotClients(endpoint string) []*websocket.Conn {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clientsMap := cm.clients[endpoint]
	if len(clientsMap) == 0 {
		return nil
	}

	out := make([]*websocket.Conn, 0, len(clientsMap))
	for c := range clientsMap {
		out = append(out, c)
	}
	return out
}

// GetAllEndpoints returns all registered endpoints
func (cm *ChannelManager) GetAllEndpoints() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	endpoints := make([]string, 0, len(cm.clients))
	for endpoint := range cm.clients {
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

var channelManager = NewChannelManager()

// globalSettings stores loaded global_settings from configuration
var globalSettings GlobalSettings

// channelLoggers maps endpoint path to a dedicated logger writing to the
// configured logfile (if any). It is populated once at startup.
var channelLoggers = make(map[string]*log.Logger)

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// loadChannelsFromFile loads channel endpoints from a JSON configuration file.
// The file is expected to contain global_settings and channels (see AppConfig).
func loadChannelsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Store global_settings for future use (currently not processed further)
	globalSettings = cfg.GlobalSettings

	var endpoints []string

	for _, ch := range cfg.Channels {
		path := strings.TrimSpace(ch.Path)
		if path == "" {
			continue
		}

		// Ensure endpoint ends with /
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}

		endpoints = append(endpoints, path)

		// If logfile is configured for this path, prepare a dedicated logger
		if ch.Logfile != "" {
			f, err := os.OpenFile(ch.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Error opening logfile %s for path %s: %v", ch.Logfile, path, err)
				continue
			}

			channelLoggers[path] = log.New(f, "", log.LstdFlags)
			log.Printf("Enabled logging for %s to %s", path, ch.Logfile)
		}
	}

	return endpoints, nil
}

func main() {
	// Parse command line flags
	configFile := flag.String("c", "", "Configuration file in JSON format (global_settings and channels)")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("Error: configuration file must be specified with -c flag")
	}

	// Load channels from configuration file
	endpoints, err := loadChannelsFromFile(*configFile)
	if err != nil {
		log.Fatalf("Error loading configuration file: %v", err)
	}

	if len(endpoints) == 0 {
		log.Fatal("Error: no endpoints found in configuration file")
	}

	log.Printf("Loaded %d endpoints from configuration file", len(endpoints))

	// Register all endpoints
	for _, endpoint := range endpoints {
		channelManager.RegisterEndpoint(endpoint)
		http.HandleFunc(endpoint, handleConnections)
		// Also register the non-slash variant to avoid net/http redirect,
		// which breaks WebSocket upgrades.
		trimmed := strings.TrimSuffix(endpoint, "/")
		if trimmed != "" && trimmed != endpoint {
			http.HandleFunc(trimmed, handleConnections)
		}
	}

	// Start listening for incoming messages for each endpoint
	for _, endpoint := range endpoints {
		go handleMessages(endpoint)
	}

	// Start the server on localhost port 8023 and log any errors
	log.Println("http server started on :8023")
	err = http.ListenAndServe(":8023", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}


// logMessageForEndpoint logs an incoming message to a per-endpoint logfile
// if one has been configured for the given endpoint.
func logMessageForEndpoint(endpoint string, msg []byte) {
	logger, ok := channelLoggers[endpoint]
	if !ok || logger == nil {
		return
	}

	// logger is configured with standard flags, so it will automatically
	// prepend date/time to each log line.
	logger.Printf("%s", msg)
}

// handleConnections handles WebSocket connections for any endpoint
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	requestPath := r.URL.Path
	endpoint, ok := channelManager.ResolveEndpoint(requestPath)
	log.Printf("Wake up on URL: %s", requestPath)

	// Check if endpoint is registered
	if !ok || !channelManager.AddClient(endpoint, ws) {
		log.Printf("Unrouted URL: %s", requestPath)
		return
	}

	for {
		// Read in a new message as JSON and map it to a Message object
		_, amsg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error reading from %s: %v", endpoint, err)
			channelManager.RemoveClient(endpoint, ws)
			break
		}

		log.Printf("Read from %s: [%s]", endpoint, amsg)

		// If a logfile is configured for this endpoint, log the message there
		logMessageForEndpoint(endpoint, amsg)

		// Send the newly received message to the broadcast channel
		channelManager.Broadcast(endpoint, amsg)
	}
}

// handleMessages handles messages for a specific endpoint
func handleMessages(endpoint string) {
	broadcastChan := channelManager.GetBroadcastChannel(endpoint)
	
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcastChan

		// Send it out to every client that is currently connected
		clients := channelManager.SnapshotClients(endpoint)
		for _, client := range clients {
			// Never let a single stuck client block the whole channel.
			_ = client.SetWriteDeadline(time.Now().Add(2 * time.Second))
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("error writing to client on %s: %v", endpoint, err)
				client.Close()
				channelManager.RemoveClient(endpoint, client)
			}
		}
	}
}
