package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

const (
	RuleTypeSystem = "system"
	RuleTypeCustom = "custom"
)

type Rule struct {
	Id         bson.ObjectId `json:"_id" bson:"_id"`
	Method     string        `json:"method"`
	Path       string
	I18n       string
	Alias      string
	Type       string
	GroupAlias string    `json:"group_alias" bson:"group_alias"`
	GroupI18n  string    `json:"group_i18n" bson:"group_i18n"`
	CreateTs   time.Time `json:"create_ts" bson:"create_ts"`
	UpdateTs   time.Time `json:"update_ts" bson:"update_ts"`
	DeleteTs   time.Time `json:"update_ts" bson:"delete_ts"`
}
