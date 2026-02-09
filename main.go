package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

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
		cm.broadcast[endpoint] = make(chan []byte)
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

// GetBroadcastChannel returns the broadcast channel for an endpoint
func (cm *ChannelManager) GetBroadcastChannel(endpoint string) chan []byte {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.broadcast[endpoint]
}

// GetClients returns the clients map for an endpoint
func (cm *ChannelManager) GetClients(endpoint string) map[*websocket.Conn]bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.clients[endpoint]
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

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// loadChannelsFromFile loads channel endpoints from a configuration file
func loadChannelsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var endpoints []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			// Ensure endpoint ends with /
			if !strings.HasSuffix(line, "/") {
				line += "/"
			}
			endpoints = append(endpoints, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return endpoints, nil
}

func main() {
	// Parse command line flags
	configFile := flag.String("c", "", "Configuration file with channel endpoints (one per line)")
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

	endpoint := r.URL.Path
	log.Printf("Wake up on URL: %s", endpoint)

	// Check if endpoint is registered
	if !channelManager.AddClient(endpoint, ws) {
		log.Printf("Unrouted URL: %s", endpoint)
		return
	}

	for {
		// Read in a new message as JSON and map it to a Message object
		_, amsg, err := ws.ReadMessage()

		log.Printf("Read from %s: [%s]", endpoint, amsg)

		if err != nil {
			log.Printf("error: %v", err)
			channelManager.RemoveClient(endpoint, ws)
			break
		}

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
		clients := channelManager.GetClients(endpoint)
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("error writing to client on %s: %v", endpoint, err)
				client.Close()
				channelManager.RemoveClient(endpoint, client)
			}
		}
	}
}
