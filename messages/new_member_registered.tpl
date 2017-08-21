# New Member {{.String}} registered

New member {{.Name}} registered at {{.CreatedAt}}.

* Name     : {{.Name}}
* Email    : {{.Email}}

### Credentials
{{range $credential := .Credentials}}
* {{.String}}
  * Name    : {{$credential.Name}}
  * Email   : {{$credential.Email}}
  * Provider: {{$credential.Provider}}
  * UserID  : {{$credential.UserID}}

{{end}}
