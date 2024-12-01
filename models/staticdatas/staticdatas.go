package staticdatas

import (
	"example/aibooks-backend/config"
	"example/aibooks-backend/errorHandling"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StaticData struct {
	Id       primitive.ObjectID `bson:"_id" json:"id"`
	DataType string             `bson:"dataType" json:"dataType"`
	Data     any                `bson:"data" json:"data"`
}

var StaticDataCollectionName string = "staticdatas"
var StaticDataCollection *mongo.Collection

func GetStaticDataByType(dataType string) (StaticData, error) {
	if StaticDataCollection == nil {
		StaticDataCollection = config.GetCollection(StaticDataCollectionName)
	}

	var staticData StaticData
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	err := StaticDataCollection.FindOne(ctx, bson.M{"dataType": dataType}).Decode(&staticData)
	if err == mongo.ErrNoDocuments {
		return staticData, errorHandling.NewAPIError(404, GetStaticDataByType, "Static data not found")
	} else if err != nil {
		return staticData, errorHandling.NewAPIError(500, GetStaticDataByType, err.Error())
	}

	return staticData, nil
}
