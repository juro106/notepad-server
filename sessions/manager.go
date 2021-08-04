package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
)

// セッションマネージャー構造体
type Manager struct {
	database map[string]interface{}
}

var mg Manager

// 新規マネージャー生成
func NewManager() *Manager {
	return &mg
}

// セッションIDの発行
func (m *Manager) NewSeesionID() string {
	b := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 新規セッションの生成
func (m *Manager) New(r *http.Request, cookieName string) (*Session, error) {
	cookie, err := r.Cookie(cookieName)
	if err == nil && m.Exists(cookie.Value) {
		return nil, errors.New("sessionIDは既に発行されています")
	}

	session := NewSeesion(m, cookieName)
	session.ID = m.NewSessionID()
	session.request = r

	return session, nil
}

// セッション情報の保存
func (m *Manager) Save(r *http.Request, w http.ResponseWriter, session *Session) error {
	m.database[session.ID] = session

	c := &http.Cookie{
		Name:  session.Name(),
		Value: session.ID,
		Path:  "/",
	}

	http.SetCookie(session.writer, c)
	return nil
}

// 既存セッションの存在チェック
func (m *Manager) Exists(sessionID string) bool {
	_, r := m.database[sessionID]
	return r
}

// 既存セッションの取得
func (m *Manager) Get(r *http.Request, cookieName string) (*Session, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		// リクエストからcookie情報を取得できない場合
		return nil, err
	}

	sessionID := cookie.Value
	// cookie情報からセッション情報を取得
	buffer, exists := m.database[sessionID]
	if !exists {
		return nil, errors.New("無効なセッションIDです")
	}

	session := buffer.(*Session)
	session.request = r
	return session, nil
}

// セッションの破棄
func (m *Manager) Destroy(sessionID string) {
	delete(m.database, sessionID)
}
