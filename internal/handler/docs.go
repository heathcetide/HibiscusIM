package handlers

import (
	"HibiscusIM/internal/apidocs"
	"HibiscusIM/internal/models"
	"HibiscusIM/pkg/config"
	"HibiscusIM/pkg/search"
	"net/http"
)

func (h *Handlers) GetDocs() []apidocs.UriDoc {
	// Define the API documentation
	uriDocs := []apidocs.UriDoc{ // test
		{
			Group:   "User Authorization",
			Path:    config.GlobalConfig.APIPrefix + "/auth/login",
			Method:  http.MethodPost,
			Desc:    "User login with email and password",
			Request: apidocs.GetDocDefine(models.LoginForm{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "email", Type: apidocs.TYPE_STRING},
					{Name: "activation", Type: apidocs.TYPE_BOOLEAN, CanNull: true},
				},
			},
		},
		{
			Group:        "User Authorization",
			Path:         config.GlobalConfig.APIPrefix + "/auth/logout",
			Method:       http.MethodPost,
			AuthRequired: true,
			Desc:         "User logout, if `?next={NEXT_URL}`is not empty, redirect to {NEXT_URL}",
		},
		{
			Group:        "User Authorization",
			Path:         config.GlobalConfig.APIPrefix + "/auth/register",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "User register with email and password",
			Request:      apidocs.GetDocDefine(models.RegisterUserForm{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "email", Type: apidocs.TYPE_STRING, Desc: "The email address"},
					{Name: "activation", Type: apidocs.TYPE_BOOLEAN, Desc: "Is the account activated"},
					{Name: "expired", Type: apidocs.TYPE_STRING, Default: "180d", CanNull: true, Desc: "If email verification is required, it will be verified within the valid time"},
				},
			},
		},
		{
			Group:        "User Authorization",
			Path:         config.GlobalConfig.APIPrefix + "/auth/reset_password",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Send a verification code to the email address, and then click the link in the email to reset the password",
			Request:      apidocs.GetDocDefine(models.ResetPasswordForm{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "expired", Type: apidocs.TYPE_STRING, Default: "30m", Desc: "Must be verified within the valid time"},
				},
			},
		},
		{
			Group:        "User Authorization",
			Path:         config.GlobalConfig.APIPrefix + "/auth/reset_password_done",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Setup new password",
			Request:      apidocs.GetDocDefine(models.ResetPasswordDoneForm{}),
			Response: &apidocs.DocField{
				Type: apidocs.TYPE_BOOLEAN,
				Desc: "true if success",
			},
		},
		{
			Group:        "User Authorization",
			Path:         config.GlobalConfig.APIPrefix + "/auth/change_password",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Setup new password when user is logged in",
			Request:      apidocs.GetDocDefine(models.ChangePasswordForm{}),
			Response: &apidocs.DocField{
				Type: apidocs.TYPE_BOOLEAN,
				Desc: "true if success",
			},
		},
		{
			Group:        "User Authorization",
			Path:         config.GlobalConfig.APIPrefix + "/auth/send/email",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Send email verification code",
			Request:      apidocs.GetDocDefine(models.SendEmailVerifyEmail{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "expired", Type: apidocs.TYPE_STRING, Default: "30m", Desc: "Must be verified within the valid time"},
				},
			},
		},
		{
			Group:        "System Module",
			Path:         "/api/system/health",
			Method:       http.MethodGet,
			Summary:      "数据库健康状态",
			AuthRequired: false,
			Desc:         `检查数据库健康状态`,
		},
	}

	if config.GlobalConfig.SearchEnabled {
		uriDocs = append(uriDocs, []apidocs.UriDoc{
			{
				Group:   "Search",
				Path:    config.GlobalConfig.APIPrefix + "/search",
				Method:  http.MethodPost,
				Desc:    "Execute a search query",
				Request: apidocs.GetDocDefine(search.SearchRequest{}),
				Response: &apidocs.DocField{
					Type: "object",
					Fields: []apidocs.DocField{
						{Name: "Total", Type: apidocs.TYPE_INT},
						{Name: "Took", Type: apidocs.TYPE_INT},
						{Name: "Hits", Type: "array", Fields: []apidocs.DocField{
							{Name: "ID", Type: apidocs.TYPE_STRING},
							{Name: "Score", Type: apidocs.TYPE_FLOAT},
							{Name: "Fields", Type: "object"},
						}},
					},
				},
			},
			{
				Group:        "Search",
				Path:         config.GlobalConfig.APIPrefix + "/search/index",
				Method:       http.MethodPost,
				AuthRequired: true,
				Desc:         "Index a new document",
				Request:      apidocs.GetDocDefine(search.Doc{}),
				Response: &apidocs.DocField{
					Type: apidocs.TYPE_BOOLEAN,
					Desc: "true if document is indexed successfully",
				},
			},
			{
				Group:        "Search",
				Path:         config.GlobalConfig.APIPrefix + "/search/delete",
				Method:       http.MethodPost,
				AuthRequired: true,
				Desc:         "Delete a document by its ID",
				Request: &apidocs.DocField{
					Type: "object",
					Fields: []apidocs.DocField{
						{Name: "ID", Type: apidocs.TYPE_STRING},
					},
				},
				Response: &apidocs.DocField{
					Type: apidocs.TYPE_BOOLEAN,
					Desc: "true if document is deleted successfully",
				},
			},
			{
				Group:        "Search",
				Path:         config.GlobalConfig.APIPrefix + "/search/auto-complete",
				Method:       http.MethodPost,
				AuthRequired: false,
				Desc:         "Get search query auto-completion suggestions",
				Request: &apidocs.DocField{
					Type: "object",
					Fields: []apidocs.DocField{
						{Name: "Keyword", Type: apidocs.TYPE_STRING},
					},
				},
				Response: &apidocs.DocField{
					Type: "object",
					Fields: []apidocs.DocField{
						{Name: "suggestions", Type: "array", Fields: []apidocs.DocField{
							{Name: "suggestion", Type: apidocs.TYPE_STRING},
						}},
					},
				},
			},
			{
				Group:        "Search",
				Path:         config.GlobalConfig.APIPrefix + "/search/suggest",
				Method:       http.MethodPost,
				AuthRequired: false,
				Desc:         "Get search suggestions based on the keyword",
				Request: &apidocs.DocField{
					Type: "object",
					Fields: []apidocs.DocField{
						{Name: "Keyword", Type: apidocs.TYPE_STRING},
					},
				},
				Response: &apidocs.DocField{
					Type: "object",
					Fields: []apidocs.DocField{
						{Name: "suggestions", Type: "array", Fields: []apidocs.DocField{
							{Name: "suggestion", Type: apidocs.TYPE_STRING},
						}},
					},
				},
			},
		}...)
	}
	return uriDocs
}
