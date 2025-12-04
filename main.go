package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var clients_settings = make(map[*websocket.Conn]bool) // connected clients
var clients_overview = make(map[*websocket.Conn]bool) // connected clients
var clients_jailscontainers = make(map[*websocket.Conn]bool) // connected clients
var clients_instance_jail = make(map[*websocket.Conn]bool) // connected clients
var clients_bhyvevms = make(map[*websocket.Conn]bool) // connected clients
var clients_vm_packages = make(map[*websocket.Conn]bool) // connected clients
var clients_nodes = make(map[*websocket.Conn]bool) // connected clients
var clients_vpnet = make(map[*websocket.Conn]bool) // connected clients
var clients_authkey = make(map[*websocket.Conn]bool) // connected clients
var clients_media = make(map[*websocket.Conn]bool) // connected clients
var clients_repo = make(map[*websocket.Conn]bool) // connected clients
var clients_bases = make(map[*websocket.Conn]bool) // connected clients
var clients_sources = make(map[*websocket.Conn]bool) // connected clients
var clients_jail_marketplace = make(map[*websocket.Conn]bool) // connected clients
var clients_bhyve_marketplace = make(map[*websocket.Conn]bool) // connected clients
var clients_tasklog = make(map[*websocket.Conn]bool) // connected clients
var clients_users = make(map[*websocket.Conn]bool) // connected clients
var clients_imported = make(map[*websocket.Conn]bool) // connected imported
var clients_k8s = make(map[*websocket.Conn]bool) // connected k8s

