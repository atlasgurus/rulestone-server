package server

import (
	"context"
	"github.com/atlasgurus/rulestone-server/grpc"
	"github.com/atlasgurus/rulestone/actors"
	"github.com/atlasgurus/rulestone/condition"
)

type RulestoneGrpcServer struct {
	BaseRulestoneServer
	grpc.UnimplementedRulestoneServiceServer
}

func NewGrpcRulestoneServer() *RulestoneGrpcServer {
	result := RulestoneGrpcServer{
		BaseRulestoneServer: BaseRulestoneServer{
			base:        &RulestoneGrpcServer{},
			ruleEngines: make([]*RuleEngineInfo, 0),
			matchActor:  actors.NewActor(nil, 10000),
		},
	}
	return &result
}

func (rs *RulestoneGrpcServer) Match(server grpc.RulestoneService_MatchServer) error {
	for {
		req, err := server.Recv()
		if err != nil {
			return err
		}
		err = rs.PerformMatch(req.RequestId, int(req.RuleEngineId), req.JsonData, server)
		if err != nil {
			return err
		}
	}
}

func (rs *RulestoneGrpcServer) CreateRuleEngine(context.Context, *grpc.EmptyRequest) (*grpc.RuleEngineResponse, error) {
	ruleEngineId := rs.CreateNewRuleEngine()
	return &grpc.RuleEngineResponse{RuleEngineId: int32(ruleEngineId)}, nil
}

func (rs *RulestoneGrpcServer) AddRuleFromJsonString(ctx context.Context, req *grpc.RuleFromStringRequest) (*grpc.RuleResponse, error) {
	ruleId := rs.AddRuleFromString(int(req.RuleEngineId), req.RuleString, "json")
	return &grpc.RuleResponse{RuleId: int32(ruleId)}, nil
}

func (rs *RulestoneGrpcServer) AddRuleFromYamlString(ctx context.Context, req *grpc.RuleFromStringRequest) (*grpc.RuleResponse, error) {
	ruleId := rs.AddRuleFromString(int(req.RuleEngineId), req.RuleString, "yaml")
	return &grpc.RuleResponse{RuleId: int32(ruleId)}, nil
}
func (rs *RulestoneGrpcServer) AddRulesFromFilePath(ctx context.Context, req *grpc.RuleFromFileRequest) (*grpc.NumRulesResponse, error) {
	numRules := rs.AddRulesFromFile(int(req.RuleEngineId), req.RuleFilePath)
	return &grpc.NumRulesResponse{NumRules: int32(numRules)}, nil
}
func (rs *RulestoneGrpcServer) AddRulesFromDirectoryPath(ctx context.Context, req *grpc.RuleFromDirectoryRequest) (*grpc.NumRulesResponse, error) {
	numRules := rs.AddRulesFromDirectory(int(req.RuleEngineId), req.RulesDirectoryPath)
	return &grpc.NumRulesResponse{NumRules: int32(numRules)}, nil
}

func (rs *RulestoneGrpcServer) Activate(ctx context.Context, req *grpc.RuleEngineRequest) (*grpc.EmptyResponse, error) {
	rs.ActivateRuleEngine(int(req.RuleEngineId))
	return &grpc.EmptyResponse{}, nil
}

func (rs *RulestoneGrpcServer) responseMatches(requestId int64, matches []condition.RuleIdType, commObj interface{}) {
	server := commObj.(grpc.RulestoneService_MatchServer)
	int32Matches := make([]int32, len(matches))
	for i, match := range matches {
		int32Matches[i] = int32(match)
	}
	server.Send(&grpc.MatchResponse{RequestId: requestId, MatchIds: int32Matches})
}
