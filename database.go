package main

import (
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"log"
)

var initialSession *mgo.Session

func init() {
	var session, err = mgo.Dial(Connectionstring)

	if err != nil {
		log.Fatal(err)
	}

	session.DB("tagi").Login(MongoUser, MongoPassword)
	session.DB("blick").Login(MongoUser, MongoPassword)
	session.DB("min20").Login(MongoUser, MongoPassword)

	initialSession = session
}

func ReadBatch(database string, skip, take int) ([]*Article, error) {
	var session = initialSession.Copy()
	var db = session.DB(database)
	var a []*Article

	defer session.Close()
	var err = db.C("articles").Find(nil).Skip(skip).Limit(take).All(&a)

	return a, err
}

func UpdateBatch(database string, batch []*Article) error {
	var session = initialSession.Copy()
	var db = session.DB(database)
	var c = db.C("articles")

	defer session.Close()

	for _, a := range batch {
		if err := c.Update(bson.M{"id": a.Id}, a); err != nil {
			return err
		}
	}

	return nil
}

func Articles(database string) (*mgo.Iter, func()) {
	var session = initialSession.Copy()
	var db = session.DB(database)

	return db.C("articles").Find(nil).Iter(), func() { session.Close() }
}

func Update(a *Article, database string) error {
	var session = initialSession.Copy()
	var db = session.DB(database)

	defer session.Close()

	return db.C("articles").Update(bson.M{"id": a.Id}, a)
}
