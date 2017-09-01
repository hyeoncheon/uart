//** OAuth2.0 authorization service provider code built on osin library.
//
//	It supports simple oauth2 authorization cycle with userinfo endpoint,
//	and currently does not support jwt and oidc.
//
// osin library: https://github.com/RangelReale/osin

package actions

// TODO CHECK RFC AND IMPLEMENT AGAIN BUT NOT NOW

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/RangelReale/osin"
	"github.com/gobuffalo/buffalo"

	"github.com/hyeoncheon/uart/models"
	"github.com/hyeoncheon/uart/utils"
)

var svr *osin.Server
var logger buffalo.Logger

const (
	privKeyFile = "files/jwt.private.pem"
	pubKeyFile  = "files/jwt.public.pem"
)

func initProvider(l buffalo.Logger) {
	logger = l

	conf := osin.NewServerConfig()
	conf.AccessExpiration = 60
	svr = osin.NewServer(conf, newProvider())
	keyFile := uartHome + "/" + privKeyFile
	if _, err := os.Stat(keyFile); err != nil {
		logger.Error("Abort! cannot found key file: $UART_HOME/", privKeyFile)
		fmt.Println("Abort! cannot found key file: $UART_HOME/" + privKeyFile)
		os.Exit(1)
	}
	svr.AccessTokenGen = utils.NewRS256AccessTokenGen(brandName, keyFile)

	logger.Info("oauth2 provider with jwt support initialized!")
	return
}

func authorizeHandler(c buffalo.Context) error {
	resp := svr.NewResponse()
	defer resp.Close()
	logger = c.Logger().WithField("category", "oauth2")

	logger.Debugf("---- authorize params ---- %v", c.Params())
	if c.Param("client_id") == "" || c.Param("response_type") == "" {
		logger.Error("invalid request!")
		return c.Redirect(http.StatusBadRequest, "/")
	}
	if ar := svr.HandleAuthorizeRequest(resp, c.Request()); ar != nil {
		app := ar.Client.(*AppClient).GetApp()
		user := currentMember(c)
		if !user.IsActive {
			c.Flash().Add("danger", t(c, "no.perm.inactive.member"))
			return c.Redirect(http.StatusTemporaryRedirect, "/membership/me")
		}

		if ar.Authorized = user.Granted(app.ID, ar.Scope); !ar.Authorized {
			logger.Warnf("request for grant: %v, %v, %v", app, user, ar.Scope)
			c.Set("app", app)
			c.Set("scope", ar.Scope)
			c.Set("appkey", ar.Client.GetId())
			c.Session().Set("origin", c.Request().RequestURI)
			return c.Render(200, r.HTML("oauth2/grant.html"))
		}
		logger.Infof("app %v has grant for %v with scope %v", app, user, ar.Scope)

		if strings.Contains(ar.Scope, "profile") {
			ar.UserData = map[string]interface{}{
				"user_id":      user.ID,
				"name":         user.Name,
				"email":        user.Email,
				"picture":      user.Icon,
				"phone_number": user.Mobile,
				"roles":        user.GetAppRoleCodes(app.Code),
			}
		}
		svr.FinishAuthorizeRequest(resp, c.Request(), ar)
	}
	logger.Debugf("authorization response: --- %v ---", resp.Output)

	if resp.IsError && resp.InternalError != nil {
		c.Logger().Error("internal error: ", resp.InternalError)
		c.Logger().Error("error: ", resp.ErrorId)
	}
	if !resp.IsError {
		resp.Output["xtra"] = "uart"
	}
	return osin.OutputJSON(resp, c.Response(), c.Request())
}

func tokenHandler(c buffalo.Context) error {
	resp := svr.NewResponse()
	defer resp.Close()
	logger = c.Logger().WithField("category", "oauth2")

	if ar := svr.HandleAccessRequest(resp, c.Request()); ar != nil {
		ar.Authorized = true
		svr.FinishAccessRequest(resp, c.Request(), ar)
	}
	logger.Debugf("token response: --- %v ---", resp.Output)

	if resp.IsError && resp.InternalError != nil {
		c.Logger().Error("internal error: ", resp.InternalError)
		c.Logger().Error("error: ", resp.ErrorId)
	}
	if !resp.IsError {
		resp.Output["xtra"] = "uart"
	}
	return osin.OutputJSON(resp, c.Response(), c.Request())
}

func userInfoHandler(c buffalo.Context) error {
	resp := svr.NewResponse()
	defer resp.Close()
	logger = c.Logger().WithField("category", "oauth2")
	if ir := svr.HandleInfoRequest(resp, c.Request()); ir != nil {
		resp.Output = ir.AccessData.UserData.(map[string]interface{})
	}
	logger.Debugf("userinfo response: --- %v ---", resp.Output)
	return osin.OutputJSON(resp, c.Response(), c.Request())
}

