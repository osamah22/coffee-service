package auth

import (
	"net/http"

	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword/epmodels"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
)

func InitSuperTokens(cfg Config) error {
	apiBasePath := cfg.APIBasePath
	websiteBasePath := cfg.WebsiteBasePath

	return supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: cfg.ConnectionURI,
		},
		AppInfo: supertokens.AppInfo{
			AppName:         cfg.AppName,
			APIDomain:       cfg.APIDomain,
			WebsiteDomain:   cfg.WebsiteDomain,
			APIBasePath:     &apiBasePath,
			WebsiteBasePath: &websiteBasePath,
		},
		RecipeList: []supertokens.Recipe{
			emailpassword.Init(&epmodels.TypeInput{
				Override: &epmodels.OverrideStruct{
					APIs: func(original epmodels.APIInterface) epmodels.APIInterface {
						if original.SignUpPOST != nil {
							signUp := *original.SignUpPOST
							*original.SignUpPOST = func(formFields []epmodels.TypeFormField, tenantId string, options epmodels.APIOptions, userContext supertokens.UserContext) (epmodels.SignUpPOSTResponse, error) {
								res, err := signUp(formFields, tenantId, options, userContext)
								if err == nil && res.OK != nil {
									_ = res.OK.Session.MergeIntoAccessTokenPayload(accessTokenPayload(res.OK.User.Email, cfg.RoleForEmail(res.OK.User.Email)))
								}
								return res, err
							}
						}

						if original.SignInPOST != nil {
							signIn := *original.SignInPOST
							*original.SignInPOST = func(formFields []epmodels.TypeFormField, tenantId string, options epmodels.APIOptions, userContext supertokens.UserContext) (epmodels.SignInPOSTResponse, error) {
								res, err := signIn(formFields, tenantId, options, userContext)
								if err == nil && res.OK != nil {
									_ = res.OK.Session.MergeIntoAccessTokenPayload(accessTokenPayload(res.OK.User.Email, cfg.RoleForEmail(res.OK.User.Email)))
								}
								return res, err
							}
						}

						return original
					},
				},
			}),
			session.Init(nil),
		},
	})
}

func Handler(next http.Handler) http.Handler {
	return supertokens.Middleware(next)
}

func VerifySession(sessionRequired bool, next http.HandlerFunc) http.HandlerFunc {
	return session.VerifySession(&sessmodels.VerifySessionOptions{
		SessionRequired: &sessionRequired,
	}, next)
}

func SessionFromRequest(r *http.Request) sessmodels.SessionContainer {
	return session.GetSessionFromRequestContext(r.Context())
}

func CORSHeaders() []string {
	return supertokens.GetAllCORSHeaders()
}

func accessTokenPayload(email string, role Role) map[string]interface{} {
	return map[string]interface{}{
		"email": email,
		"role":  string(role),
	}
}
