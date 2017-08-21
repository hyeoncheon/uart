# Membership status was {{if .IsActive}}Activated{{ else }}Locked{{end}}

Your membership status was {{if .IsActive}}Activated{{ else }}Locked{{end}} at {{.UpdatedAt}}.

* Name     : {{.Name}}
* Email    : {{.Email}}
* Status   : {{if .IsActive}}Active{{ else }}Locked{{end}}

