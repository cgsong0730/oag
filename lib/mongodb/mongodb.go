package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	m "oag/model"

	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var conn *mongo.Client
var dbName string

func Init() {
	fmt.Println("mongodb init")
	dbName = "oag"
}

func ConnectDB() (client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	Authrization := getAuth()

	clientOptions := options.Client().ApplyURI("mongodb://" + Authrization.Hostname + ":" + Authrization.Port).SetAuth(
		options.Credential{
			Username: Authrization.Username,
			Password: Authrization.Password})

	client, err := mongo.Connect(ctx, clientOptions)

	checkErr(err)
	checkErr(client.Ping(ctx, readpref.Primary()))

	return client, ctx, cancel
}

func GetCollection(client *mongo.Client, colName string) *mongo.Collection {
	return client.Database(dbName).Collection(colName)
}

func getAuth() m.Auth {
	data, err := os.Open("./auth.json")
	checkErr(err)

	var auth m.Auth
	byteValue, _ := ioutil.ReadAll(data)
	json.Unmarshal(byteValue, &auth)

	return auth
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func CreateData(params, data string) {

	conn, ctx, cancel := ConnectDB()

	defer conn.Disconnect(ctx)
	defer cancel()

	filter := bson.M{"params": params}

	num, err := GetCollection(conn, "data").CountDocuments(ctx, filter)
	checkErr(err)

	newData := m.OpenData{
		Params: params,
		Data:   data,
	}

	if num == 0 {
		_, err := GetCollection(conn, "data").InsertOne(ctx, newData)
		checkErr(err)
	}
}

func ReadData(collection string, filter bson.M, sort bson.M) []bson.M {

	conn, ctx, cancel := ConnectDB()

	defer conn.Disconnect(ctx)
	defer cancel()

	var datas []bson.M
	res, err := GetCollection(conn, collection).Find(ctx, filter)
	checkErr(err)

	if err = res.All(ctx, &datas); err != nil {
		fmt.Println(err)
	}

	return datas
}

func UpdateData(index string, update bson.M) {

	conn, ctx, cancel := ConnectDB()

	defer conn.Disconnect(ctx)
	defer cancel()

	filter := bson.M{"index": index}

	update = bson.M{
		"$set": update,
	}

	_, err := GetCollection(conn, "data").UpdateOne(ctx, filter, update)
	checkErr(err)

}

func DeleteData(index string) {

	conn, ctx, cancel := ConnectDB()

	defer conn.Disconnect(ctx)
	defer cancel()

	filter := bson.M{"index": index}

	_, err := GetCollection(conn, "data").DeleteOne(ctx, filter)
	checkErr(err)

}
