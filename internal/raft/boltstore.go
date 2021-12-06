package raft

import (
	"encoding/json"
	"strconv"
	"sync"

	bolt "go.etcd.io/bbolt"
	pb "go.etcd.io/etcd/raft/v3/raftpb"
)

// BoltStorage implements the Storage interface backed by an
// boltdb
type BoltStorage struct {
	sync.Mutex
	db *bolt.DB
}

func NewBoltStorage(db *bolt.DB) *BoltStorage {
	return &BoltStorage{db: db}
}

// InitialState returns the saved HardState and ConfState information.
func (bs *BoltStorage) InitialState() (pb.HardState, pb.ConfState, error) {
	return pb.HardState{}, pb.ConfState{}, nil
}

func limitSize(ents []pb.Entry, maxSize uint64) []pb.Entry {
	if len(ents) == 0 {
		return ents
	}
	size := ents[0].Size()
	var limit int
	for limit = 1; limit < len(ents); limit++ {
		size += ents[limit].Size()
		if uint64(size) > maxSize {
			break
		}
	}
	return ents[:limit]
}

// Entries returns a slice of log entries in the range [lo,hi).
// MaxSize limits the total size of the log entries returned, but
// Entries returns at least one entry if any.
func (bs *BoltStorage) Entries(lo, hi, maxSize uint64) ([]pb.Entry, error) {
	tx, err := bs.db.Begin(true)
	if err != nil {
		return []pb.Entry{}, err
	}
	defer tx.Rollback()
	entsBkt := tx.Bucket([]byte("entries"))

	ents := make([]pb.Entry, 0)
	entsBkt.ForEach(func(k, v []byte) error {
		i, _ := strconv.ParseUint(string(k), 10, 64)
		if i >= lo && i < hi {
			var en map[string]interface{}
			json.Unmarshal(v, &en)
			entry := pb.Entry{
				Term:  en["Term"].(uint64),
				Index: en["Index"].(uint64),
				Type:  pb.EntryType(en["Type"].(int)),
				Data:  en["Data"].([]byte),
			}
			ents = append(ents, entry)
		}
		return nil
	})

	return limitSize(ents, maxSize), nil
}

// Term returns the term of entry i, which must be in the range
// [FirstIndex()-1, LastIndex()]. The term of the entry before
// FirstIndex is retained for matching purposes even though the
// rest of that entry may not be available.
func (bs *BoltStorage) Term(i uint64) (uint64, error) {
	tx, err := bs.db.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	entsBkt := tx.Bucket([]byte("entries"))

	ent := entsBkt.Get([]byte(string(i)))
	if ent != nil {
		var e pb.Entry
		json.Unmarshal(ent, &e)
		return e.Term, nil
	}

	return 0, nil
}

// LastIndex returns the index of the last entry in the log.
func (bs *BoltStorage) LastIndex() (uint64, error) {
	return 0, nil
}

// FirstIndex returns the index of the first log entry that is
// possibly available via Entries (older entries have been incorporated
// into the latest Snapshot; if storage only contains the dummy entry the
// first log entry is not available).
func (bs *BoltStorage) FirstIndex() (uint64, error) {
	return 0, nil
}

// Snapshot returns the most recent snapshot.
// If snapshot is temporarily unavailable, it should return ErrSnapshotTemporarilyUnavailable,
// so raft state machine could know that Storage needs some time to prepare
// snapshot and call Snapshot later.
func (bs *BoltStorage) Snapshot() (pb.Snapshot, error) {
	return pb.Snapshot{}, nil
}
