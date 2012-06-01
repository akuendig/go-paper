package main

import (
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"log"
)

var initialSession map[string]*mgo.Session

func init() {
	initialSession = make(map[string]*mgo.Session)

	// Dial to Tagesanzeiger
	var session, err = mgo.Dial(TagiUrl)

	if err != nil {
		log.Fatal(err)
	}

	session.DB("tagi").Login(MongoUser, MongoPassword)
	initialSession["tagi"] = session

	// Dial to blick
	session, err = mgo.Dial(BlickUrl)

	if err != nil {
		log.Fatal(err)
	}
	session.DB("blick").Login(MongoUser, MongoPassword)
	initialSession["blick"] = session

	// Dial to 20 Minuten
	session, err = mgo.Dial(MinutenUrl)

	session.DB("min20").Login(MongoUser, MongoPassword)
	initialSession["min20"] = session
}

func copyDb(database string) (*mgo.Session, *mgo.Database) {
	var initial = initialSession[database]
	var copy = initial.Copy()

	return copy, copy.DB(database)
}

func ReadBatch(database string, skip, take int) ([]*Article, error) {
	var session, db = copyDb(database)
	var a []*Article

	defer session.Close()
	var err = db.C("articles").Find(nil).Skip(skip).Limit(take).All(&a)

	return a, err
}

func ReadOldBatch(database string, skip, take int) ([]*Article, error) {
	var session, db = copyDb(database)
	var a []*Article

	defer session.Close()
	var err = db.C("articles").
		Find(bson.M{"site": bson.M{"$exists": true}}).
		Skip(skip).
		Limit(take).
		All(&a)

	return a, err
}

func UpdateBatch(database string, batch []*Article) error {
	var session, db = copyDb(database)
	var c = db.C("articles")

	defer session.Close()

	for _, a := range batch {
		if err := c.Update(bson.M{"id": a.Id}, a); err != nil {
			return err
		}
	}

	return nil
}

func UpdateWebsiteBatch(database string, batch []*Article) error {
	var session, db = copyDb(database)
	var c = db.C("articles")

	defer session.Close()

	for _, a := range batch {
		var err error

		if a.SiteData == nil {
			err = c.Update(
				bson.M{"id": a.Id},
				bson.M{"websiteraw": a.WebsiteRaw, "$unset": bson.M{"site": 1}})
		} else {
			err = c.Update(
				bson.M{"id": a.Id},
				bson.M{"websiteraw": a.WebsiteRaw, "site": a.SiteData})
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func NewIds(database string, ids []string) ([]string, error) {
	var session, db = copyDb(database)

	defer session.Close()

	var exist []struct{ id string }
	var err = db.C("articles").
		Find(bson.M{"id": ids}).
		Select(bson.M{"id": 1}).
		All(&exist)

	if err != nil {
		return nil, err
	}

	var res []string

outer:
	for i := range ids {
		for _, id := range exist {
			if id.id == ids[i] {
				continue outer
			}
		}

		res = append(res, ids[i])
	}

	return res, nil
}

func Articles(database string) (*mgo.Iter, func()) {
	var session, db = copyDb(database)

	return db.C("articles").Find(nil).Iter(), func() { session.Close() }
}

func Update(database string, a *Article) error {
	var session, db = copyDb(database)

	defer session.Close()

	return db.C("articles").Update(bson.M{"id": a.Id}, a)
}
