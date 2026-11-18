package backend

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const playbackHeartbeatTimeout = 3 * time.Minute

type streamKey struct {
	UserID   string
	ServerID string
}

type streamEntry struct {
	ItemID        string
	LastHeartbeat time.Time
}

type ActiveSession struct {
	ID            string    `json:"id"`
	SessionKey    string    `json:"sessionKey"`
	UserID        string    `json:"userId"`
	Username      string    `json:"username"`
	Client        string    `json:"client"`
	ClientVersion string    `json:"clientVersion"`
	DeviceName    string    `json:"deviceName"`
	DeviceID      string    `json:"deviceId"`
	IP            string    `json:"ip"`
	ServerID      string    `json:"serverId"`
	ServerName    string    `json:"serverName"`
	ItemID        string    `json:"itemId"`
	ItemName      string    `json:"itemName"`
	ItemType      string    `json:"itemType"`
	PositionTicks int64     `json:"positionTicks"`
	RunTimeTicks  int64     `json:"runTimeTicks"`
	IsPaused      bool      `json:"isPaused"`
	PlayMethod    string    `json:"playMethod"`
	StartedAt     time.Time `json:"startedAt"`
	LastHeartbeat time.Time `json:"lastHeartbeat"`
}

type PlaybackLimiter struct {
	mu             sync.Mutex
	streams        map[streamKey]*streamEntry
	sessions       map[string]*ActiveSession // 键为 SessionKey (UserID:DeviceID)
	sessionByID    map[string]*ActiveSession // 键为 ID (随机唯一标识)
	kickedSessions map[string]time.Time      // 键为 SessionKey -> 封禁过期时间
}

func NewPlaybackLimiter() *PlaybackLimiter {
	return &PlaybackLimiter{
		streams:        make(map[streamKey]*streamEntry),
		sessions:       make(map[string]*ActiveSession),
		sessionByID:    make(map[string]*ActiveSession),
		kickedSessions: make(map[string]time.Time),
	}
}

// TryStart 兼容旧版测试的播放流尝试注册接口
func (l *PlaybackLimiter) TryStart(userID string, serverID string, itemID string, maxConcurrent int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if maxConcurrent <= 0 {
		key := streamKey{UserID: userID, ServerID: serverID}
		l.streams[key] = &streamEntry{
			ItemID:        itemID,
			LastHeartbeat: time.Now(),
		}
		return true
	}

	key := streamKey{UserID: userID, ServerID: serverID}

	if existing, ok := l.streams[key]; ok {
		if time.Since(existing.LastHeartbeat) >= playbackHeartbeatTimeout {
			delete(l.streams, key)
		} else {
			existing.ItemID = itemID
			existing.LastHeartbeat = time.Now()
			return true
		}
	}

	now := time.Now()
	count := 0
	for k, entry := range l.streams {
		if k.ServerID == serverID && now.Sub(entry.LastHeartbeat) < playbackHeartbeatTimeout {
			count++
		}
	}

	if count >= maxConcurrent {
		return false
	}

	l.streams[key] = &streamEntry{
		ItemID:        itemID,
		LastHeartbeat: now,
	}
	return true
}

