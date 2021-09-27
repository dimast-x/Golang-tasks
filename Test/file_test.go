package person

import (
	"testing"
)

func TestNewPersonPositiveAge(t *testing.T) {
	_, err := NewPerson(1)
	if err != nil {
		t.Errorf("Expected person, received %v", err)
	}
}

func TestNewPersonNegativeAge(t *testing.T) {
	p, err := NewPerson(-1)
	if err == nil {
		t.Errorf("Expected error, received %v", p)
	}
}