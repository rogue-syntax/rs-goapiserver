package routeroles

import (
	"context"
	"net/http"

	"github.com/rogue-syntax/rs-goapiserver/apicontext"
	"github.com/rogue-syntax/rs-goapiserver/apireturn"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
)

func SetRouteRoles(roleMap map[string][]int) {
	RouteRoles = roleMap
}

var RouteRoles = map[string][]int{
	"/v1/test/testRoleAuthentication": {1},
	"/v1/api":                         {1},
}

func TestRoleAuth(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	_, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apireturn.ApiJSONReturn(err, apierrorkeys.NOError, &w)
		return
	}

	apireturn.ApiJSONReturn("User is admin role", apierrorkeys.NOError, &w)
}
