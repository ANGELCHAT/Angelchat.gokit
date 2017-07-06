package mongo

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

var session = map[string]*mgo.Session{}

type DB struct {
	database *mgo.Database
	session  *mgo.Session
}

func RegisterSession(name, url string) error {
	var err error

	session[name], err = mgo.Dial(url)
	if err != nil {
		return err
	}

	return nil
}

func Session(name string) (*DB, error) {
	ses, exists := session[name]
	if !exists {
		return nil, fmt.Errorf("session %v is not registered", name)
	}

	db := DB{
		session:  ses.Copy(),
		database: ses.DB(name),
	}

	return &db, nil
}

func (d *DB) Execute(collection string, f func(*mgo.Collection) error) error {
	return f(d.database.C(collection))
}

func (d *DB) Close() {
	d.session.Close()
}