// TryStartSession 综合校验服务器并发与用户并发限制，注册或更新会话
func (l *PlaybackLimiter) TryStartSession(sess *ActiveSession, serverMaxConcurrent int, userMaxConcurrent int) (bool, string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// 1. 检查是否处于刚被踢出的短时防抖期内
	if kickUntil, ok := l.kickedSessions[sess.SessionKey]; ok {
		if now.Before(kickUntil) {
			return false, "该会话已被管理员终止，请稍候 5 秒后再试"
		}
		delete(l.kickedSessions, sess.SessionKey)
	}

	// 2. 同一用户在同一设备上切换媒体：直接复用已有插槽并刷新心跳
	if existing, ok := l.sessions[sess.SessionKey]; ok {
		if now.Sub(existing.LastHeartbeat) < playbackHeartbeatTimeout {
			existing.ItemID = sess.ItemID
			if sess.ItemName != "" {
				existing.ItemName = sess.ItemName
			}
			if sess.ItemType != "" {
				existing.ItemType = sess.ItemType
			}
			if sess.ServerID != "" {
				existing.ServerID = sess.ServerID
			}
			if sess.ServerName != "" {
				existing.ServerName = sess.ServerName
			}
			if sess.PlayMethod != "" {
				existing.PlayMethod = sess.PlayMethod
			}
			if sess.IP != "" {
				existing.IP = sess.IP
			}
			existing.LastHeartbeat = now

			l.streams[streamKey{UserID: sess.UserID, ServerID: sess.ServerID}] = &streamEntry{
				ItemID:        sess.ItemID,
				LastHeartbeat: now,
			}
			return true, ""
		}
		// 过期残留会话：先清理再按新会话评估
		l.deleteSessionLocked(sess.SessionKey)
	}

	// 3. 服务器总并发校验
	if serverMaxConcurrent > 0 {
		serverCount := l.countActiveForServerLocked(sess.ServerID)
		if serverCount >= serverMaxConcurrent {
			return false, fmt.Sprintf("已达到服务器最大同时播放数限制 (%d)", serverMaxConcurrent)
		}
	}

	// 4. 单用户跨设备并发限制校验
	if userMaxConcurrent > 0 {
		userCount := l.countActiveForUserLocked(sess.UserID)
		if userCount >= userMaxConcurrent {
			return false, fmt.Sprintf("已达到账号最大同时播放设备限制 (%d 台)", userMaxConcurrent)
		}
	}

	// 5. 登记新播放会话
	if sess.ID == "" {
		sess.ID = randomHex(16)
	}
	sess.StartedAt = now
	sess.LastHeartbeat = now
	l.sessions[sess.SessionKey] = sess
	l.sessionByID[sess.ID] = sess
	l.streams[streamKey{UserID: sess.UserID, ServerID: sess.ServerID}] = &streamEntry{
		ItemID:        sess.ItemID,
		LastHeartbeat: now,
	}

	return true, ""
}

// RecordAdminSession 记录管理员的播放会话，不施加并发拦截
func (l *PlaybackLimiter) RecordAdminSession(sess *ActiveSession) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if existing, ok := l.sessions[sess.SessionKey]; ok {
		existing.ItemID = sess.ItemID
		if sess.ItemName != "" {
			existing.ItemName = sess.ItemName
		}
		if sess.ItemType != "" {
			existing.ItemType = sess.ItemType
		}
		if sess.ServerID != "" {
			existing.ServerID = sess.ServerID
		}
		if sess.ServerName != "" {
			existing.ServerName = sess.ServerName
		}
		if sess.PlayMethod != "" {
			existing.PlayMethod = sess.PlayMethod
		}
		if sess.IP != "" {
			existing.IP = sess.IP
		}
		existing.LastHeartbeat = now
		return
	}

	if sess.ID == "" {
		sess.ID = randomHex(16)
	}
	sess.StartedAt = now
	sess.LastHeartbeat = now
	l.sessions[sess.SessionKey] = sess
	l.sessionByID[sess.ID] = sess
}

// Heartbeat 刷新用户在指定上游服务器上的流心跳
func (l *PlaybackLimiter) Heartbeat(userID string, serverID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	key := streamKey{UserID: userID, ServerID: serverID}
	if entry, ok := l.streams[key]; ok {
		entry.LastHeartbeat = now
	}
	for _, sess := range l.sessions {
		if sess.UserID == userID && sess.ServerID == serverID {
			sess.LastHeartbeat = now
		}
	}
}

