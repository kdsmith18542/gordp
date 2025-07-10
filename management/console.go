package management

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ManagementConsole represents the management console
type ManagementConsole struct {
	servers      map[string]*ServerInfo
	users        map[string]*UserInfo
	sessions     map[string]*SessionInfo
	loadBalancer *LoadBalancer
	auditLogger  *AuditLogger
	recorder     *SessionRecorder
	config       *ConsoleConfig
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
}

// ConsoleConfig contains configuration for the management console
type ConsoleConfig struct {
	Port                int           `json:"port"`
	AdminPort           int           `json:"admin_port"`
	SessionTimeout      time.Duration `json:"session_timeout"`
	MaxSessions         int           `json:"max_sessions"`
	EnableLoadBalancing bool          `json:"enable_load_balancing"`
	EnableRecording     bool          `json:"enable_recording"`
	EnableAuditLog      bool          `json:"enable_audit_log"`
	DatabaseURL         string        `json:"database_url"`
	StaticPath          string        `json:"static_path"`
}

// DefaultConsoleConfig returns default console configuration
func DefaultConsoleConfig() *ConsoleConfig {
	return &ConsoleConfig{
		Port:                8080,
		AdminPort:           8081,
		SessionTimeout:      30 * time.Minute,
		MaxSessions:         1000,
		EnableLoadBalancing: true,
		EnableRecording:     true,
		EnableAuditLog:      true,
		DatabaseURL:         "sqlite://gordp.db",
		StaticPath:          "./management/web",
	}
}

// ServerInfo represents information about an RDP server
type ServerInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Port        int       `json:"port"`
	Status      string    `json:"status"`
	Load        float64   `json:"load"`
	Sessions    int       `json:"sessions"`
	MaxSessions int       `json:"max_sessions"`
	LastSeen    time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserInfo represents information about a user
type UserInfo struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	LastLogin   time.Time `json:"last_login"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

// SessionInfo represents information about an RDP session
type SessionInfo struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	ServerID      string        `json:"server_id"`
	Status        string        `json:"status"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time,omitempty"`
	Duration      time.Duration `json:"duration,omitempty"`
	IPAddress     string        `json:"ip_address"`
	UserAgent     string        `json:"user_agent"`
	IsRecorded    bool          `json:"is_recorded"`
	RecordingPath string        `json:"recording_path,omitempty"`
}

// NewManagementConsole creates a new management console
func NewManagementConsole(cfg *ConsoleConfig) *ManagementConsole {
	if cfg == nil {
		cfg = DefaultConsoleConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	console := &ManagementConsole{
		servers:  make(map[string]*ServerInfo),
		users:    make(map[string]*UserInfo),
		sessions: make(map[string]*SessionInfo),
		config:   cfg,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Initialize components
	if cfg.EnableLoadBalancing {
		console.loadBalancer = NewLoadBalancer()
	}

	if cfg.EnableAuditLog {
		console.auditLogger = NewAuditLogger()
	}

	if cfg.EnableRecording {
		console.recorder = NewSessionRecorder()
	}

	return console
}

// Start starts the management console
func (mc *ManagementConsole) Start() error {
	// Set up HTTP routes
	http.HandleFunc("/api/servers", mc.handleServersAPI)
	http.HandleFunc("/api/users", mc.handleUsersAPI)
	http.HandleFunc("/api/sessions", mc.handleSessionsAPI)
	http.HandleFunc("/api/connect", mc.handleConnectAPI)
	http.HandleFunc("/api/disconnect", mc.handleDisconnectAPI)
	http.HandleFunc("/api/stats", mc.handleStatsAPI)
	http.HandleFunc("/api/audit", mc.handleAuditAPI)

	// Serve static files for web interface
	if mc.config.StaticPath != "" {
		http.Handle("/", http.FileServer(http.Dir(mc.config.StaticPath)))
	}

	// Start background tasks
	go mc.cleanupSessions()
	go mc.monitorServers()

	log.Printf("Management Console starting on port %d", mc.config.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", mc.config.Port), nil)
}

// Stop stops the management console
func (mc *ManagementConsole) Stop() {
	mc.cancel()

	// Close all sessions
	mc.mu.Lock()
	for _, session := range mc.sessions {
		mc.DisconnectSession(session.ID)
	}
	mc.mu.Unlock()
}

// AddServer adds a server to the management console
func (mc *ManagementConsole) AddServer(server *ServerInfo) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.servers[server.ID]; exists {
		return fmt.Errorf("server with ID %s already exists", server.ID)
	}

	server.CreatedAt = time.Now()
	server.LastSeen = time.Now()
	mc.servers[server.ID] = server

	if mc.auditLogger != nil {
		mc.auditLogger.Log("server_added", map[string]interface{}{
			"server_id": server.ID,
			"name":      server.Name,
			"address":   server.Address,
		})
	}

	return nil
}

// RemoveServer removes a server from the management console
func (mc *ManagementConsole) RemoveServer(serverID string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.servers[serverID]; !exists {
		return fmt.Errorf("server with ID %s not found", serverID)
	}

	delete(mc.servers, serverID)

	if mc.auditLogger != nil {
		mc.auditLogger.Log("server_removed", map[string]interface{}{
			"server_id": serverID,
		})
	}

	return nil
}

// AddUser adds a user to the management console
func (mc *ManagementConsole) AddUser(user *UserInfo) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.users[user.ID]; exists {
		return fmt.Errorf("user with ID %s already exists", user.ID)
	}

	user.CreatedAt = time.Now()
	user.IsActive = true
	mc.users[user.ID] = user

	if mc.auditLogger != nil {
		mc.auditLogger.Log("user_added", map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"role":     user.Role,
		})
	}

	return nil
}

