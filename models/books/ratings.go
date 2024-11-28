package books

import (
	"example/aibooks-backend/config"
	"example/aibooks-backend/errorHandling"
	"example/aibooks-backend/models/users"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Rating struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	UserId    primitive.ObjectID `bson:"userId" json:"userId"`
	BookId    primitive.ObjectID `bson:"bookId" json:"bookId"`
	Rating    int                `bson:"rating" json:"rating"`
	Review    string             `bson:"review" json:"review"`
	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

type RatingResponse struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	UserId    primitive.ObjectID `bson:"userId" json:"userId"`
	BookId    primitive.ObjectID `bson:"bookId" json:"bookId"`
	Rating    int                `bson:"rating" json:"rating"`
	Review    string             `bson:"review" json:"review"`
	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
	User      users.UserShort    `bson:"user" json:"user"`
}

var RatingsCollectionName string = "ratings"
var RatingsCollection *mongo.Collection

func AddRating(rating Rating) (primitive.ObjectID, error) {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	rating.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	existingRating := Rating{}
	err := RatingsCollection.FindOne(ctx, bson.M{"userId": rating.UserId, "bookId": rating.BookId}).Decode(&existingRating)
	if err == nil && existingRating.UserId == rating.UserId && existingRating.BookId == rating.BookId {
		_, err = RatingsCollection.UpdateByID(ctx, existingRating.Id, bson.M{"$set": bson.M{
			"rating":    rating.Rating,
			"review":    rating.Review,
			"updatedAt": rating.UpdatedAt,
		}})
		if err != nil {
			return primitive.NilObjectID, errorHandling.NewAPIError(500, AddRating, err.Error())
		}
		return primitive.NilObjectID, nil
	}

	rating.Id = primitive.NewObjectID()
	rating.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	doc, err := RatingsCollection.InsertOne(ctx, rating)
	if err != nil {
		return primitive.NilObjectID, errorHandling.NewAPIError(500, AddRating, err.Error())
	}

	return doc.InsertedID.(primitive.ObjectID), nil
}

func GetRatingsById(id string) (Rating, error) {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	var rating Rating
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	idObj, err := primitive.ObjectIDFromHex(id)
	if err == primitive.ErrInvalidHex {
		return rating, errorHandling.NewAPIError(400, GetRatingsById, "Invalid rating id")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetRatingsById, "Something went wrong")
	}

	err = RatingsCollection.FindOne(ctx, bson.M{"_id": idObj}).Decode(&rating)
	if err == mongo.ErrNoDocuments {
		return rating, errorHandling.NewAPIError(404, err, "Rating not found")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetRatingsById, "Something went wrong")
	}

	return rating, nil
}

func GetRatingByUserIdBookId(userId string, bookId string) (RatingResponse, error) {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	var rating RatingResponse
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err == primitive.ErrInvalidHex {
		return rating, errorHandling.NewAPIError(400, GetRatingByUserIdBookId, "Invalid user id")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetRatingByUserIdBookId, "Something went wrong")
	}

	bookIdObj, err := primitive.ObjectIDFromHex(bookId)
	if err == primitive.ErrInvalidHex {
		return rating, errorHandling.NewAPIError(400, GetRatingByUserIdBookId, "Invalid book id")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetRatingByUserIdBookId, "Something went wrong")
	}

	err = RatingsCollection.FindOne(ctx, bson.M{"userId": userIdObj, "bookId": bookIdObj}).Decode(&rating)
	if err == mongo.ErrNoDocuments {
		return rating, errorHandling.NewAPIError(404, err, "Rating not found")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetRatingByUserIdBookId, "Something went wrong")
	}

	return rating, nil
}

func GetRatingsByBookId(bookId string, limit int, page int, sortBy string, sortOrder int) ([]RatingResponse, error) {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	var ratings []RatingResponse
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	idObj, err := primitive.ObjectIDFromHex(bookId)
	if err == primitive.ErrInvalidHex {
		return ratings, errorHandling.NewAPIError(400, GetRatingsByBookId, "Invalid book id")
	} else if err != nil {
		return ratings, errorHandling.NewAPIError(500, GetRatingsByBookId, "Something went wrong")
	}

	skip := limit * (page - 1)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"bookId": idObj,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "userId",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{
			"$unwind": "$user",
		},
		{
			"$project": bson.M{
				"_id":       1,
				"userId":    1,
				"bookId":    1,
				"rating":    1,
				"review":    1,
				"createdAt": 1,
				"updatedAt": 1,
				"user": bson.M{
					"_id":      1,
					"picture":  1,
					"name":     1,
					"nickname": 1,
				},
			},
		},
		{
			"$sort": bson.M{
				sortBy: sortOrder,
			},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := RatingsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return ratings, errorHandling.NewAPIError(500, GetRatingsByBookId, "Something went wrong")
	}

	err = cursor.All(ctx, &ratings)
	if err != nil {
		return ratings, errorHandling.NewAPIError(500, GetRatingsByBookId, "Something went wrong")
	}

	return ratings, nil
}

func GetBookRatingSummary(bookId string) (interface{}, error) {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	var rating interface{}
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	bookIdObj, err := primitive.ObjectIDFromHex(bookId)
	if err == primitive.ErrInvalidHex {
		return rating, errorHandling.NewAPIError(400, GetBookRatingSummary, "Invalid book id")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetBookRatingSummary, "Something went wrong")
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"book": bookIdObj,
			},
		},
		{
			"$group": bson.M{
				"_id": "$book",
				"rating": bson.M{
					"$sum": "$rating",
				},
			},
		},
	}

	cursor, err := RatingsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, errorHandling.NewAPIError(500, GetBookRatingSummary, "Something went wrong")
	}

	err = cursor.All(ctx, &rating)
	if err != nil {
		return 0, errorHandling.NewAPIError(500, GetBookRatingSummary, "Something went wrong")
	}

	return rating, nil
}

func DeleteRatingById(id string, userId string) error {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	idObj, err := primitive.ObjectIDFromHex(id)
	if err == primitive.ErrInvalidHex {
		return errorHandling.NewAPIError(400, DeleteRatingById, "Invalid rating id")
	} else if err != nil {
		return errorHandling.NewAPIError(500, DeleteRatingById, "Something went wrong")
	}

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err == primitive.ErrInvalidHex {
		return errorHandling.NewAPIError(400, DeleteRatingById, "Invalid user id")
	} else if err != nil {
		return errorHandling.NewAPIError(500, DeleteRatingById, "Something went wrong")
	}

	_, err = RatingsCollection.DeleteOne(ctx, bson.M{"_id": idObj, "userId": userIdObj})
	if err == mongo.ErrNoDocuments {
		return errorHandling.NewAPIError(404, err, "Rating not found")
	} else if err != nil {
		return errorHandling.NewAPIError(500, DeleteRatingById, "Something went wrong")
	}
	return nil
}