// UpdateSessionProgress 更新指定会话的播放进度、暂停状态及媒体信息
func (l *PlaybackLimiter) UpdateSessionProgress(sessionKey string, positionTicks, runtimeTicks int64, isPaused bool, itemName, itemType string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if sess, ok := l.sessions[sessionKey]; ok {
		sess.PositionTicks = positionTicks
		if runtimeTicks > 0 {
			sess.RunTimeTicks = runtimeTicks
		}
		sess.IsPaused = isPaused
		if itemName != "" {
			sess.ItemName = itemName
		}
		if itemType != "" {
			sess.ItemType = itemType
		}
		sess.LastHeartbeat = now

		key := streamKey{UserID: sess.UserID, ServerID: sess.ServerID}
		if entry, ok := l.streams[key]; ok {
			entry.LastHeartbeat = now
		}
	}
}

// Stop 移除用户在指定服务器上的播放记录
func (l *PlaybackLimiter) Stop(userID string, serverID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.streams, streamKey{UserID: userID, ServerID: serverID})
	for key, sess := range l.sessions {
		if sess.UserID == userID && (serverID == "" || sess.ServerID == serverID) {
			delete(l.sessionByID, sess.ID)
			delete(l.sessions, key)
		}
	}
}

// StopSession 按会话键移除活动会话
func (l *PlaybackLimiter) StopSession(sessionKey string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.deleteSessionLocked(sessionKey)
}

func (l *PlaybackLimiter) deleteSessionLocked(sessionKey string) {
	if sess, ok := l.sessions[sessionKey]; ok {
		delete(l.sessionByID, sess.ID)
		delete(l.streams, streamKey{UserID: sess.UserID, ServerID: sess.ServerID})
		delete(l.sessions, sessionKey)
	}
}

// KickSession 按会话 ID 强制踢出会话
func (l *PlaybackLimiter) KickSession(sessionID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	sess, ok := l.sessionByID[sessionID]
	if !ok {
		return false
	}
	sessionKey := sess.SessionKey
	// 5 秒短时防抖：仅阻断当前播放器后台自动重试，不长期封禁用户设备
	l.kickedSessions[sessionKey] = time.Now().Add(5 * time.Second)
	l.deleteSessionLocked(sessionKey)
	return true
}

// ClearKick 解除指定会话或设备的踢出状态
func (l *PlaybackLimiter) ClearKick(sessionKey string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.kickedSessions, sessionKey)
}

// ClearAllKicks 清空所有踢出限制
func (l *PlaybackLimiter) ClearAllKicks() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.kickedSessions = make(map[string]time.Time)
}

// IsKicked 查询设备/会话当前是否处于踢出防抖期内
func (l *PlaybackLimiter) IsKicked(sessionKey string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if until, ok := l.kickedSessions[sessionKey]; ok {
		if time.Now().Before(until) {
			return true
		}
		delete(l.kickedSessions, sessionKey)
	}
	return false
}

// GetActiveSessions 获取所有活跃（未超时）的播放会话列表，按最近活跃时间倒序排列
func (l *PlaybackLimiter) GetActiveSessions() []*ActiveSession {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	result := make([]*ActiveSession, 0, len(l.sessions))
	for _, sess := range l.sessions {
		if now.Sub(sess.LastHeartbeat) < playbackHeartbeatTimeout {
			sessCopy := *sess
			result = append(result, &sessCopy)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastHeartbeat.After(result[j].LastHeartbeat)
	})
	return result
}

// CountForUser 统计指定用户的活跃在线播放设备数
func (l *PlaybackLimiter) CountForUser(userID string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.countActiveForUserLocked(userID)
}

func (l *PlaybackLimiter) countActiveForUserLocked(userID string) int {
	now := time.Now()
	count := 0
	for _, sess := range l.sessions {
		if sess.UserID == userID && now.Sub(sess.LastHeartbeat) < playbackHeartbeatTimeout {
			count++
		}
	}
	return count
}

// CountForServer 统计指定服务器的活跃流数量
func (l *PlaybackLimiter) CountForServer(serverID string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.countActiveForServerLocked(serverID)
}

