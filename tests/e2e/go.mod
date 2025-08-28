module e2e-test

go 1.24.6

require (
	code.gitea.io/sdk/gitea v0.19.0
	github.com/Frantche/gitea-backup-restore-process v0.0.0
)

require (
	github.com/davidmz/go-pageant v1.0.2 // indirect
	github.com/go-fed/httpsig v1.1.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
)

replace github.com/Frantche/gitea-backup-restore-process => ../..