// RemoveUser removes a user from the management console
func (mc *ManagementConsole) RemoveUser(userID string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.users[userID]; !exists {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	delete(mc.users, userID)

	if mc.auditLogger != nil {
		mc.auditLogger.Log("user_removed", map[string]interface{}{
			"user_id": userID,
		})
	}

	return nil
}

// CreateSession creates a new RDP session
func (mc *ManagementConsole) CreateSession(userID, serverID, ipAddress, userAgent string) (*SessionInfo, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if user exists
	if _, exists := mc.users[userID]; !exists {
		return nil, fmt.Errorf("user with ID %s not found", userID)
	}

	// Check if server exists
	if _, exists := mc.servers[serverID]; !exists {
		return nil, fmt.Errorf("server with ID %s not found", serverID)
	}

	// Check session limits
	if len(mc.sessions) >= mc.config.MaxSessions {
		return nil, fmt.Errorf("maximum sessions reached")
	}

	sessionID := generateSessionID()
	session := &SessionInfo{
		ID:        sessionID,
		UserID:    userID,
		ServerID:  serverID,
		Status:    "connecting",
		StartTime: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	mc.sessions[sessionID] = session

	if mc.auditLogger != nil {
		mc.auditLogger.Log("session_created", map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
			"server_id":  serverID,
			"ip_address": ipAddress,
		})
	}

	return session, nil
}

// DisconnectSession disconnects a session
func (mc *ManagementConsole) DisconnectSession(sessionID string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	session, exists := mc.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session with ID %s not found", sessionID)
	}

	session.Status = "disconnected"
	session.EndTime = time.Now()
	session.Duration = session.EndTime.Sub(session.StartTime)

	if mc.auditLogger != nil {
		mc.auditLogger.Log("session_disconnected", map[string]interface{}{
			"session_id": sessionID,
			"duration":   session.Duration.String(),
		})
	}

	return nil
}

// GetSessionStats returns session statistics
func (mc *ManagementConsole) GetSessionStats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := map[string]interface{}{
		"total_sessions":        len(mc.sessions),
		"active_sessions":       0,
		"disconnected_sessions": 0,
		"total_servers":         len(mc.servers),
		"total_users":           len(mc.users),
	}

	for _, session := range mc.sessions {
		if session.Status == "connected" {
			stats["active_sessions"] = stats["active_sessions"].(int) + 1
		} else if session.Status == "disconnected" {
			stats["disconnected_sessions"] = stats["disconnected_sessions"].(int) + 1
		}
	}

	return stats
}

// cleanupSessions periodically cleans up old sessions
func (mc *ManagementConsole) cleanupSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.mu.Lock()
			now := time.Now()
			for id, session := range mc.sessions {
				if session.Status == "disconnected" &&
					now.Sub(session.EndTime) > mc.config.SessionTimeout {
					delete(mc.sessions, id)
				}
			}
			mc.mu.Unlock()
		}
	}
}

// monitorServers periodically monitors server health
func (mc *ManagementConsole) monitorServers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.mu.Lock()
			now := time.Now()
			for _, server := range mc.servers {
				if now.Sub(server.LastSeen) > 2*time.Minute {
					server.Status = "offline"
				}
			}
			mc.mu.Unlock()
		}
	}
}

