# Membership was {{if .IsActive}}Activated{{ else }}Locked{{end}} by Admin

Your membership was {{if .IsActive}}ACTIVATED{{ else }}LOCKED{{end}} at {{.UpdatedAt}} by administrator.

* Name     : {{.Name}}
* Email    : {{.Email}}
* Status   : {{if .IsActive}}Active{{ else }}Locked{{end}}

