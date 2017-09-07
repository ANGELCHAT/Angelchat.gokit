package mongo

//
//import (
//	"fmt"
//
//	"github.com/sokool/gokit/mongo"
//	"gopkg.in/mgo.v2"
//)
//
//type mongoStorage struct {
//	dba *mongo.DB
//}
//
//type magg struct {
//	ID     Identity
//	Events []Event
//}
//
//func (s *mongoStorage) Save(id Identity, rs []Event) error {
//	return s.dba.Execute(fmt.Sprintf("tavern:%s", id), func(c *mgo.Collection) error {
//		for _, e := range rs {
//			c.Insert(e)
//		}
//		return nil
//	})
//}
//
//func (s *mongoStorage) Load(id Identity) ([]Event, error) {
//	n := fmt.Sprintf("tavern:%s", id)
//	var es []Event
//
//	return es, s.dba.Execute(n, func(c *mgo.Collection) error {
//		return c.Find(nil).All(&es)
//	})
//}
//
//func mongoStore(db *mongo.DB, collection string) Store {
//	return &mongoStorage{db}
//}
