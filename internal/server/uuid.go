package server

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

func (u *webService) searchById(id uuid.UUID) (*Message, error) {
	if msg, ok := u.messages[id]; ok {
		return msg, nil
	}
	return nil, fmt.Errorf("id %v not found", id)
}

func (u *webService) deleteById(id uuid.UUID) {
	delete(u.messages, id)
}

func (u *webService) genMessageWithId() (*Message, error) {
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
func (u *webService) completeById(id uuid.UUID, textResult ...string) error {
	// Search in map
	msg, err := u.searchById(id)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	// Set complete flag
	msg.Complete = true
	msg.Response = strings.Join(textResult, "\n")

	return nil
}
