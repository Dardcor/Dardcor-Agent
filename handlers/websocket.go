package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"dardcor-agent/models"
	"dardcor-agent/services"
	"dardcor-agent/storage"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSClient represents a single connected WebSocket client.
// send is a buffered channel; the writePump drains it onto the wire.
// agentCancel cancels any in-flight agentic loop for this client so that
// incoming messages (e.g. interrupt / new prompt) are never blocked.
type WSClient struct {
	conn        *websocket.Conn
	send        chan []byte
	mu          sync.Mutex
	agentCancel context.CancelFunc
	agentMu     sync.Mutex // guards agentCancel
}

// cancelAgent cancels the currently running agentic loop (if any) and
// stores the new cancel function so the next loop can be cancelled later.
func (c *WSClient) cancelAgent(newCancel context.CancelFunc) {
	c.agentMu.Lock()
	defer c.agentMu.Unlock()
	if c.agentCancel != nil {
		c.agentCancel() // cancel previous loop
	}
	c.agentCancel = newCancel
}

// stopAgent cancels any running agentic loop and clears the stored cancel.
func (c *WSClient) stopAgent() {
	c.agentMu.Lock()
	defer c.agentMu.Unlock()
	if c.agentCancel != nil {
		c.agentCancel()
		c.agentCancel = nil
	}
}

type WebSocketHandler struct {
	clients    sync.Map
	agentSvc   *services.AgentService
	cmdService *services.CommandService
}

func NewWebSocketHandler(agentSvc *services.AgentService, cmdService *services.CommandService) *WebSocketHandler {
	return &WebSocketHandler{
		agentSvc:   agentSvc,
		cmdService: cmdService,
	}
}

func (wsh *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS Error: %v", err)
		return
	}

	// Buffer size 512: large enough to absorb burst output during a long agent
	// loop without dropping messages, but bounded to avoid unbounded memory use.
	client := &WSClient{
		conn: conn,
		send: make(chan []byte, 512),
	}

	clientID := r.RemoteAddr
	wsh.clients.Store(clientID, client)

	wsh.sendToClient(client, models.WSMessage{
		Type: "connected",
		Payload: map[string]string{
			"message":   "Connected to Dardcor Agent",
			"client_id": clientID,
		},
	})

	go wsh.writePump(client, clientID)
	wsh.readPump(client, clientID)
}

func (wsh *WebSocketHandler) readPump(client *WSClient, clientID string) {
	defer func() {
		// Cancel any in-flight agent loop before cleaning up.
		client.stopAgent()
		wsh.clients.Delete(clientID)
		client.conn.Close()
		close(client.send)
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			wsh.sendToClient(client, models.WSMessage{
				Type: "error",
				Payload: map[string]string{
					"error": "Invalid message format",
				},
			})
			continue
		}

		// handleMessage is non-blocking for every message type — the agent
		// path uses its own goroutine so the read loop is never stalled.
		go wsh.handleMessage(client, wsMsg)
	}
}

func (wsh *WebSocketHandler) writePump(client *WSClient, clientID string) {
	for message := range client.send {
		client.mu.Lock()
		err := client.conn.WriteMessage(websocket.TextMessage, message)
		client.mu.Unlock()

		if err != nil {
			return
		}
	}
}

func (wsh *WebSocketHandler) handleMessage(client *WSClient, msg models.WSMessage) {
	switch msg.Type {
	case "agent_message":
		wsh.handleAgentMessage(client, msg)
	case "stop_agent":
		// Explicit stop signal from the client — cancel any running loop.
		client.stopAgent()
		wsh.sendToClient(client, models.WSMessage{
			Type:    "agent_stopped",
			Payload: map[string]string{"status": "stopped"},
		})
	case "execute_command":
		wsh.handleStreamingCommand(client, msg)
	case "create_conversation":
		wsh.handleCreateConversation(client, msg)
	case "get_conversations":
		wsh.handleGetConversations(client)
	case "get_conversation":
		wsh.handleGetConversation(client, msg)
	case "rename_conversation":
		wsh.handleRenameConversation(client, msg)
	case "delete_conversation":
		wsh.handleDeleteConversation(client, msg)
	case "kill_command":
		wsh.handleKillCommand(client, msg)
	case "delete_conversation":
		wsh.handleDeleteConversation(client, msg)
	case "ping":
		wsh.sendToClient(client, models.WSMessage{
			Type:    "pong",
			Payload: map[string]string{"status": "alive"},
		})
	default:
		wsh.sendToClient(client, models.WSMessage{
			Type: "error",
			Payload: map[string]string{
				"error": "Unknown message type: " + msg.Type,
			},
		})
	}
}

// handleAgentMessage runs the agentic loop in a dedicated goroutine so the
// WebSocket read loop stays unblocked and can immediately process any incoming
// message (including stop/interrupt signals) while the LLM is thinking.
func (wsh *WebSocketHandler) handleAgentMessage(client *WSClient, msg models.WSMessage) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "error",
			Payload: map[string]string{"error": "Invalid payload"},
		})
		return
	}

	message, _ := payload["message"].(string)
	convID, _ := payload["conversation_id"].(string)

	// Create a cancellable context for this agent run.
	// cancelAgent cancels any previously running loop before storing the new one.
	ctx, cancel := context.WithCancel(context.Background())
	client.cancelAgent(cancel)

	// Run the agentic loop in its own goroutine so the WS reader is never
	// blocked waiting for LLM responses (which can take tens of seconds per turn).
	go func() {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "typing",
			Payload: map[string]bool{"typing": true},
		})

		response, err := wsh.agentSvc.ProcessMessage(ctx, models.AgentRequest{
			Message:        message,
			ConversationID: convID,
		}, func(part *models.AgentResponse) {
			wsh.sendToClient(client, models.WSMessage{
				Type:    "agent_turn",
				Payload: part,
			})
		})

		wsh.sendToClient(client, models.WSMessage{
			Type:    "typing",
			Payload: map[string]bool{"typing": false},
		})

		if err != nil {
			// context.Canceled means the user (or a new message) interrupted the loop.
			if err == context.Canceled {
				return
			}
			wsh.sendToClient(client, models.WSMessage{
				Type: "error",
				Payload: map[string]string{
					"error": err.Error(),
				},
			})
			return
		}

		wsh.sendToClient(client, models.WSMessage{
			Type:    "agent_response",
			Payload: response,
		})
	}()
}

