package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

func main() {
	client := &oauth2.Config{
		ClientID:     os.Getenv("UART_CLIENT_ID"),
		ClientSecret: os.Getenv("UART_SECRET_KEY"),
		Scopes:       []string{"id", "name", "email", "roles"},
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
		fmt.Printf("phase #1: get authorization code: %v\n", code)

		t, err := client.Exchange(context.Background(), code)
		if err != nil {
			fmt.Printf("token error: %v\n", err)
		}
		fmt.Printf("phase #2: get access token: %v %v %v\n",
			t.TokenType, t.AccessToken, t.RefreshToken)

		req, _ := http.NewRequest("GET", userinfo, nil)
		req.Header.Set("Authorization", "Bearer "+t.AccessToken)
		ctx, _ := context.WithTimeout(context.TODO(), 200*time.Millisecond)
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			fmt.Printf("userinfo error: %v\n", err)
		}
		defer resp.Body.Close()

		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("phase #3: get userinfo: \n%v\n", string(data))
		w.Write(data)
	})

	fmt.Printf("starting test client... connect to http://localhost:3090\n")
	http.ListenAndServe(":3090", nil)
}
