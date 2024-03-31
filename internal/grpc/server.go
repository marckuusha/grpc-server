package server

import (
	"context"
	"grpc-server/internal/logic"

	protoc "github.com/marckuusha/protoc/gen/go/hwserv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerGrpc struct {
	protoc.UnimplementedHWServServer
	logicApp *logic.LogicApp
}

func Register(gRPC *grpc.Server, logicApp *logic.LogicApp) {
	protoc.RegisterHWServServer(gRPC, &ServerGrpc{logicApp: logicApp})
}

func (s *ServerGrpc) SendMsg(ctx context.Context, req *protoc.SendMsgReq) (*protoc.SendMsgResp, error) {

	if req.UserName == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	templatestr, err := s.logicApp.GenerateText(req.UserName)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed generate")
	}

	return &protoc.SendMsgResp{
		Msg: templatestr,
	}, nil
}
