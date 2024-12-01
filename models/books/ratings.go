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

	if BooksCollection == nil {
		BooksCollection = config.GetCollection(BooksCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	client := config.GetDB().Client()
	session, err := client.StartSession()
	if err != nil {
		return primitive.NilObjectID, errorHandling.NewAPIError(500, "AddRating", "Failed to start session")
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		rating.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

		// Check if the user has already rated this book
		existingRating := Rating{}
		err := RatingsCollection.FindOne(sessCtx, bson.M{"userId": rating.UserId, "bookId": rating.BookId}).Decode(&existingRating)
		if err == nil && existingRating.UserId == rating.UserId && existingRating.BookId == rating.BookId {
			_, err = RatingsCollection.UpdateByID(sessCtx, existingRating.Id, bson.M{
				"$set": bson.M{
					"rating":    rating.Rating,
					"review":    rating.Review,
					"updatedAt": rating.UpdatedAt,
				},
			})
			if err != nil {
				return nil, err
			}

			// Adjust the book's ratings (existing rating updated)
			_, err := BooksCollection.UpdateOne(
				sessCtx,
				bson.M{"_id": rating.BookId},
				bson.D{
					{
						Key: "$inc", Value: bson.D{
							{Key: "sumRatings", Value: rating.Rating - existingRating.Rating},
						},
					},
				},
			)
			if err != nil {
				return nil, err
			}

			return existingRating.Id, nil
		}

		// Add a new rating
		rating.Id = primitive.NewObjectID()
		rating.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

		_, err = RatingsCollection.InsertOne(sessCtx, rating)
		if err != nil {
			return nil, err
		}

		// Adjust the book's ratings (new rating added)
		_, err = BooksCollection.UpdateOne(
			sessCtx,
			bson.M{"_id": rating.BookId},
			bson.D{
				{
					Key: "$inc", Value: bson.D{
						{Key: "sumRatings", Value: rating.Rating},
						{Key: "totalRatings", Value: 1},
					},
				},
			},
		)
		if err != nil {
			return nil, err
		}

		return rating.Id, nil
	})

	if err != nil {
		return primitive.NilObjectID, errorHandling.NewAPIError(500, "AddRating", err.Error())
	}

	return result.(primitive.ObjectID), nil
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
		return rating, errorHandling.NewAPIError(500, GetRatingsById, "Uh oh! Something went wrong.")
	}

	err = RatingsCollection.FindOne(ctx, bson.M{"_id": idObj}).Decode(&rating)
	if err == mongo.ErrNoDocuments {
		return rating, errorHandling.NewAPIError(404, err, "Rating not found")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetRatingsById, "Uh oh! Something went wrong.")
	}

	return rating, nil
}

func GetMyRatingForBookId(userId string, bookId string) (RatingResponse, error) {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	var rating RatingResponse
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err == primitive.ErrInvalidHex {
		return rating, errorHandling.NewAPIError(400, GetMyRatingForBookId, "Invalid user id")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetMyRatingForBookId, "Uh oh! Something went wrong.")
	}

	bookIdObj, err := primitive.ObjectIDFromHex(bookId)
	if err == primitive.ErrInvalidHex {
		return rating, errorHandling.NewAPIError(400, GetMyRatingForBookId, "Invalid book id")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetMyRatingForBookId, "Uh oh! Something went wrong.")
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userId": userIdObj,
				"bookId": bookIdObj,
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
	}

	cursor, err := RatingsCollection.Aggregate(ctx, pipeline)
	if err == mongo.ErrNoDocuments {
		return rating, errorHandling.NewAPIError(404, err, "Rating not found")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, GetMyRatingForBookId, err.Error())
	}

	var results []RatingResponse
	err = cursor.All(ctx, &results)
	if err != nil {
		return rating, errorHandling.NewAPIError(500, GetMyRatingForBookId, err.Error())
	}

	if len(results) == 0 {
		return rating, nil
	}
	return results[0], nil
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
		return ratings, errorHandling.NewAPIError(500, GetRatingsByBookId, err.Error())
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
		return ratings, errorHandling.NewAPIError(500, GetRatingsByBookId, err.Error())
	}

	err = cursor.All(ctx, &ratings)
	if err != nil {
		return ratings, errorHandling.NewAPIError(500, GetRatingsByBookId, err.Error())
	}

	return ratings, nil
}

func GetBookRatingSummary(bookId string) (float64, error) {
	if RatingsCollection == nil {
		RatingsCollection = config.GetCollection(RatingsCollectionName)
	}

	var rating float64
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	bookIdObj, err := primitive.ObjectIDFromHex(bookId)
	if err == primitive.ErrInvalidHex {
		return rating, errorHandling.NewAPIError(400, "GetBookRatingSummary", "Invalid book id")
	} else if err != nil {
		return rating, errorHandling.NewAPIError(500, "GetBookRatingSummary", err.Error())
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"bookId": bookIdObj,
			},
		},
		{
			"$group": bson.M{
				"_id":        "$bookId",
				"sumRatings": bson.M{"$sum": "$rating"},
				"count":      bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"ratingAverage": bson.M{
					"$divide": []interface{}{"$sumRatings", "$count"},
				},
			},
		},
	}

	cursor, err := RatingsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return rating, errorHandling.NewAPIError(500, "GetBookRatingSummary", err.Error())
	}

	var result []interface{}
	err = cursor.All(ctx, &result)
	if err != nil {
		return rating, errorHandling.NewAPIError(500, "GetBookRatingSummary", err.Error())
	}

	if len(result) == 0 {
		return rating, nil
	}

	// Process the first aggregation result
	doc, ok := result[0].(primitive.D)
	if !ok {
		return rating, errorHandling.NewAPIError(500, "GetBookRatingSummary", "Unexpected aggregation result format")
	}

	docMap := doc.Map()
	rating, ok = docMap["ratingAverage"].(float64)
	if !ok {
		return rating, errorHandling.NewAPIError(500, "GetBookRatingSummary", "Invalid ratingAverage format")
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
		return errorHandling.NewAPIError(500, DeleteRatingById, err.Error())
	}

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err == primitive.ErrInvalidHex {
		return errorHandling.NewAPIError(400, DeleteRatingById, "Invalid user id")
	} else if err != nil {
		return errorHandling.NewAPIError(500, DeleteRatingById, err.Error())
	}

	_, err = RatingsCollection.DeleteOne(ctx, bson.M{"_id": idObj, "userId": userIdObj})
	if err == mongo.ErrNoDocuments {
		return errorHandling.NewAPIError(404, err, "Rating not found")
	} else if err != nil {
		return errorHandling.NewAPIError(500, DeleteRatingById, err.Error())
	}
	return nil
}
