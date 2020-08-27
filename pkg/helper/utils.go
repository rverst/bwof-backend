package helper

import (
	"encoding/binary"
	"github.com/google/uuid"
	"math"
	"time"
)

func Itob(id uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)
	return b
}

func Btoi(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}

func UUIDtoBytes(id uuid.UUID) []byte {
	if b, err := id.MarshalBinary(); err == nil {
		return b
	}
	return make([]byte, 0)
}

func BytesToUUID(b []byte) uuid.UUID {
	if id, err := uuid.ParseBytes(b); err == nil {
		return id
	}
	return uuid.Nil
}

func HotRotation(t time.Time) int {
	d := time.Now().Sub(t).Hours() / 24.0
	f := (-0.2)*math.Pow(d, 2.0) + 4.0

	return int(math.Max(1, math.Floor(f)))
}
