package main

import (
	"launchpad.net/mgo"
	"log"
)

const (
	Connectionstring = "mongodb://ds029267.mongolab.com:29267"
)

func openConnection(database string) (db *mgo.Database, err error) {
	session, err := mgo.Dial(Connectionstring)

	if err != nil {
		return
	}

	db = session.DB(database)
	db.Login(MongoUser, MongoPassword)

	return
}

func FirstArticle() (article *Article, err error) {
	db, err := openConnection("tagi")

	if err != nil {
		log.Fatal(err)
		return
	}

	defer db.Session.Close()

	article = new(Article)
	err = db.C("articles").Find(nil).One(article)

	return
}
