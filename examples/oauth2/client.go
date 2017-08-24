package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
)

func main() {
	client := &oauth2.Config{
		ClientID:     os.Getenv("UART_CLIENT_ID"),
		ClientSecret: os.Getenv("UART_SECRET_KEY"),
		Scopes:       []string{"profile", "auth:all"},
		RedirectURL:  "http://localhost:3090/auth/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:3000/oauth/authorize",
			TokenURL: "http://localhost:3000/oauth/token",
		},
	}
	userinfo := "http://localhost:3000/userinfo"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>"))
		w.Write([]byte("<p>click button below to run:</p>"))
		w.Write([]byte(fmt.Sprintf(`<a href="%s">Login</a>`,
			client.AuthCodeURL("state"))))
		w.Write([]byte("</boby></html>"))
	})

	http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		code := r.Form.Get("code")
		fmt.Println("# callback request! -----------------------------------------")
		fmt.Printf("phase #1: get authorization code: %v\n", code)

		t, err := client.Exchange(context.Background(), code)
		if err != nil {
			fmt.Printf("token error: %v\n", err)
		}
		fmt.Printf("phase #2: get access token: %v %v %v\n",
			t.TokenType, t.AccessToken, t.RefreshToken)

		// try to decode access token as it is jwt.
		const pemfile = "../../files/jwt.public.pem"
		jt, err := jwt.Parse(t.AccessToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, err
			}
			k, err := ioutil.ReadFile(pemfile)
			if err != nil {
				return nil, err
			}
			return jwt.ParseRSAPublicKeyFromPEM(k)
		})
		if claims, ok := jt.Claims.(jwt.MapClaims); ok && jt.Valid {
			if err := claims.Valid(); err == nil {
				fmt.Println("\nok, token validated! see it!")
				now := time.Now().Unix()
				fmt.Println(" *- issuer: ", claims.VerifyIssuer("UART", true))
				fmt.Println(" *- audience: ", claims.VerifyAudience(client.ClientID, true))
				fmt.Println(" *- not expired: ", claims.VerifyExpiresAt(now, true))
				fmt.Println(" *- issued at: ", claims.VerifyIssuedAt(now, true))
				fmt.Println(" *- not before: ", claims.VerifyNotBefore(now, true))
				fmt.Printf(" v- issued at: %v\n", time.Unix(int64(claims["iat"].(float64)), 0))
				fmt.Printf(" v- expires at: %v\n", time.Unix(int64(claims["exp"].(float64)), 0))
				fmt.Printf(" v- not before: %v\n", time.Unix(int64(claims["nbf"].(float64)), 0))
			}
			for k, v := range claims {
				if len(k) > 3 {
					fmt.Printf(" -- %s: %v (%T)\n", k, v, v)
				}
			}
			fmt.Println("")
		} else {
			fmt.Println("")
			fmt.Printf("the token is not valid jwt: %v\n", err)
			fmt.Println("")
		}

		req, _ := http.NewRequest("GET", userinfo, nil)
		req.Header.Set("Authorization", "Bearer "+t.AccessToken)
		ctx, cancel := context.WithTimeout(context.TODO(), 200*time.Millisecond)
		defer cancel()
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			fmt.Printf("userinfo error: %v\n", err)
		}
		defer resp.Body.Close()

		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("phase #3: get userinfo: \n%v\n", string(data))
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	fmt.Printf("starting test client... connect to http://localhost:3090\n")
	http.ListenAndServe(":3090", nil)
}
