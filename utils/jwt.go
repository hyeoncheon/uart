//** JWT utils built on top of dgrijalva's jwt-go.

package utils

import (
	"crypto/rsa"
	"io/ioutil"
	"log"

	"github.com/RangelReale/osin"
	"github.com/golang-jwt/jwt"
)

// RS256AccessTokenGen is JWT access token generator
type RS256AccessTokenGen struct {
	Issuer     string
	PrivateKey *rsa.PrivateKey
}

// NewRS256AccessTokenGen returns RS256AccessTokenGen as osin.AccessTokenGen
func NewRS256AccessTokenGen(issuer, key string) osin.AccessTokenGen {
	var t RS256AccessTokenGen
	if pem, err := ioutil.ReadFile(key); err != nil {
		log.Fatalf("cannot read private key file: %v", err)
	} else {
		if t.PrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(pem); err != nil {
			log.Fatalf("cannot parse private key: %v", err)
		}
	}
	t.Issuer = issuer
	return &t
}

// GenerateAccessToken implementation for JWT
func (g RS256AccessTokenGen) GenerateAccessToken(data *osin.AccessData, refresh bool) (string, string, error) {
	claims := jwt.MapClaims{
		"iss": g.Issuer,
		"sub": "Access Token",
		"aud": data.Client.GetId(),
		"exp": data.ExpireAt().Unix(),
		"iat": data.CreatedAt.Unix(),
		"nbf": data.CreatedAt.Unix() - 600,
	}
	if userdata, ok := data.UserData.(map[string]interface{}); ok {
		for k, v := range userdata {
			claims[k] = v
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	accesstoken, err := token.SignedString(g.PrivateKey)
	if !refresh || err != nil {
		return accesstoken, "", err
	}

	token = jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": g.Issuer,
		"sub": "Refresh Token",
		"aud": data.Client.GetId(),
		"iat": data.CreatedAt.Unix(),
	})
	refreshtoken, err := token.SignedString(g.PrivateKey)

	return accesstoken, refreshtoken, err
}