func (l *PlaybackLimiter) countActiveForServerLocked(serverID string) int {
	now := time.Now()
	count := 0
	for _, sess := range l.sessions {
		if sess.ServerID == serverID && now.Sub(sess.LastHeartbeat) < playbackHeartbeatTimeout {
			count++
		}
	}
	if len(l.sessions) == 0 {
		for k, entry := range l.streams {
			if k.ServerID == serverID && now.Sub(entry.LastHeartbeat) < playbackHeartbeatTimeout {
				count++
			}
		}
	}
	return count
}

// Cleanup 定期清理超时失效的流、会话以及踢出封禁黑名单
func (l *PlaybackLimiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for key, entry := range l.streams {
		if now.Sub(entry.LastHeartbeat) >= playbackHeartbeatTimeout {
			delete(l.streams, key)
		}
	}
	for key, sess := range l.sessions {
		if now.Sub(sess.LastHeartbeat) >= playbackHeartbeatTimeout {
			delete(l.sessionByID, sess.ID)
			delete(l.sessions, key)
		}
	}
	for key, until := range l.kickedSessions {
		if now.After(until) {
			delete(l.kickedSessions, key)
		}
	}
}

func buildSessionKey(userID, deviceID string) string {
	if userID == "" {
		userID = "guest"
	}
	if deviceID == "" {
		deviceID = "default"
	}
	return userID + ":" + deviceID
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

// extractClientDeviceInfo 从 HTTP 请求头、复合认证头或 URL 查询参数中提取客户端及设备元数据
func extractClientDeviceInfo(r *http.Request) (client, clientVersion, deviceName, deviceID string) {
	if r == nil {
		return "Unknown", "", "Unknown Device", "unknown"
	}

	client = r.Header.Get("X-Emby-Client")
	clientVersion = r.Header.Get("X-Emby-Client-Version")
	deviceName = r.Header.Get("X-Emby-Device-Name")
	deviceID = r.Header.Get("X-Emby-Device-Id")

	if client == "" || clientVersion == "" || deviceName == "" || deviceID == "" {
		for _, authKey := range []string{"X-Emby-Authorization", "Authorization"} {
			authVal := r.Header.Get(authKey)
			if authVal == "" {
				continue
			}
			parsed, ok := parseAuthorizationIdentityStrict(authVal)
			if !ok {
				continue
			}
			if client == "" {
				client = parsed["Client"]
			}
			if clientVersion == "" {
				clientVersion = parsed["Version"]
			}
			if deviceName == "" {
				deviceName = parsed["Device"]
			}
			if deviceID == "" {
				deviceID = parsed["DeviceId"]
			}
			break
		}
	}

	q := r.URL.Query()
	if client == "" {
		client = q.Get("X-Emby-Client")
		if client == "" {
			client = q.Get("Client")
		}
	}
	if clientVersion == "" {
		clientVersion = q.Get("X-Emby-Client-Version")
		if clientVersion == "" {
			clientVersion = q.Get("Version")
		}
	}
	if deviceName == "" {
		deviceName = q.Get("X-Emby-Device-Name")
		if deviceName == "" {
			deviceName = q.Get("Device")
		}
	}
	if deviceID == "" {
		deviceID = q.Get("X-Emby-Device-Id")
		if deviceID == "" {
			deviceID = q.Get("DeviceId")
		}
	}

	if client == "" {
		ua := r.Header.Get("User-Agent")
		if strings.Contains(ua, "Hills") {
			client = "Hills Lite"
		} else if strings.Contains(ua, "Infuse") {
			client = "Infuse"
		} else if strings.Contains(ua, "Mozilla") || strings.Contains(ua, "Chrome") || strings.Contains(ua, "Safari") {
			client = "Emby Web"
		} else if ua != "" {
			client = ua
		} else {
			client = "Emby Client"
		}
	}
	if deviceName == "" {
		deviceName = client
	}
	if deviceID == "" {
		remote := r.RemoteAddr
		if colon := strings.LastIndex(remote, ":"); colon != -1 {
			remote = remote[:colon]
		}
		deviceID = "dev_" + hashString(remote + "_" + r.Header.Get("User-Agent"))
	}

	return client, clientVersion, deviceName, deviceID
}
