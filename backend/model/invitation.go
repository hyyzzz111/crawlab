package model

import (
	"crawlab/database"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"runtime/debug"
	"time"
)

type Invitation struct {
	Id            bson.ObjectId `json:"_id" bson:"_id"`
	Token         string        `json:"token"`
	Used          bool          `json:"used"`
	Account       string        `json:"account"`
	EncryptRT     string        `json:"encrypt_result" bson:"encrypt_result"`
	Status        bool          `json:"status" bson:"status"`
	Roles         []string      `json:"roles" bson:"roles"`
	ACLs          []string      `json:"acls"`
	RegisterLimit int           `json:"limit" bson:"limit"`
	Salt          string        `json:"salt" bson:"salt"`
	ExpireTs      time.Time     `json:"expire_ts" bson:"expire_ts"`
	UpdateTs      time.Time     `json:"update_ts" bson:"update_ts"`
	CreateTs      time.Time     `json:"create_ts" bson:"create_ts"`
	UpdateTsUnix  int64         `json:"update_ts_unix" bson:"update_ts_unix"`
}

func (n *Invitation) Add() error {
	s, c := database.GetCol("invitations")
	defer s.Close()
	n.Id = bson.NewObjectId()
	n.UpdateTs = time.Now()
	n.UpdateTsUnix = time.Now().Unix()
	n.CreateTs = time.Now()
	if err := c.Insert(&n); err != nil {
		debug.PrintStack()
		return err
	}
	return nil
}

func (n *Invitation) Save() error {
	s, c := database.GetCol("invitations")
	defer s.Close()
	n.UpdateTs = time.Now()
	if err := c.UpdateId(n.Id, n); err != nil {
		return err
	}
	return nil
}
func UpdateInvitation(id bson.ObjectId, values bson.M) error {
	s, c := database.GetCol("invitations")
	defer s.Close()

	fmt.Println(bson.IsObjectIdHex(id.Hex()))
	if err := c.UpdateId(id, bson.M{"$set": values}); err != nil {
		return err
	}
	return nil
}
func GetInvitation(id bson.ObjectId) (inv *Invitation, err error) {
	s, c := database.GetCol("invitations")
	defer s.Close()
	err = c.FindId(id).One(&inv)
	return inv, err
}
func GetInvitationList(filter interface{}) ([]*Invitation, error) {
	s, c := database.GetCol("invitations")
	defer s.Close()

	var results []*Invitation
	if err := c.Find(filter).Sort("-_id").All(&results); err != nil {
		debug.PrintStack()
		return results, err
	}
	return results, nil
}
