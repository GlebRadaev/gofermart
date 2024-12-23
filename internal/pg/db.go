package pg

type db struct {
	Database
}

func New(dbase Database) Database {
	return &db{
		Database: dbase,
	}
}
