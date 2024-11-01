package api

import (
	"errors"
	"net/http"

	"github.com/grafana/grafana/pkg/apimachinery/errutil"
	"github.com/grafana/grafana/pkg/infra/metrics"
	"github.com/grafana/grafana/pkg/middleware/cookies"
	"github.com/grafana/grafana/pkg/services/authn"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/login"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/web"
)

const (
	OauthStateCookieName = "oauth_state"
	OauthPKCECookieName  = "oauth_code_verifier"
)

func (hs *HTTPServer) OAuthLogin(reqCtx *contextmodel.ReqContext) {
	name := web.Params(reqCtx.Req)[":name"]

	if errorParam := reqCtx.Query("error"); errorParam != "" {
		errorDesc := reqCtx.Query("error_description")
		hs.log.Error("failed to login ", "error", errorParam, "errorDesc", errorDesc)
		hs.redirectWithError(reqCtx, errutil.Unauthorized("oauth.login", errutil.WithPublicMessage(hs.Cfg.OAuthLoginErrorMessage)).Errorf("Login provider denied login request"))
		return
	}

	code := reqCtx.Query("code")
	redirectTo := reqCtx.Query("redirectTo")

	req := &authn.Request{HTTPRequest: reqCtx.Req}
	if code == "" {
		redirect, err := hs.authnService.RedirectURL(reqCtx.Req.Context(), authn.ClientWithPrefix(name), req)
		if err != nil {
			reqCtx.Redirect(hs.redirectURLWithErrorCookie(reqCtx, err))
			return
		}

		cookies.WriteCookie(reqCtx.Resp, OauthStateCookieName, redirect.Extra[authn.KeyOAuthState], hs.Cfg.OAuthCookieMaxAge, hs.CookieOptionsFromCfg)

		if hs.Features.IsEnabledGlobally(featuremgmt.FlagUseSessionStorageForRedirection) {
			cookies.WriteCookie(reqCtx.Resp, "redirectTo", redirectTo, hs.Cfg.OAuthCookieMaxAge, hs.CookieOptionsFromCfg)
		}
		if pkce := redirect.Extra[authn.KeyOAuthPKCE]; pkce != "" {
			cookies.WriteCookie(reqCtx.Resp, OauthPKCECookieName, pkce, hs.Cfg.OAuthCookieMaxAge, hs.CookieOptionsFromCfg)
		}

		reqCtx.Redirect(redirect.URL)
		return
	}

	identity, err := hs.authnService.Login(reqCtx.Req.Context(), authn.ClientWithPrefix(name), req)
	// NOTE: always delete these cookies, even if login failed
	cookies.DeleteCookie(reqCtx.Resp, OauthStateCookieName, hs.CookieOptionsFromCfg)
	cookies.DeleteCookie(reqCtx.Resp, OauthPKCECookieName, hs.CookieOptionsFromCfg)

	if err != nil {
		reqCtx.Redirect(hs.redirectURLWithErrorCookie(reqCtx, err))
		return
	}

	metrics.MApiLoginOAuth.Inc()
	authn.HandleLoginRedirect(reqCtx.Req, reqCtx.Resp, hs.Cfg, identity, hs.ValidateRedirectTo, hs.Features)
}

type PretendOAuthLoginDTO struct {
	Create *user.CreateUserCommand
	Info   *struct {
		AuthModule string
		AuthId     string
		UserId     int64
	}
}

func (hs *HTTPServer) PretendOAuthLogin(reqCtx *contextmodel.ReqContext) {
	var dto PretendOAuthLoginDTO
	if err := web.Bind(reqCtx.Req, &dto); err != nil {
		reqCtx.WriteErrOrFallback(http.StatusBadRequest, "request was not valid DTO", err)
		return
	}

	if dto.Create != nil {
		_, err := hs.userService.GetByEmail(reqCtx.Req.Context(), &user.GetUserByEmailQuery{Email: dto.Create.Email})
		if err != nil && !errors.Is(err, user.ErrUserNotFound) {
			reqCtx.WriteErrOrFallback(http.StatusInternalServerError, "request failed on fetching user", err)
			return
		} else if err != nil {
			_, err = hs.userService.Create(reqCtx.Req.Context(), dto.Create)
			if err != nil {
				reqCtx.WriteErrOrFallback(http.StatusInternalServerError, "user could not be created", err)
				return
			}
		}
	}

	if dto.Info != nil {
		ai, err := hs.authInfoService.GetAuthInfo(reqCtx.Req.Context(), &login.GetAuthInfoQuery{
			UserId:     dto.Info.UserId,
			AuthId:     dto.Info.AuthId,
			AuthModule: dto.Info.AuthModule,
		})
		if err != nil {
			reqCtx.WriteErrOrFallback(http.StatusInternalServerError, "auth info failed fetching", err)
			return
		}

		if ai != nil {
			err = hs.authInfoService.UpdateAuthInfo(reqCtx.Req.Context(), &login.UpdateAuthInfoCommand{
				AuthModule: dto.Info.AuthModule,
				AuthId:     dto.Info.AuthId,
				UserId:     dto.Info.UserId,
			})
		} else {
			err = hs.authInfoService.SetAuthInfo(reqCtx.Req.Context(), &login.SetAuthInfoCommand{
				AuthModule: dto.Info.AuthModule,
				AuthId:     dto.Info.AuthId,
				UserId:     dto.Info.UserId,
			})
		}
		if err != nil {
			reqCtx.WriteErrOrFallback(http.StatusInternalServerError, "auth info failed upsert", err)
			return
		}
	}

	reqCtx.Resp.WriteHeader(http.StatusOK)
}