//** Provider implementation ----------------------------------------------

// Provider handles authentication statuses and connections.
//
type Provider struct {
	clients   map[string]osin.Client
	authorize map[string]*osin.AuthorizeData
	access    map[string]*osin.AccessData
	refresh   map[string]string
}

// AppClient is struct for storing client information
// It implements osin.Client by osin.DefaultClient and store app additionally.
type AppClient struct {
	*osin.DefaultClient
	app *models.App
}

// GetApp returns app model of the client.
func (c *AppClient) GetApp() *models.App {
	return c.app
}

func newProvider() *Provider {
	stg := &Provider{
		clients:   make(map[string]osin.Client), // TODO: for cache, later
		authorize: make(map[string]*osin.AuthorizeData),
		access:    make(map[string]*osin.AccessData),
		refresh:   make(map[string]string),
	}
	return stg
}

// Clone just return it self.
func (s *Provider) Clone() osin.Storage {
	return s
}

// Close does nothing because there is no clone or additional resources.
func (s *Provider) Close() {
	for k, v := range s.access {
		if v.IsExpired() {
			logger.Debugf("delete expired ak for %v", v.Client.GetId()[0:8])
			s.RemoveAccess(k)
		}
	}
	for k, v := range s.authorize {
		if v.IsExpired() {
			logger.Debugf("rarely, delete expired auth_code %v", v.Code)
			s.RemoveAuthorize(k)
		}
	}
}

// GetClient searches and returns osin.Client instance with clientID.
// Called by HandleAuthorizeRequest, HandleAccessRequest, ...
func (s *Provider) GetClient(clientID string) (osin.Client, error) {
	logger.Debug("op.getclient invoked for: ", clientID[0:9])
	app := models.GetAppByKey(clientID)
	if app != nil {
		return &AppClient{
			DefaultClient: &osin.DefaultClient{
				Id:          app.AppKey,
				Secret:      app.AppSecret,
				RedirectUri: app.CallbackURL,
			},
			app: app,
		}, nil
	}
	return nil, osin.ErrNotFound
}

//** Functions for phase #1, authorize request ----------------------------

// SaveAuthorize stores authorization information into provider storage.
// Called by FinishAuthorizeRequest
func (s *Provider) SaveAuthorize(data *osin.AuthorizeData) error {
	logger.Debug("op.save authorize for ", data.Code[0:9])
	s.authorize[data.Code] = data
	return nil
}

//** Functions for phase #2, access token ---------------------------------

// LoadAuthorize read and returns authorize information from provider storage.
// Called by HandleAccessRequest
func (s *Provider) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	logger.Debug("op.load authorize for ", code[0:9])
	if d, ok := s.authorize[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

// SaveAccess stores given access data on the provider storage.
// Called by FinishAccessRequest, before call RemoveAuthorize
func (s *Provider) SaveAccess(data *osin.AccessData) error {
	logger.Debug("save access for ", data.AuthorizeData.Code[0:9])
	s.access[data.AccessToken] = data
	// TODO: do not store refresh token until fully implemented
	if false && data.RefreshToken != "" {
		s.refresh[data.RefreshToken] = data.AccessToken
	}
	return nil
}

// RemoveAuthorize delete authorize information from provider storage.
// Called by FinishAccessRequest, after call SaveAccess
func (s *Provider) RemoveAuthorize(code string) error {
	logger.Debug("op.remove authorize for ", code[0:9])
	delete(s.authorize, code)
	return nil
}

//** Functions for resource accessing -------------------------------------

// LoadAccess read and returns access information from provider storage.
// Called by HandleInfoRequest to validate access token
func (s *Provider) LoadAccess(code string) (*osin.AccessData, error) {
	logger.Debug("op.load access invoked for ", code[0:9])
	if d, ok := s.access[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

// Others: Refresh Token related things maybe.

// LoadRefresh is not used yet
func (s *Provider) LoadRefresh(code string) (*osin.AccessData, error) {
	logger.Debug("provider.loadrefresh invoked")
	if d, ok := s.refresh[code]; ok {
		return s.LoadAccess(d)
	}
	return nil, osin.ErrNotFound
}

// RemoveAccess is not used yet
func (s *Provider) RemoveAccess(code string) error {
	logger.Debug("provider.removeaccess invoked")
	delete(s.access, code)
	return nil
}

// RemoveRefresh is not used yet
func (s *Provider) RemoveRefresh(code string) error {
	logger.Debug("provider.removerefresh invoked")
	delete(s.refresh, code)
	return nil
}
