package routeroles

import (
	"context"
	"net/http"

	"rs-apiserver.com/apicontext"
	"rs-apiserver.com/apireturn"
	"rs-apiserver.com/apireturn/apierrorkeys"
)

type intSlice = []int

var RouteRoles = map[string][]int{
	"/v1/test/testRoleAuthentication": {1},
}

func TestRoleAuth(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	_, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apireturn.ApiJSONReturn(err, apierrorkeys.NOError, &w)
		return
	}

	apireturn.ApiJSONReturn("User is admin role", apierrorkeys.NOError, &w)
}
