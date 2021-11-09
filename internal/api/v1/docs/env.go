package docs

import "github.com/epinio/epinio/pkg/api/core/v1/models"

// Env -- Application Environment

// swagger:route GET /namespaces/{namespace}/applications/{app}/environment app-env EnvList
// responses:
//   200: EnvListResponse

// swagger:parameters EnvList
type EnvListParams struct {
	Namespace string
	App       string
}

// swagger:response EnvListResponse
type EnvListResponse struct {
	// in: body
	Body models.EnvVariableMap
}

// swagger:route GET /namespaces/{namespace}/applications/{app}/environment/{env}/match/{pattern} app-env EnvMatch
// responses:
//   200: EnvMatchResponse

// swagger:parameters EnvMatch
type EnvMatchParams struct {
	Namespace string
	App       string
	Pattern   string
}

// swagger:response EnvMatchResponse
type EnvMatchResponse struct {
	// in: body
	Body models.EnvMatchResponse
}

// swagger:route GET /namespaces/{namespace}/applications/{app}/environment/{env}/match app-env EnvMatch0
// responses:
//   200: EnvMatchResponse

// swagger:parameters EnvMatch0
type EnvMatch0Params struct {
	Namespace string
	App       string
}

// See EnvMatch above

// swagger:route POST /namespaces/{namespace}/applications/{app}/environment app-env EnvSet
// responses:
//   200: EnvSetResponse

// swagger:parameters EnvSet
type EnvSetParams struct {
	Namespace string
	App       string
	// in: body
	Body models.EnvVariableMap
}

// swagger:response EnvSetResponse
type EnvSetResponse struct {
	// in: body
	Body models.Response
}

// swagger:route GET /namespaces/{namespace}/applications/{app}/environment/{env} app-env EnvShow
// responses:
//   200: EnvShowResponse

// swagger:parameters EnvShow
type EnvShowParams struct {
	Namespace string
	App       string
}

// swagger:response EnvShowResponse
type EnvShowResponse struct {
	// in: body
	Body models.EnvVariable
}

// swagger:route DELETE /namespaces/{namespace}/applications/{app}/environment/{env} app-env EnvUnset
// responses:
//   200: EnvUnsetResponse

// swagger:parameters EnvUnset
type EnvUnsetParams struct {
	Namespace string
	App       string
}

// swagger:response EnvUnsetResponse
type EnvUnsetResponse struct {
	// in: body
	Body models.Response
}
