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
	err = db.Login(MongoUser, MongoPassword)

	return
}

func PrintFirstArticle() (err error) {
	db, err := openConnection("tagi")

	if db != nil {
		defer db.Session.Close()
	}

	if err != nil {
		log.Fatal(err)
		return
	}

	var result Article

	err = db.C("articles").Find(nil).One(&result)

	log.Println(result)

	return nil
}
