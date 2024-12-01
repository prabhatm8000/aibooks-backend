package userlibrarys

import (
	"context"
	"example/aibooks-backend/config"
	"example/aibooks-backend/errorHandling"
	"example/aibooks-backend/models/books"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserLibrary struct {
	Id         primitive.ObjectID   `bson:"_id" json:"id"`
	UserId     primitive.ObjectID   `bson:"userId" json:"userId"`
	BookIds    []primitive.ObjectID `bson:"bookIds" json:"bookIds"`
	TotalBooks int64                `bson:"totalBooks" json:"totalBooks"`
}

type UserLibraryResponse struct {
	Id         primitive.ObjectID `bson:"_id" json:"id"`
	UserId     primitive.ObjectID `bson:"userId" json:"userId"`
	Books      []books.BookData   `bson:"books" json:"books"`
	TotalBooks int64              `bson:"totalBooks" json:"totalBooks"`
}

var UserLibraryCollectionName = "userlibrarys"
var UserLibraryCollection *mongo.Collection

func InitLibrary(ctx context.Context, userId primitive.ObjectID, bookId primitive.ObjectID) (primitive.ObjectID, error) {
	if UserLibraryCollection == nil {
		UserLibraryCollection = config.GetCollection(UserLibraryCollectionName)
	}

	var library UserLibrary
	library.Id = primitive.NewObjectID()
	library.UserId = userId
	library.TotalBooks = 1
	library.BookIds = []primitive.ObjectID{
		bookId,
	}

	result, err := UserLibraryCollection.InsertOne(ctx, library)
	if err != nil {
		return primitive.NilObjectID, errorHandling.NewAPIError(500, AddBookToLibrary, err.Error())
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func IsBookInLibrary(userId primitive.ObjectID, bookId primitive.ObjectID, ctx context.Context) (bool, error) {
	if UserLibraryCollection == nil {
		UserLibraryCollection = config.GetCollection(UserLibraryCollectionName)
	}

	var c context.Context = ctx
	if ctx == nil {
		ctx, cancel := config.GetDBCtx()
		defer cancel()
		c = ctx
	}

	filter := bson.M{
		"userId":  userId,
		"bookIds": bson.M{"$in": []primitive.ObjectID{bookId}},
	}
	result := UserLibraryCollection.FindOne(c, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else if result.Err() != nil {
		return false, errorHandling.NewAPIError(500, AddBookToLibrary, result.Err().Error())
	}

	return true, nil
}

func AddBookToLibrary(userId primitive.ObjectID, bookId primitive.ObjectID) error {
	if UserLibraryCollection == nil {
		UserLibraryCollection = config.GetCollection(UserLibraryCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	exists, err := IsBookInLibrary(userId, bookId, ctx)
	if err != nil {
		return errorHandling.NewAPIError(500, AddBookToLibrary, err.Error())
	}

	if exists {
		return nil
	}

	filter := bson.M{"userId": userId}
	update := bson.M{
		"$addToSet": bson.M{"bookIds": bookId},
		"$inc":      bson.M{"totalBooks": 1},
	}

	result, err := UserLibraryCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errorHandling.NewAPIError(500, AddBookToLibrary, err.Error())
	}

	if result.MatchedCount == 0 {
		_, err = InitLibrary(ctx, userId, bookId)
		if err != nil {
			return errorHandling.NewAPIError(500, AddBookToLibrary, err.Error())
		}
	}

	return nil
}

func RemoveBookFromLibrary(userId primitive.ObjectID, bookId primitive.ObjectID) error {
	if UserLibraryCollection == nil {
		UserLibraryCollection = config.GetCollection(UserLibraryCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	filter := bson.M{"userId": userId}
	update := bson.M{
		"$pull": bson.M{"bookIds": bookId},
		"$inc":  bson.M{"totalBooks": -1},
	}

	result, err := UserLibraryCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errorHandling.NewAPIError(500, RemoveBookFromLibrary, err.Error())
	}

	if result.MatchedCount == 0 {
		return errorHandling.NewAPIError(404, RemoveBookFromLibrary, "Library not found")
	}

	return nil
}

func GetLibraryByUserId(userId primitive.ObjectID, page int64, limit int64) (UserLibraryResponse, error) {
	if UserLibraryCollection == nil {
		UserLibraryCollection = config.GetCollection(UserLibraryCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	var results []UserLibraryResponse
	var library UserLibraryResponse
	var pipeline []bson.M

	skip := (page - 1) * limit
	pipeline = append(pipeline, bson.M{"$match": bson.M{"userId": userId}})

	// Slice the bookIds array first
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"bookIds":    bson.M{"$slice": []interface{}{"$bookIds", skip, limit}},
			"userId":     1,
			"totalBooks": 1,
		},
	})

	pipeline = append(pipeline, bson.M{
		"$lookup": bson.M{
			"from":         "bookdatas",
			"localField":   "bookIds",
			"foreignField": "_id",
			"as":           "books",
		},
	})

	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"id":         1,
			"userId":     1,
			"books":      1,
			"totalBooks": 1,
		},
	})

	cursor, err := UserLibraryCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return library, errorHandling.NewAPIError(500, "GetLibraryByUserId", err.Error())
	}

	if err = cursor.All(ctx, &results); err != nil {
		return library, errorHandling.NewAPIError(500, "GetLibraryByUserId", err.Error())
	}

	if len(results) > 0 {
		library = results[0]
	}

	return library, nil
}