var broadcast_settings = make(chan []byte)           // broadcast channel
var broadcast_overview = make(chan []byte)           // broadcast channel
var broadcast_jailscontainers = make(chan []byte)           // broadcast channel
var broadcast_instance_jail = make(chan []byte)           // broadcast channel
var broadcast_bhyvevms = make(chan []byte)           // broadcast channel
var broadcast_vm_packages = make(chan []byte)           // broadcast channel
var broadcast_nodes = make(chan []byte)           // broadcast channel
var broadcast_vpnet = make(chan []byte)           // broadcast channel
var broadcast_authkey = make(chan []byte)           // broadcast channel
var broadcast_media = make(chan []byte)           // broadcast channel
var broadcast_repo = make(chan []byte)           // broadcast channel
var broadcast_bases = make(chan []byte)           // broadcast channel
var broadcast_sources = make(chan []byte)           // broadcast channel
var broadcast_jail_marketplace = make(chan []byte)           // broadcast channel
var broadcast_bhyve_marketplace = make(chan []byte)           // broadcast channel
var broadcast_tasklog = make(chan []byte)           // broadcast channel
var broadcast_users = make(chan []byte)           // broadcast channel
var broadcast_imported = make(chan []byte)           // broadcast channel
var broadcast_k8s = make(chan []byte)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Configure websocket route
	http.HandleFunc("/clonos/clonos/settings/", handleConnections)
	http.HandleFunc("/clonos/clonos/overview/", handleConnections)
	http.HandleFunc("/clonos/clonos/jailscontainers/", handleConnections)
	http.HandleFunc("/clonos/clonos/instance_jail/", handleConnections)
	http.HandleFunc("/clonos/clonos/bhyvevms/", handleConnections)
	http.HandleFunc("/clonos/clonos/vm_packages/", handleConnections)
	http.HandleFunc("/clonos/clonos/nodes/", handleConnections)
	http.HandleFunc("/clonos/clonos/vpnet/", handleConnections)
	http.HandleFunc("/clonos/clonos/authkey/", handleConnections)
	http.HandleFunc("/clonos/clonos/media/", handleConnections)
	http.HandleFunc("/clonos/clonos/repo/", handleConnections)
	http.HandleFunc("/clonos/clonos/bases/", handleConnections)
	http.HandleFunc("/clonos/clonos/sources/", handleConnections)
	http.HandleFunc("/clonos/clonos/jail_marketplace/", handleConnections)
	http.HandleFunc("/clonos/clonos/bhyve_marketplace/", handleConnections)
	http.HandleFunc("/clonos/clonos/tasklog/", handleConnections)
	http.HandleFunc("/clonos/clonos/users/", handleConnections)
	http.HandleFunc("/clonos/clonos/imported/", handleConnections)
	http.HandleFunc("/clonos/clonos/k8s/", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages_overview()
	go handleMessages_settings()
	go handleMessages_jailscontainers()
	go handleMessages_instance_jail()
	go handleMessages_bhyvevms()
	go handleMessages_vm_packages()
	go handleMessages_nodes()
	go handleMessages_vpnet()
	go handleMessages_authkey()
	go handleMessages_media()
	go handleMessages_repo()
	go handleMessages_bases()
	go handleMessages_sources()
	go handleMessages_jail_marketplace()
	go handleMessages_bhyve_marketplace()
	go handleMessages_tasklog()
	go handleMessages_users()
	go handleMessages_imported()
	go handleMessages_k8s()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8023")
	err := http.ListenAndServe(":8023", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}


// TODO: need to get pattern and use it as channel route
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	log.Println("Wake up on URL: ",r.URL.Path)

	//todo: we need arrays if clients[r.URL.Path][ws]
	// enum for { channel1, channel2 } ?
	switch r.URL.Path {
		case "/clonos/clonos/settings/":
			clients_settings[ws] = true
		case "/clonos/clonos/overview/":
			clients_overview[ws] = true
		case "/clonos/clonos/jailscontainers/":
			clients_jailscontainers[ws] = true
		case "/clonos/clonos/instance_jail/":
			clients_instance_jail[ws] = true
		case "/clonos/clonos/bhyvevms/":
			clients_bhyvevms[ws] = true
		case "/clonos/clonos/vm_packages/":
			clients_vm_packages[ws] = true
		case "/clonos/clonos/nodes/":
			clients_nodes[ws] = true
		case "/clonos/clonos/vpnet/":
			clients_vpnet[ws] = true
		case "/clonos/clonos/authkey/":
			clients_authkey[ws] = true
		case "/clonos/clonos/media/":
			clients_media[ws] = true
		case "/clonos/clonos/repo/":
			clients_repo[ws] = true
		case "/clonos/clonos/bases/":
			clients_bases[ws] = true
		case "/clonos/clonos/sources/":
			clients_sources[ws] = true
		case "/clonos/clonos/jail_marketplace/":
			clients_jail_marketplace[ws] = true
		case "/clonos/clonos/bhyve_marketplace/":
			clients_bhyve_marketplace[ws] = true
		case "/clonos/clonos/tasklog/":
			clients_tasklog[ws] = true
		case "/clonos/clonos/users/":
			clients_users[ws] = true
		case "/clonos/clonos/imported/":
			clients_imported[ws] = true
		case "/clonos/clonos/k8s/":
			clients_k8s[ws] = true
		default:
			log.Println("Urouted URL: ",r.URL.Path)
			return
	}

	for {
		// Read in a new message as JSON and map it to a Message object
		_, amsg, err := ws.ReadMessage()

		log.Printf("Read /: [%s]",amsg)

		if err != nil {
			log.Printf("error: %v", err)

			switch r.URL.Path {
				case "/clonos/clonos/overview/":
					delete(clients_overview,ws)
				case "/clonos/clonos/settings/":
					delete(clients_settings,ws)
				case "/clonos/clonos/jailscontainers/":
					delete(clients_jailscontainers, ws)
				case "/clonos/clonos/instance_jail/":
					delete(clients_instance_jail, ws)
				case "/clonos/clonos/bhyvevms/":
					delete(clients_bhyvevms, ws)
				case "/clonos/clonos/vm_packages/":
					delete(clients_vm_packages, ws)
				case "/clonos/clonos/nodes/":
					delete(clients_nodes, ws)
				case "/clonos/clonos/vpnet/":
					delete(clients_vpnet, ws)
				case "/clonos/clonos/authkey/":
					delete(clients_authkey, ws)
				case "/clonos/clonos/media/":
					delete(clients_media, ws)
				case "/clonos/clonos/repo/":
					delete(clients_repo, ws)
				case "/clonos/clonos/bases/":
					delete(clients_bases, ws)
				case "/clonos/clonos/sources/":
					delete(clients_sources, ws)
				case "/clonos/clonos/jail_marketplace/":
					delete(clients_jail_marketplace, ws)
				case "/clonos/clonos/bhyve_marketplace/":
					delete(clients_bhyve_marketplace, ws)
				case "/clonos/clonos/tasklog/":
					delete(clients_tasklog, ws)
				case "/clonos/clonos/users/":
					delete(clients_users, ws)
				case "/clonos/clonos/imported/":
					delete(clients_imported, ws)
				case "/clonos/clonos/k8s/":
					delete(clients_k8s, ws)
				default:
					log.Println("Urouted URL: ",r.URL.Path)
					return
			}

			break
		}
		// Send the newly received message to the broadcast channel

		switch r.URL.Path {
			case "/clonos/clonos/settings/":
				broadcast_settings <- amsg
			case "/clonos/clonos/overview/":
				broadcast_overview <- amsg
			case "/clonos/clonos/jailscontainers/":
				broadcast_jailscontainers <- amsg
			case "/clonos/clonos/instance_jail/":
				broadcast_instance_jail <- amsg
			case "/clonos/clonos/bhyvevms/":
				broadcast_bhyvevms <- amsg
			case "/clonos/clonos/vm_packages/":
				broadcast_vm_packages <- amsg
			case "/clonos/clonos/nodes/":
				broadcast_nodes <- amsg
			case "/clonos/clonos/vpnet/":
				broadcast_vpnet <- amsg
			case "/clonos/clonos/authkey/":
				broadcast_authkey <- amsg
			case "/clonos/clonos/media/":
				broadcast_media <- amsg
			case "/clonos/clonos/repo/":
				broadcast_repo <- amsg
			case "/clonos/clonos/bases/":
				broadcast_bases <- amsg
			case "/clonos/clonos/sources/":
				broadcast_sources <- amsg
			case "/clonos/clonos/jail_marketplace/":
				broadcast_jail_marketplace <- amsg
			case "/clonos/clonos/bhyve_marketplace/":
				broadcast_bhyve_marketplace <- amsg
			case "/clonos/clonos/tasklog/":
				broadcast_tasklog <- amsg
			case "/clonos/clonos/users/":
				broadcast_users <- amsg
			case "/clonos/clonos/imported/":
				broadcast_imported <- amsg
			case "/clonos/clonos/k8s/":
				broadcast_k8s <- amsg
			default:
				log.Println("Urouted URL: ",r.URL.Path)
				return
		}
	}
}

