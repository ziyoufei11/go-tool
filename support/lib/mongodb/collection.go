package mongodb

import "go.mongodb.org/mongo-driver/mongo"

func (m *MongoInfo) NewDatabaseCollection(database, table string) *mongo.Collection {
	return m.Client().Database(database).Collection(table)
}

func (m *MongoInfo) NewCollection(table string) *mongo.Collection {
	return m.Client().Database(m.Database).Collection(table)
}
