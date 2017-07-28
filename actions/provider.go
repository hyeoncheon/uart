//** OAuth2.0 authorization service provider code built on osin library.
//
//	It supports simple oauth2 authorization cycle with userinfo endpoint,
//	and currently does not support jwt and oidc.
//
// osin library: https://github.com/RangelReale/osin

package actions

// TODO CHECK RFC AND IMPLEMENT AGAIN BUT NOT NOW

import (
	"net/http"

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
	svr.AccessTokenGen = utils.NewRS256AccessTokenGen(brandName, privKeyFile)

	logger.Info("oauth2 provider with jwt support initialized!")
	return
}

func authorizeHandler(c buffalo.Context) error {
	resp := svr.NewResponse()
	defer resp.Close()
	logger = c.Logger().WithField("category", "oauth2")
	logger.Debug("oauth2 authorization started...")

	if ar := svr.HandleAuthorizeRequest(resp, c.Request()); ar != nil {
		app := models.GetAppByKey(ar.Client.GetId())
		user := currentMember(c)
		if !user.IsActive {
			c.Flash().Add("danger", t(c, "no.perm.inactive.member"))
			return c.Redirect(http.StatusTemporaryRedirect, "/membership/me")
		}
		logger.Infof("trying to grant access from %v against %v...", app, user)
		if ar.Authorized = user.HasGrantFor(app.ID); !ar.Authorized {
			logger.Error("NOT GRANTED!!!")
			c.Set("app", app)
			c.Set("scope", ar.Scope)
			c.Set("appkey", ar.Client.GetId())
			c.Session().Set("origin", c.Request().RequestURI)
			return c.Render(200, r.HTML("oauth2/grant.html"))
		}

		// TODO: implement scope handling
		logger.Debugf("--- requested scopes: %v", ar.Scope)
		ar.UserData = map[string]interface{}{
			"user_id":      user.ID,
			"name":         user.Name,
			"email":        user.Email,
			"picture":      user.Icon,
			"phone_number": user.Mobile,
			"roles":        user.GetAppRoleCodes(app.Code),
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
	logger.Info("oauth2 access token requested...")

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
	logger.Debugf("--- resp.output: %v", resp.Output)
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
	logger.Info("oauth2 provider.getclient invoked for: ", clientID)
	app := models.GetAppByKey(clientID)
	if app != nil {
		return &osin.DefaultClient{
			Id:          app.AppKey,
			Secret:      app.AppSecret,
			RedirectUri: app.CallbackURL,
		}, nil
	}
	return nil, osin.ErrNotFound
}

//** Functions for phase #1, authorize request ----------------------------

// SaveAuthorize stores authorization information into provider storage.
// Called by FinishAuthorizeRequest
func (s *Provider) SaveAuthorize(data *osin.AuthorizeData) error {
	logger.Info("oauth2 provider.save authorize for ", data.Code)
	logger.Debug("---- userdata: ", data.UserData)
	s.authorize[data.Code] = data
	return nil
}

//** Functions for phase #2, access token ---------------------------------

// LoadAuthorize read and returns authorize information from provider storage.
// Called by HandleAccessRequest
func (s *Provider) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	logger.Info("oauth2 provider.load authorize for ", code)
	if d, ok := s.authorize[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

// SaveAccess stores given access data on the provider storage.
// Called by FinishAccessRequest, before call RemoveAuthorize
func (s *Provider) SaveAccess(data *osin.AccessData) error {
	logger.Info("oauth2 provider.save access for ", data.AccessToken)
	logger.Debug("---- userdata: ", data.UserData)
	s.access[data.AccessToken] = data
	// TODO: do not store refresh token until fully implemented
	return nil
	if data.RefreshToken != "" {
		s.refresh[data.RefreshToken] = data.AccessToken
	}
	return nil
}

// RemoveAuthorize delete authorize information from provider storage.
// Called by FinishAccessRequest, after call SaveAccess
func (s *Provider) RemoveAuthorize(code string) error {
	logger.Info("oauth2 provider.remove authorize for ", code)
	delete(s.authorize, code)
	return nil
}

//** Functions for resource accessing -------------------------------------

// LoadAccess read and returns access information from provider storage.
// Called by HandleInfoRequest to validate access token
func (s *Provider) LoadAccess(code string) (*osin.AccessData, error) {
	logger.Info("oauth2 provider.load access invoked for ", code)
	if d, ok := s.access[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

// Others: Refresh Token related things maybe.

// LoadRefresh is not used yet
func (s *Provider) LoadRefresh(code string) (*osin.AccessData, error) {
	logger.Info("provider.loadrefresh invoked")
	if d, ok := s.refresh[code]; ok {
		return s.LoadAccess(d)
	}
	return nil, osin.ErrNotFound
}

// RemoveAccess is not used yet
func (s *Provider) RemoveAccess(code string) error {
	logger.Info("provider.removeaccess invoked")
	delete(s.access, code)
	return nil
}

// RemoveRefresh is not used yet
func (s *Provider) RemoveRefresh(code string) error {
	logger.Info("provider.removerefresh invoked")
	delete(s.refresh, code)
	return nil
}
