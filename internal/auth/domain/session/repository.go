package session

import "bitmerchant/internal/common"

// Repository defines operations for Session persistence.
type Repository interface {
	Save(session *Session) error
	Get(id string) (*Session, error)
	Delete(id string) error
	DeleteByUserID(userID common.UserID) error
}