// API handlers
func (mc *ManagementConsole) handleServersAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mc.mu.RLock()
		servers := make([]*ServerInfo, 0, len(mc.servers))
		for _, server := range mc.servers {
			servers = append(servers, server)
		}
		mc.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"servers": servers,
			"count":   len(servers),
		})

	case http.MethodPost:
		var server ServerInfo
		if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := mc.AddServer(&server); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (mc *ManagementConsole) handleUsersAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mc.mu.RLock()
		users := make([]*UserInfo, 0, len(mc.users))
		for _, user := range mc.users {
			users = append(users, user)
		}
		mc.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
			"count": len(users),
		})

	case http.MethodPost:
		var user UserInfo
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := mc.AddUser(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (mc *ManagementConsole) handleSessionsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mc.mu.RLock()
	sessions := make([]*SessionInfo, 0, len(mc.sessions))
	for _, session := range mc.sessions {
		sessions = append(sessions, session)
	}
	mc.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

func (mc *ManagementConsole) handleConnectAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID    string `json:"user_id"`
		ServerID  string `json:"server_id"`
		IPAddress string `json:"ip_address"`
		UserAgent string `json:"user_agent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	session, err := mc.CreateSession(req.UserID, req.ServerID, req.IPAddress, req.UserAgent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (mc *ManagementConsole) handleDisconnectAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "session ID required", http.StatusBadRequest)
		return
	}

	if err := mc.DisconnectSession(sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (mc *ManagementConsole) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := mc.GetSessionStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (mc *ManagementConsole) handleAuditAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if mc.auditLogger == nil {
		http.Error(w, "audit logging not enabled", http.StatusServiceUnavailable)
		return
	}

	logs := mc.auditLogger.GetLogs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// LoadBalancer represents a load balancer for RDP servers
type LoadBalancer struct {
	algorithm string
	mu        sync.RWMutex
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		algorithm: "round_robin",
	}
}

// SelectServer selects the best server based on load balancing algorithm
func (lb *LoadBalancer) SelectServer(servers []*ServerInfo) *ServerInfo {
	if len(servers) == 0 {
		return nil
	}

	// Simple round-robin implementation
	// In a real implementation, this would be more sophisticated
	return servers[0]
}

// AuditLogger represents an audit logger
type AuditLogger struct {
	logs []AuditLog
	mu   sync.RWMutex
}

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"`
	UserID    string                 `json:"user_id,omitempty"`
	Details   map[string]interface{} `json:"details"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make([]AuditLog, 0),
	}
}

// Log logs an audit event
func (al *AuditLogger) Log(action string, details map[string]interface{}) {
	al.mu.Lock()
	defer al.mu.Unlock()

	log := AuditLog{
		Timestamp: time.Now(),
		Action:    action,
		Details:   details,
	}

	al.logs = append(al.logs, log)
}

// GetLogs returns all audit logs
func (al *AuditLogger) GetLogs() []AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	logs := make([]AuditLog, len(al.logs))
	copy(logs, al.logs)
	return logs
}

// SessionRecorder represents a session recorder
type SessionRecorder struct {
	recordings map[string]*Recording
	mu         sync.RWMutex
}

// Recording represents a session recording
type Recording struct {
	SessionID  string    `json:"session_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time,omitempty"`
	FilePath   string    `json:"file_path"`
	Size       int64     `json:"size"`
	IsComplete bool      `json:"is_complete"`
}

// NewSessionRecorder creates a new session recorder
func NewSessionRecorder() *SessionRecorder {
	return &SessionRecorder{
		recordings: make(map[string]*Recording),
	}
}

// StartRecording starts recording a session
func (sr *SessionRecorder) StartRecording(sessionID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.recordings[sessionID]; exists {
		return fmt.Errorf("recording already exists for session %s", sessionID)
	}

	recording := &Recording{
		SessionID: sessionID,
		StartTime: time.Now(),
		FilePath:  fmt.Sprintf("recordings/%s.rec", sessionID),
	}

	sr.recordings[sessionID] = recording
	return nil
}

// StopRecording stops recording a session
func (sr *SessionRecorder) StopRecording(sessionID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	recording, exists := sr.recordings[sessionID]
	if !exists {
		return fmt.Errorf("no recording found for session %s", sessionID)
	}

	recording.EndTime = time.Now()
	recording.IsComplete = true
	return nil
}

// GetRecording returns a recording
func (sr *SessionRecorder) GetRecording(sessionID string) (*Recording, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	recording, exists := sr.recordings[sessionID]
	if !exists {
		return nil, fmt.Errorf("no recording found for session %s", sessionID)
	}

	return recording, nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
