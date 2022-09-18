package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/utils"
)

type MW struct {
	users interfaces.Users
}

//	New is function to Create a user.
func New(users interfaces.Users) *MW {
	return &MW{
		users: users,
	}
}

func (M *MW) UserIDInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var userID string
	var token string
	var err error
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(interfaces.UserIDCtxName.String())
		if len(values) > 0 {
			userID, err = M.users.ReadSessionID(values[0])
			if err == nil {
				return handler(context.WithValue(ctx, interfaces.UserIDCtxName.String(), userID), req)
			}
		}
	}
	uid := utils.CreateID(16)
	token, err = M.users.CreateSissionID(uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `create token error: `+err.Error())
	}
	md := metadata.New(map[string]string{interfaces.UserIDCtxName.String(): token})
	err = grpc.SetTrailer(ctx, md)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `set trailer err: `+err.Error())
	}
	return handler(context.WithValue(ctx, interfaces.UserIDCtxName.String(), userID), req)
}
