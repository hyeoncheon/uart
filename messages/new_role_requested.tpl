# New Role requested by {{.String}}

New role requested by {{.Name}}.

* Name     : {{.Name}}
* Email    : {{.Email}}

### Credentials
{{range $role := .Roles}}
* Role {{.String}}
  * Role        : {{$role.Name}}
  * Code        : {{$role.Code}}
  * Description : {{$role.Description}}
  * App         : {{$role.App.Name}}

{{end}}
