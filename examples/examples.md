# Examples

## Simple Client

Just compile and run with environment variables. the environment variable
`UART_CLIENT_ID` and `UART_SECRET_KEY` must be same as on main app.

```console
$ go build client.go && UART_CLIENT_ID=b8B2... UART_SECRET_KEY=Vt7D... ./client
starting test client... connect to http://localhost:3090
phase #1: get authorization code: ontdF-LlTnClipaoETKcPA
phase #2: get access token: Bearer qGBMbLVKTnSrnvLr7GIZbg hG5i2JFLRkS1WokFl-CVfA
phase #3: get userinfo: 
{"avatar_url":"https://lh6.googleusercontent.com/XXX/AAA/AAA/JJJ/photo.jpg","mail":"user@example.com","name":"Yonghwan SO","phone_number":"","roles":["admin"],"user_id":"2496b9b7"}

^C
$ 
```

