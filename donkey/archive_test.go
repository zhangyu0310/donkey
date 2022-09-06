package donkey

import (
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"
)

func initArchive(t *testing.T, id int) *Archive {
	archive, err := NewArchive(id)
	if err != nil {
		t.Error("Init archive", id, "failed, err:", err)
	}
	return archive
}

func TestArchive_AppendEntry(t *testing.T) {
	archive := initArchive(t, 1)
	for i := 0; i < 100; i++ {
		err := archive.AppendEntry(&Entry{
			Id:   rand.Uint64(),
			Uuid: uuid.New().String(),
		})
		if err != nil {
			t.Error("Append entry failed, err:", err)
		}
	}
	_ = os.Remove("donkey_archive_1")
}

func TestArchive_GetOneEntry(t *testing.T) {
	archive := initArchive(t, 2)
	entries := make([]*Entry, 0, 100)
	for i := 0; i < 100; i++ {
		entry := &Entry{
			Id:   rand.Uint64(),
			Uuid: uuid.New().String(),
		}
		entries = append(entries, entry)
		err := archive.AppendEntry(entry)
		if err != nil {
			t.Error("Append entry failed, err:", err)
		}
	}

	for i := 0; i < 100; i++ {
		e, err := archive.GetOneEntry()
		if err != nil {
			t.Error("Get one entry failed, err:", err)
		}
		entry := entries[i]
		if e.Id != entry.Id {
			t.Error("Entry id is different")
		}
		if e.Uuid != entry.Uuid {
			t.Error("Entry uuid is different")
		}
	}
	_ = os.Remove("donkey_archive_2")
}