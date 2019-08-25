package utils

import (
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/globalsign/mgo"
	"github.com/pkg/errors"
)

// CasbinRule represents a rule in Casbin.
type CasbinRule struct {
	PType string
	V0    string
	V1    string
	V2    string
	V3    string
	V4    string
	V5    string
}

// adapter represents the MongoDB adapter for policy storage.
type adapter struct {
	dialInfo   *mgo.DialInfo
	session    *mgo.Session
	collection *mgo.Collection
	filtered   bool
}

func NewAdapterWithMGOSession(session *mgo.Session, collection *mgo.Collection) persist.Adapter {
	a := &adapter{}
	a.session = session
	a.collection = collection

	indexes := []string{"ptype", "v0", "v1", "v2", "v3", "v4", "v5"}
	for _, k := range indexes {
		if err := a.collection.EnsureIndexKey(k); err != nil {
			panic(err)
		}
	}
	return a
}

// NewFilteredAdapter is the constructor for FilteredAdapter.
// Casbin will not automatically call LoadPolicy() for a filtered adapter.
func NewFilteredAdapter(session *mgo.Session, collection *mgo.Collection) persist.FilteredAdapter {
	a := NewAdapterWithMGOSession(session, collection).(*adapter)
	a.filtered = true

	return a
}

func (a *adapter) close() {
	a.session.Close()
}

func (a *adapter) dropTable() error {
	session := a.session.Copy()
	defer session.Close()

	err := a.collection.With(session).DropCollection()
	if err != nil {
		if err.Error() != "ns not found" {
			return err
		}
	}
	return nil
}

func loadPolicyLine(line CasbinRule, model model.Model) {
	key := line.PType
	sec := key[:1]

	var tokens []string
	if line.V0 != "" {
		tokens = append(tokens, line.V0)
	} else {
		goto LineEnd
	}

	if line.V1 != "" {
		tokens = append(tokens, line.V1)
	} else {
		goto LineEnd
	}

	if line.V2 != "" {
		tokens = append(tokens, line.V2)
	} else {
		goto LineEnd
	}

	if line.V3 != "" {
		tokens = append(tokens, line.V3)
	} else {
		goto LineEnd
	}

	if line.V4 != "" {
		tokens = append(tokens, line.V4)
	} else {
		goto LineEnd
	}

	if line.V5 != "" {
		tokens = append(tokens, line.V5)
	} else {
		goto LineEnd
	}

LineEnd:
	model[sec][key].Policy = append(model[sec][key].Policy, tokens)
}

// LoadPolicy loads policy from database.
func (a *adapter) LoadPolicy(model model.Model) error {
	return a.LoadFilteredPolicy(model, nil)
}

// LoadFilteredPolicy loads matching policy lines from database. If not nil,
// the filter must be a valid MongoDB selector.
func (a *adapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {
	if filter == nil {
		a.filtered = false
	} else {
		a.filtered = true
	}
	line := CasbinRule{}

	session := a.session.Copy()
	defer session.Close()

	iter := a.collection.With(session).Find(filter).Iter()
	for iter.Next(&line) {
		loadPolicyLine(line, model)
	}

	return iter.Close()
}

// IsFiltered returns true if the loaded policy has been filtered.
func (a *adapter) IsFiltered() bool {
	return a.filtered
}

func savePolicyLine(ptype string, rule []string) CasbinRule {
	line := CasbinRule{
		PType: ptype,
	}

	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}

	return line
}

// SavePolicy saves policy to database.
func (a *adapter) SavePolicy(model model.Model) error {
	if a.filtered {
		return errors.New("cannot save a filtered policy")
	}
	if err := a.dropTable(); err != nil {
		return err
	}

	var lines []interface{}

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, &line)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, &line)
		}
	}

	session := a.session.Copy()
	defer session.Close()

	return a.collection.With(session).Insert(lines...)
}

// AddPolicy adds a policy rule to the storage.
func (a *adapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)

	session := a.session.Copy()
	defer session.Close()

	return a.collection.With(session).Insert(line)
}

// RemovePolicy removes a policy rule from the storage.
func (a *adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)

	session := a.session.Copy()
	defer session.Close()

	if err := a.collection.With(session).Remove(line); err != nil {
		switch err {
		case mgo.ErrNotFound:
			return nil
		default:
			return err
		}
	}
	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	selector := make(map[string]interface{})
	selector["ptype"] = ptype

	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		if fieldValues[0-fieldIndex] != "" {
			selector["v0"] = fieldValues[0-fieldIndex]
		}
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		if fieldValues[1-fieldIndex] != "" {
			selector["v1"] = fieldValues[1-fieldIndex]
		}
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		if fieldValues[2-fieldIndex] != "" {
			selector["v2"] = fieldValues[2-fieldIndex]
		}
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		if fieldValues[3-fieldIndex] != "" {
			selector["v3"] = fieldValues[3-fieldIndex]
		}
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		if fieldValues[4-fieldIndex] != "" {
			selector["v4"] = fieldValues[4-fieldIndex]
		}
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		if fieldValues[5-fieldIndex] != "" {
			selector["v5"] = fieldValues[5-fieldIndex]
		}
	}

	session := a.session.Copy()
	defer session.Close()

	_, err := a.collection.With(session).RemoveAll(selector)
	return err
}
