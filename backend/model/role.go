package model

import (
	"crawlab/database"
	"github.com/globalsign/mgo/bson"
	"time"
)

type Role struct {
	Id       bson.ObjectId `json:"_id" bson:"_id"`
	Name     string
	Alias    string
	Enabled  bool
	CreateTs time.Time `json:"create_ts" bson:"create_ts"`
	UpdateTs time.Time `json:"update_ts" bson:"update_ts"`
	DeleteTs time.Time `json:"update_ts" bson:"delete_ts"`
}

func (r Role) GetRolesByIds(alias []string) (roles []*Role, err error) {
	s, c := database.GetCol("roles")
	defer s.Close()
	err = c.Find(&bson.M{"alias": &bson.M{"$in": alias}}).All(&roles)
	return
}
