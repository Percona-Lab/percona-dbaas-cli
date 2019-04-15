package dbaas

type Backupper interface {
	CR()
}

func Backup(b Backupper) {}
