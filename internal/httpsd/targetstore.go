package httpsd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

// GET    /api/v1/target/                                        # return targets list
// POST   /api/v1/target/                                        # creates a new target group
// GET    /api/v1/target/<target_group_id>                       # retrieves the target group
// PUT    /api/v1/target/<target_group_id>                       # add a label or server to target group
// PATCH  /api/v1/target/<target_group_id>/label/<label_key>     # updates a label in a target group
// DELETE /api/v1/target/<target_group_id>/label/<label_key>     # deletes a label in a target group
// DELETE /api/v1/target/<target_group_id>/server/<server_addr>  # deletes a server in a target group

type Target struct {
	ID   uint64 `json:"id"`
	Addr string `json:"addr"`
}

type TargetGroup struct {
	ID      uint64                 `json:"id"`
	Targets []Target               `json:"targets"`
	Labels  map[string]interface{} `json:"labels"`
}

type TargetStore struct {
	db         *bolt.DB
	rootBucket string
}

//New create a new HTTP service discovery
func New(db *bolt.DB) *TargetStore {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("root"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return &TargetStore{db: db, rootBucket: "root"}
}

//GetAllTargets returns a list of all target groups
func (ts *TargetStore) GetAllTargetGroups() ([]TargetGroup, error) {
	var tgs []TargetGroup
	tx, err := ts.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b := tx.Bucket([]byte(ts.rootBucket))
	tgBkt := b.Bucket([]byte("TargetGroup")) //target group bucket

	tgBkt.ForEach(func(k, v []byte) error {
		tgid, err := strconv.ParseUint(string(k), 10, 64)
		if err != nil {
			return err
		}
		tgObj := TargetGroup{ID: tgid}
		tgiBkt := tgBkt.Bucket(k) //target group id bucket
		tgiBkt.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, []byte("label")) {
				var labels map[string]interface{}
				json.Unmarshal(v, &labels)
				tgObj.Labels = labels
			} else if v == nil {
				bkt := tgiBkt.Bucket(k) // targets bucket
				targets := []Target{}
				if bkt != nil {
					bkt.ForEach(func(k, v []byte) error {
						tid, err := strconv.ParseUint(string(k), 10, 64)
						if err != nil {
							return err
						}
						target := Target{ID: tid, Addr: string(v)}
						targets = append(targets, target)
						return nil
					})
					tgObj.Targets = targets
				}
			}
			return nil
		})
		tgs = append(tgs, tgObj)
		return nil
	})
	return tgs, nil
}

//CreateTargetGroup creates a new target group, returns error if
//target group couldn't be created
func (ts *TargetStore) CreateTargetGroup(tg *TargetGroup) error {
	// Start the transaction.
	tx, err := ts.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Retrieve the root bucket.
	// Assume this has already been created when the store was set up.
	root := tx.Bucket([]byte(ts.rootBucket))

	// Setup the TargetGroup bucket.
	bkt, err := root.CreateBucketIfNotExists([]byte("TargetGroup"))
	if err != nil {
		return err
	}
	tgID, err := bkt.NextSequence()
	if err != nil {
		return err
	}

	targetGroupBkt, err := bkt.CreateBucket([]byte(strconv.FormatUint(tgID, 10)))
	if err != nil {
		return err
	}
	// Marshal and save the encoded user.
	if buf, err := json.Marshal(tg.Labels); err != nil {
		return err
	} else if err := targetGroupBkt.Put([]byte("label"), buf); err != nil {
		return err
	}

	targetBkt, err := targetGroupBkt.CreateBucket([]byte("target"))
	if err != nil {
		return err
	}
	for i, tgt := range tg.Targets {
		targetBkt.Put([]byte(strconv.FormatUint(uint64(i), 10)), []byte(tgt.Addr))
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

//GetTargetGroup returns a target group with ID, returns error if
//target group doesn't exist
func (ts *TargetStore) GetTargetGroup(id int) (*TargetGroup, error) {
	return nil, nil
}

// UpdateTarget updates port of a target, returns error if
// target group doesn't exist
func (ts *TargetStore) UpdateTargetGroup(tg *TargetGroup) (*TargetGroup, error) {
	return nil, nil
}

//DeleteTarget deletes a target from targets of a target group, returns error if
//target group doesn't exist
func (ts *TargetStore) DeleteTargetGroup(tg *TargetGroup) error {
	return nil
}
