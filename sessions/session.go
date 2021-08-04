package sessions

type Session struct {
	cookieName string
	ID         string
	manager    *Manager
}
