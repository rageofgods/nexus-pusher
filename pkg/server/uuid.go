package server

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

func (u *uploadService) searchById(id uuid.UUID) (*Message, error) {
	if msg, ok := u.messages[id]; ok {
		return msg, nil
	}
	return nil, fmt.Errorf("id %v not found", id)
}

func (u *uploadService) deleteById(id uuid.UUID) {
	delete(u.messages, id)
}

func (u *uploadService) genMessageWithId() (*Message, error) {
	// Generate new random id
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	// Save to data
	m := &Message{
		ID: id,
	}
	u.messages[id] = m
	return m, nil
}

// completeById set complete flag to message which returned to client
func (u *uploadService) completeById(id uuid.UUID, textResult ...string) error {
	// Search in map
	msg, err := u.searchById(id)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	// Set complete flag
	msg.Complete = true
	msg.Response = fmt.Sprintf("%s", strings.Join(textResult, "\n"))

	return nil
}
