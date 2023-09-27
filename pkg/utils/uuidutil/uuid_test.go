package uuidutil

import "testing"

func TestUUID(t *testing.T) {
	id := UUID()
	t.Log(id)
	t.Log(len(id))
}
