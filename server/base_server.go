package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/atlasgurus/rulestone/actors"
	"github.com/atlasgurus/rulestone/condition"
	"github.com/atlasgurus/rulestone/engine"
	"github.com/atlasgurus/rulestone/types"
	"os"
	"path/filepath"
)

type RuleEngineInfo struct {
	RuleEngineId int16
	repo         *engine.RuleEngineRepo
	api          *engine.RuleApi
	ctx          *types.AppContext
	ruleEngine   *engine.RuleEngine
}

type BaseServer interface {
	ActivateRuleEngine(id int) int
	NewRulestoneEngine() *RuleEngineInfo
	CreateNewRuleEngine() int16
	AddRuleFromFile(id int, rulePath string) int
	AddRuleFromString(id int, rule string, format string) int
	AddRulesFromDirectory(id int, rulesPath string) int
	PerformMatch(requestId int64, ruleEngineID int, jsonData string, commObj interface{}) error
	responseMatches(requestId int64, matches []condition.RuleIdType, commObj interface{})
}

type BaseRulestoneServer struct {
	base        BaseServer
	ruleEngines []*RuleEngineInfo
	matchActor  *actors.Actor
}

func (rs *BaseRulestoneServer) ActivateRuleEngine(id int) int {
	fmt.Println("Activating rule engine")

	newEngine, err := engine.NewRuleEngine(rs.ruleEngines[id].repo)
	if err != nil {
		return -1
	}
	newId := len(rs.ruleEngines)
	rs.ruleEngines[id].ruleEngine = newEngine

	return newId
}

func (rs *BaseRulestoneServer) NewRulestoneEngine() *RuleEngineInfo {
	result := RuleEngineInfo{RuleEngineId: int16(len(rs.ruleEngines)), ctx: types.NewAppContext()}
	result.api = engine.NewRuleApi(result.ctx)
	result.repo = engine.NewRuleEngineRepo()
	rs.ruleEngines = append(rs.ruleEngines, &result)
	return &result
}

func (rs *BaseRulestoneServer) CreateNewRuleEngine() int16 {
	eng := rs.NewRulestoneEngine()
	return eng.RuleEngineId
}

func (rs *BaseRulestoneServer) AddRuleFromFile(id int, rulePath string) int {
	ruleId, err := rs.ruleEngines[id].repo.RegisterRuleFromFile(rulePath)
	if err != nil {
		return -1
	}

	return int(ruleId)
}

func (rs *BaseRulestoneServer) AddRuleFromString(id int, rule string, format string) int {
	ruleId, err := rs.ruleEngines[id].repo.RegisterRuleFromString(rule, format)
	if err != nil {
		return -1
	}
	return int(ruleId)
}

func (rs *BaseRulestoneServer) AddRulesFromDirectory(id int, rulesPath string) int {
	// Placeholder. This should initialize your rule engine and return an ID.
	fmt.Println("Initializing rule engine with rules from:", rulesPath)

	files, err := os.ReadDir(rulesPath)
	if err != nil {
		return -1
	}

	for _, file := range files {
		rulePath := filepath.Join(rulesPath, file.Name())
		_, err := rs.ruleEngines[id].repo.RegisterRuleFromFile(rulePath)
		if err != nil {
			return -1
		}
	}

	return len(rs.ruleEngines[id].repo.Rules)
}

// PerformMatch - Use the rule engine to match against the provided JSON
func (rs *BaseRulestoneServer) PerformMatch(requestId int64, ruleEngineID int, jsonData string, commObj interface{}) error {
	var decoded interface{}
	err := json.Unmarshal([]byte(jsonData), &decoded)
	if err != nil {
		return err
	}

	if ruleEngineID >= len(rs.ruleEngines) {
		return errors.New("invalid rule engine ID")
	}

	ruleEngine := rs.ruleEngines[ruleEngineID].ruleEngine

	rs.matchActor.Do(func(actor *actors.Actor) {
		matches := ruleEngine.MatchEvent(decoded)
		rs.base.responseMatches(requestId, matches, commObj)
	})
	return nil
}
