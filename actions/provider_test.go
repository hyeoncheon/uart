package actions_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hyeoncheon/uart/models"
)

//! TODO: testcase for missing parameter, especially for scope, state

func (as *ActionSuite) Test_OAuth2Provider_A_Authorize() {
	as.setupMembers()

	//! AUTHORIZE ---------------------------------------------------
	// call authorize before login, returns redirect to login page
	res := as.HTML("/oauth/authorize").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/login", res.HeaderMap.Get("Location"))

	as.loginAs(other) //! login as other

	// call authorize without guery string, returns bad request
	res = as.HTML("/oauth/authorize").Get()
	as.Equal(http.StatusBadRequest, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	// create testing app and make valid query string
	processFlowAppmanRole(as)
	successCreateTestingApp(as)
	app := models.GetAppByCode(AppCode)
	qs := "client_id=" + app.AppKey
	qs += "&response_type=code&scope=profile+auth%3Aall&state=state"
	qs += "&redirect_uri=" + url.QueryEscape(app.CallbackURL)

	as.loginAs(other) //! login as other again

	// call authorize by inactive member, redirect to "you're not active"
	res = as.HTML("/oauth/authorize?%v", qs).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))

	as.activateMember(other)

	// call authorize by active member, redirect to grant page
	res = as.HTML("/oauth/authorize?%v", qs).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Grant "+app.Name)

	//***** GRANT and redirect to origin (but root for this test)
	as.Session.Set("origin", "/landing")
	res = as.HTML("/grant/%v?scope=profile+auth:all", app.AppKey).Get()
	as.Equal(http.StatusTemporaryRedirect, res.Code)
	as.Equal("/landing", res.HeaderMap.Get("Location"))

	// call authorize by with bad callback, returns OK with invalid request
	res = as.HTML("/oauth/authorize?%vbadcallback", qs).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), `"error":"invalid_request"`)

	// call authorize by active, with valid guery string, after grant.
	res = as.HTML("/oauth/authorize?%v", qs).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), app.CallbackURL+"?code=")

	//! PREPARING TOKEN REQUEST
	query := map[string]string{}
	callback := res.HeaderMap.Get("Location")
	values, err := url.ParseQuery(strings.SplitN(callback, "?", 2)[1])
	if err != nil {
		as.Fail("Cannot parse query string!")
	}
	for k, v := range values {
		fmt.Printf("------ %v, %v (will take first elem)\n", k, v)
		query[k] = fmt.Sprintf("%v", v[0])
	}
	tokenRequest := url.Values{}
	tokenRequest.Add("grant_type", "authorization_code")
	tokenRequest.Add("code", query["code"])
	tokenRequest.Add("redirect_uri", app.CallbackURL)

	//! TOKEN -------------------------------------------------------
	// call token with retrieved values, returns valid access token
	req := as.HTML("/oauth/token")
	req.SetBasicAuth(app.AppKey, app.AppSecret)
	res = req.Post(tokenRequest)
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "access_token")

	retMap := make(map[string]*json.RawMessage)
	json.Unmarshal(res.Body.Bytes(), &retMap)
	value := ""
	json.Unmarshal(*retMap["access_token"], &value)
	fmt.Printf("Access Token: %v\n", value)
	as.Equal(3, len(strings.Split(value, ".")))
	accessToken := value
	//? is it enough? or testing the return value with decoder? currently no.

	time.Sleep(1 * time.Second)
	// call token with consumed authorize code, returns invalid error
	req = as.HTML("/oauth/token")
	req.SetBasicAuth(app.AppKey, app.AppSecret)
	res = req.Post(tokenRequest)
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "error")

	retMap = make(map[string]*json.RawMessage)
	json.Unmarshal(res.Body.Bytes(), &retMap)
	value = ""
	json.Unmarshal(*retMap["error"], &value)
	fmt.Printf("Error: %v\n", value)
	as.Equal("invalid_grant", value)

	//! USERINFO ----------------------------------------------------
	req = as.HTML("/userinfo")
	req.Headers["Authorization"] = "Bearer " + accessToken
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), `"name":"`+other.Name)
}
