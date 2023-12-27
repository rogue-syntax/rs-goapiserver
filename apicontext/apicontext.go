package apicontext

import (
	"context"
	"errors"

	"rs-apiserver.com/apireturn/apierrorkeys"
	"rs-apiserver.com/entities/user"
)

type devGenMsgType string

const devGenMsgKey devGenMsgType = "devGenMsg"

func CtxWithdevGenMsg(ctx context.Context, msg *string) context.Context {
	return context.WithValue(ctx, devGenMsgKey, msg)
}
func CtxGetdevGenMsg(ctx context.Context) (*string, error) {
	user, ok := ctx.Value(devGenMsgKey).(*string)
	if !ok {
		err := errors.New("failed to get devGenMsg from devGenMsgKey")
		return nil, err
	}
	return user, nil
}

type userKeyType string

const userKey userKeyType = "user"

func CtxWithUser(ctx context.Context, user *user.UserExternal) context.Context {
	return context.WithValue(ctx, userKey, user)
}
func CtxGetUser(ctx context.Context) (*user.UserExternal, error) {
	user, ok := ctx.Value(userKey).(*user.UserExternal)
	if !ok {
		err := errors.New(apierrorkeys.ContextError)
		return nil, err
	}
	return user, nil
}

//MAKE THIS FOR ISSUER
/*
type issuerKeyType string

const issuerKey issuerKeyType = "issuer"

func CtxWithIssuer(ctx context.Context, issuer *implIssuerModels.UserIssuer) context.Context {
	return context.WithValue(ctx, issuerKey, issuer)
}
func CtxGetIssuer(ctx context.Context) (*implIssuerModels.UserIssuer, error) {
	issuer, ok := ctx.Value(issuerKey).(*implIssuerModels.UserIssuer)
	if !ok {
		err := errors.New(apiErrorKeys.ContextError)
		return nil, err
	}
	return issuer, nil
}
*/
