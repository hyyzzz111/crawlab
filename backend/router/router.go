package router

import (
	"crawlab/database"
	"crawlab/model"
	"crawlab/utils"
	"fmt"
	"github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"sync"

	"time"
)

type HTTPMethod string

var defaultRouteManage *RouteManger
var once sync.Once

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	Any    HTTPMethod = "ANY"
)

type Route struct {
	Method       HTTPMethod
	Path         string
	RuleI18n     string
	RuleTemplate string
	Handlers     gin.HandlersChain
	Handler      gin.HandlerFunc
	Router       string
}
type RouteManger struct {
	RouterGroup map[string]*gin.RouterGroup
	enforcer    *casbin.Enforcer
	rules       []model.Rule
}

func (rm *RouteManger) RegisterRouterGroup(groupName string, routerGroup *gin.RouterGroup) {
	rm.RouterGroup[groupName] = routerGroup
}
func (rm *RouteManger) AppendRule(rule model.Rule) {
	if rule.Id != "" {
		fmt.Printf("warn: override mongo _id : %s", rule.Alias)
	}
	if rule.Type != "" && rule.Type != model.RuleTypeSystem {
		panic("only support system type. please change it.")
	}
	if rule.Method == "" || rule.Path == "" || rule.I18n == "" || rule.Alias == "" {
		fmt.Printf("%+v", rule)
		panic("check rule params")

	}
	if rule.Type == "" {
		rule.Type = model.RuleTypeSystem
	}
	if rule.GroupAlias == "" {
		rule.GroupAlias = "other"
	}
	if rule.GroupI18n == "" {
		rule.GroupI18n = rule.GroupAlias
	}
	rm.rules = append(rm.rules, rule)
}
func (rm *RouteManger) RegisterRoute(route Route, groupAlias string, groupI18n string) {
	if rg, ok := rm.RouterGroup[route.Router]; ok {
		if route.Handlers == nil {
			route.Handlers = []gin.HandlerFunc{route.Handler}
		}
		rg.Handle(string(route.Method), route.Path, route.Handlers...)
		rm.AppendRule(model.Rule{
			Method:     string(route.Method),
			Path:       route.Path,
			I18n:       route.RuleI18n,
			Alias:      route.RuleTemplate,
			Type:       model.RuleTypeSystem,
			GroupAlias: groupAlias,
			GroupI18n:  groupI18n,
		})
		return
	}
	panic(fmt.Sprintf("%s not found.", route.Router))
}

func (rm *RouteManger) RegisterRoutes(routes []Route, groupAlias string, groupI18n string) {
	for _, route := range routes {
		rm.RegisterRoute(route, groupAlias, groupI18n)
	}
}
func (rm *RouteManger) AppendRules(rules []model.Rule) {
	for _, rule := range rules {
		rm.AppendRule(rule)
	}
}

func (rm *RouteManger) readySystemRules() (err error) {
	s, c := database.GetCol("rules")
	defer s.Close()
	count, err := c.Find(bson.M{"type": "system"}).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		_, err = c.RemoveAll(bson.M{"type": "system"})

		if err != nil {
			return err
		}
	}

	rules := make([]interface{}, len(rm.rules))
	for _, rule := range rm.rules {
		rule.Id = bson.NewObjectId()
		rule.CreateTs = time.Now()
		rule.UpdateTs = time.Now()
		rules = append(rules, rule)
	}
	bulk := c.Bulk()
	bulk.Unordered()
	bulk.Insert(rules...)
	_, bulkErr := bulk.Run()
	if bulkErr != nil {
		panic(bulkErr)
	}
	return err
}
func (rm *RouteManger) readyCasbin() (err error) {
	//加载模型配置
	m := casbinModel.NewModel()
	m.AddDef("r", "r", "sub, obj, act,alias")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act) ||r.sub==\"admin\"")
	adapter := utils.NewAdapterWithMGOSession(database.GetCol("casbin_rules"))
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return err
	}
	rm.enforcer = enforcer
	return nil
}
func (rm *RouteManger) readySystemRoles() (err error) {
	//1.创建超级管理员角色 --> 超级管理员直接放空即可,模型定制中admin角色直接通过
	//2.创建普通账号权限. 默认只有查看Dashboard面板的功能
	return nil
}
func (rm *RouteManger) SetUp() (err error) {
	if err = rm.readySystemRules(); err != nil {
		return err
	}
	if err = rm.readyCasbin(); err != nil {
		return err
	}
	if err = rm.readySystemRules(); err != nil {
		return err
	}
	return nil
}
func NewRouteManger() *RouteManger {
	return &RouteManger{
		RouterGroup: map[string]*gin.RouterGroup{},
		rules:       make([]model.Rule, 0),
	}
}

func DefaultManager() *RouteManger {
	once.Do(func() {
		defaultRouteManage = NewRouteManger()
	})
	return defaultRouteManage
}