func handleMessages_settings() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_settings

		// Send it out to every client that is currently connected
		for client := range clients_settings {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_settings , client)
			}
		}
	}
}

func handleMessages_overview() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_overview

		// Send it out to every client that is currently connected
		for client := range clients_overview {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_overview , client)
			}
		}
	}
}

func handleMessages_jailscontainers() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_jailscontainers

		// Send it out to every client that is currently connected
		for client := range clients_jailscontainers {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_jailscontainers, client)
			}
		}
	}
}

func handleMessages_instance_jail() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_instance_jail

		// Send it out to every client that is currently connected
		for client := range clients_instance_jail {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_instance_jail,client)
			}
		}
	}
}

func handleMessages_bhyvevms() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_bhyvevms

		// Send it out to every client that is currently connected
		for client := range clients_bhyvevms {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_bhyvevms,client)
			}
		}
	}
}

func handleMessages_vm_packages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_vm_packages

		// Send it out to every client that is currently connected
		for client := range clients_vm_packages {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_vm_packages,client)
			}
		}
	}
}



func handleMessages_nodes() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_nodes

		// Send it out to every client that is currently connected
		for client := range clients_nodes {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_nodes,client)
			}
		}
	}
}

func handleMessages_vpnet() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_vpnet

		// Send it out to every client that is currently connected
		for client := range clients_vpnet {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_vpnet,client)
			}
		}
	}
}

func handleMessages_authkey() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_authkey

		// Send it out to every client that is currently connected
		for client := range clients_authkey {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_authkey,client)
			}
		}
	}
}

func handleMessages_media() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_media

		// Send it out to every client that is currently connected
		for client := range clients_media {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_media,client)
			}
		}
	}
}

func handleMessages_repo() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_repo

		// Send it out to every client that is currently connected
		for client := range clients_repo {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_repo,client)
			}
		}
	}
}

func handleMessages_bases() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_bases

		// Send it out to every client that is currently connected
		for client := range clients_bases {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_bases,client)
			}
		}
	}
}

func handleMessages_sources() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_sources

		// Send it out to every client that is currently connected
		for client := range clients_sources {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_sources,client)
			}
		}
	}
}

func handleMessages_jail_marketplace() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_jail_marketplace

		// Send it out to every client that is currently connected
		for client := range clients_jail_marketplace {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_jail_marketplace,client)
			}
		}
	}
}

func handleMessages_bhyve_marketplace() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_bhyve_marketplace

		// Send it out to every client that is currently connected
		for client := range clients_bhyve_marketplace {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_bhyve_marketplace,client)
			}
		}
	}
}

func handleMessages_tasklog() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_tasklog

		// Send it out to every client that is currently connected
		for client := range clients_tasklog {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_tasklog,client)
			}
		}
	}
}

func handleMessages_users() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_users

		// Send it out to every client that is currently connected
		for client := range clients_users {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_users,client)
			}
		}
	}
}

func handleMessages_imported() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_imported

		// Send it out to every client that is currently connected
		for client := range clients_imported {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_imported,client)
			}
		}
	}
}

func handleMessages_k8s() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast_k8s

		// Send it out to every client that is currently connected
		for client := range clients_k8s {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients_k8s,client)
			}
		}
	}
}
