package handlers

import (
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

type WSClient struct {
	conn *websocket.Conn
	send chan []byte
	mu   sync.Mutex
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

	client := &WSClient{
		conn: conn,
		send: make(chan []byte, 256),
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
	case "execute_command":
		wsh.handleStreamingCommand(client, msg)
	case "get_conversations":
		wsh.handleGetConversations(client)
	case "get_conversation":
		wsh.handleGetConversation(client, msg)
	case "rename_conversation":
		wsh.handleRenameConversation(client, msg)
	case "kill_command":
		wsh.handleKillCommand(client, msg)
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

	wsh.sendToClient(client, models.WSMessage{
		Type:    "typing",
		Payload: map[string]bool{"typing": true},
	})

	response, err := wsh.agentSvc.ProcessMessage(models.AgentRequest{
		Message:        message,
		ConversationID: convID,
	})

	wsh.sendToClient(client, models.WSMessage{
		Type:    "typing",
		Payload: map[string]bool{"typing": false},
	})

	if err != nil {
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

func (wsh *WebSocketHandler) handleGetConversations(client *WSClient) {
	conversations, err := storage.Store.ListConversations()
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
	conv, err := storage.Store.LoadConversation(id)
	if err != nil {
		wsh.sendToClient(client, models.WSMessage{
			Type:    "error",
			Payload: map[string]string{"error": err.Error()},
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
	if err := storage.Store.DeleteConversation(id); err != nil {
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

	if err := storage.Store.RenameConversation(id, title); err != nil {
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

func (wsh *WebSocketHandler) sendToClient(client *WSClient, msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case client.send <- data:
	default:
	}
}

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
		}
		return true
	})
}