func (wsh *WebSocketHandler) handleStreamingCommand(client *WSClient, msg models.WSMessage) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	command, _ := payload["command"].(string)
	workingDir, _ := payload["working_dir"].(string)

	shell, _ := payload["shell"].(string)

	req := models.CommandRequest{
		Command:    command,
		Shell:      shell,
		WorkingDir: workingDir,
		Timeout:    60,
	}

	result, err := wsh.cmdService.ExecuteCommandStreaming(req, func(id, output string, isError bool) {
		wsh.sendToClient(client, models.WSMessage{
			Type: "command_output",
			Payload: models.WSCommandOutput{
				CommandID: id,
				Output:    output,
				IsError:   isError,
				Done:      false,
			},
		})
	})

	if err != nil {
		wsh.sendToClient(client, models.WSMessage{
			Type: "command_output",
			Payload: models.WSCommandOutput{
				Output:  err.Error(),
				IsError: true,
				Done:    true,
			},
		})
		return
	}

	wsh.sendToClient(client, models.WSMessage{
		Type:    "command_complete",
		Payload: result,
	})
}

func (wsh *WebSocketHandler) handleKillCommand(client *WSClient, msg models.WSMessage) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	id, _ := payload["id"].(string)
	if id != "" {
		wsh.cmdService.KillCommand(id)
	}
}

func (wsh *WebSocketHandler) handleCreateConversation(client *WSClient, msg models.WSMessage) {
	payload, _ := msg.Payload.(map[string]interface{})
	title, _ := payload["title"].(string)
	if title == "" {
		title = "Percakapan Baru"
	}

	conv, err := storage.Store.CreateConversation(title, "web")
	if err != nil {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "error",
			Payload: map[string]string{"error": err.Error()},
		})
		return
	}

	wsh.sendToClient(client, models.WSMessage{
		Type:    "conversation_created",
		Payload: conv,
	})
}

func (wsh *WebSocketHandler) handleGetConversations(client *WSClient) {

	conversations, err := storage.Store.ListConversations("web")
	if err != nil {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "error",
			Payload: map[string]string{"error": err.Error()},
		})
		return
	}

	wsh.sendToClient(client, models.WSMessage{
		Type:    "conversations_list",
		Payload: conversations,
	})
}

func (wsh *WebSocketHandler) handleGetConversation(client *WSClient, msg models.WSMessage) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	id, _ := payload["id"].(string)
	conv, err := storage.Store.LoadConversation(id, "web")
	if err != nil {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "error",
			Payload: map[string]string{"error": "Session not found or expired."},
		})
		return
	}

	wsh.sendToClient(client, models.WSMessage{
		Type:    "conversation_detail",
		Payload: conv,
	})
}

func (wsh *WebSocketHandler) handleDeleteConversation(client *WSClient, msg models.WSMessage) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	id, _ := payload["id"].(string)
	if err := storage.Store.DeleteConversation(id, "web"); err != nil {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "error",
			Payload: map[string]string{"error": err.Error()},
		})
		return
	}

	wsh.sendToClient(client, models.WSMessage{
		Type:    "conversation_deleted",
		Payload: map[string]string{"id": id},
	})

	wsh.handleGetConversations(client)
}

func (wsh *WebSocketHandler) handleRenameConversation(client *WSClient, msg models.WSMessage) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	id, _ := payload["id"].(string)
	title, _ := payload["title"].(string)

	if err := storage.Store.RenameConversation(id, title, "web"); err != nil {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "error",
			Payload: map[string]string{"error": err.Error()},
		})
		return
	}

	wsh.sendToClient(client, models.WSMessage{
		Type: "conversation_renamed",
		Payload: map[string]string{
			"id":    id,
			"title": title,
		},
	})

	wsh.handleGetConversations(client)
}

// sendToClient marshals msg and enqueues it on the client's send channel.
// If the channel is full the message is dropped and a warning is logged so
// that silent drops become visible in server logs.
func (wsh *WebSocketHandler) sendToClient(client *WSClient, msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case client.send <- data:
	default:
		log.Printf("[WS] WARNING: send channel full for client, dropping message type=%q (buffer=%d)", msg.Type, cap(client.send))
	}
}

// Broadcast sends msg to every connected client.
// If a client's send channel is full the message is dropped and logged rather
// than silently discarded.
func (wsh *WebSocketHandler) Broadcast(msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	wsh.clients.Range(func(key, value interface{}) bool {
		client := value.(*WSClient)
		select {
		case client.send <- data:
		default:
			log.Printf("[WS] WARNING: broadcast channel full for client %v, dropping message type=%q", key, msg.Type)
		}
		return true
	})
}
