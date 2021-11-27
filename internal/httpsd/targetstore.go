package httpsd

import (
	"fmt"

	"github.com/asdine/storm/v3"
)

// GET    /api/v1/target/                                        # return targets list
// POST   /api/v1/target/                                        # creates a new target group
// GET    /api/v1/target/<target_group_id>                       # retrieves the target group
// PUT    /api/v1/target/<target_group_id>                       # add a label or server to target group
// PATCH  /api/v1/target/<target_group_id>/label/<label_key>     # updates a label in a target group
// DELETE /api/v1/target/<target_group_id>/label/<label_key>     # deletes a label in a target group
// DELETE /api/v1/target/<target_group_id>/server/<server_addr>  # deletes a server in a target group

type Target struct {
	ID   int `storm:"id,increment" json:"-"`
	Addr string
}

type TargetGroup struct {
	ID      int                    `storm:"id,increment" json:"-"`
	Targets []string               `json:"targets"`
	Labels  map[string]interface{} `json:"labels"`
}

type TargetStore struct {
	db *storm.DB
}

//New create a new HTTP service discovery
func New(db *storm.DB) *TargetStore {
	db.Init(&TargetGroup{})
	ts := &TargetStore{db: db}
	return ts
}

//GetAllTargets returns a list of all target groups
func (ts *TargetStore) GetAllTargetGroups() []TargetGroup {
	fmt.Printf("getting all target groups\n")
	var tgs []TargetGroup
	err := ts.db.All(&tgs)
	if err != nil {
		fmt.Printf("error in querying %s", err.Error())
		return []TargetGroup{}
	}
	fmt.Printf("returning %v\n", tgs)
	return tgs
}

//CreateTargetGroup creates a new target group, returns error if
//target group couldn't be created
func (ts *TargetStore) CreateTargetGroup(tg TargetGroup) (int, error) {
	fmt.Printf("creating target group %+v \n", tg)

	err := ts.db.Save(&tg)
	if err != nil {
		fmt.Printf("error when saving %s \n", err.Error())
		return -1, err
	}
	return tg.ID, nil
}

//GetTargetGroup returns a target group with ID, returns error if
//target group doesn't exist
func (ts *TargetStore) GetTargetGroup(id int) (TargetGroup, error) {
	var tg TargetGroup
	err := ts.db.One("ID", id, &tg)
	if err != nil {
		fmt.Printf("error in querying %s", err.Error())
		return TargetGroup{}, err
	}
	return tg, nil
}

// UpdateTarget updates port of a target, returns error if
// target group doesn't exist
func (ts *TargetStore) UpdateTargetGroup(tg *TargetGroup) (TargetGroup, error) {
	err := ts.db.Update(tg)
	if err != nil {
		return TargetGroup{}, err
	}
	return *tg, nil
}

//DeleteTarget deletes a target from targets of a target group, returns error if
//target group doesn't exist
func (ts *TargetStore) DeleteTargetGroup(tg *TargetGroup) error {
	err := ts.db.DeleteStruct(tg)
	if err != nil {
		return err
	}
	return nil
}
