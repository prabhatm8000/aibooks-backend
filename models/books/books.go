package books

import (
	"example/aibooks-backend/config"
	"example/aibooks-backend/errorHandling"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookData struct {
	Id                 primitive.ObjectID `bson:"_id" json:"id"`
	Title              string             `bson:"title" json:"title"`
	Summary            string             `bson:"summary" json:"summary"`
	TotalChapters      int                `bson:"totalChapters" json:"totalChapters"`
	Genre              []string           `bson:"genre" json:"genre"`
	PdfUrl             string             `bson:"pdfUrl" json:"pdfUrl"`
	PdfPublicId        string             `bson:"pdfPublicId" json:"pdfPublicId"`
	CoverImageUrl      string             `bson:"coverImageUrl" json:"coverImageUrl"`
	CoverImagePublicId string             `bson:"coverImagePublicId" json:"coverImagePublicId"`
	CreatedAt          primitive.DateTime `bson:"createdAt" json:"createdAt"`
	TotalRatings       int                `bson:"totalRatings" json:"totalRatings"`
	SumRatings         float64            `bson:"sumRatings" json:"sumRatings"`
}

type BookDataShort struct {
	Id                 primitive.ObjectID `bson:"_id" json:"id"`
	Title              string             `bson:"title" json:"title"`
	Genre              []string           `bson:"genre" json:"genre"`
	CoverImageUrl      string             `bson:"coverImageUrl" json:"coverImageUrl"`
	CoverImagePublicId string             `bson:"coverImagePublicId" json:"coverImagePublicId"`
}

type RelatedBooks struct {
	RelatedBooks []BookData `bson:"relatedBooks" json:"relatedBooks"`
}

var BooksCollectionName string = "bookdatas"
var BooksCollection *mongo.Collection

func GetBookById(id string) (BookData, error) {
	if BooksCollection == nil {
		BooksCollection = config.GetCollection(BooksCollectionName)
	}

	var bookData BookData
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	idObj, err := primitive.ObjectIDFromHex(id)
	if err == primitive.ErrInvalidHex {
		return bookData, errorHandling.NewAPIError(400, GetBookById, "Invalid book id")
	} else if err != nil {
		return bookData, errorHandling.NewAPIError(500, GetBookById, err.Error())
	}

	err = BooksCollection.FindOne(ctx, bson.M{"_id": idObj}, &options.FindOneOptions{}).Decode(&bookData)
	if err == mongo.ErrNoDocuments {
		return bookData, errorHandling.NewAPIError(404, err, "Book not found")
	} else if err != nil {
		return bookData, errorHandling.NewAPIError(500, GetBookById, err.Error())
	}

	return bookData, nil
}

func GetAllBooks(page int64, limit int64, query string, sortBy string, sortOrder int64) ([]BookData, error) {
	if BooksCollection == nil {
		BooksCollection = config.GetCollection(BooksCollectionName)
	}

	var bookDatas []BookData
	ctx, cancel := config.GetDBCtx()
	defer cancel()
	skip := limit * (page - 1)

	var pipeline []bson.M

	if query != "" {
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{
				"$or": []bson.M{
					{
						"title": bson.M{
							"$regex":   query,
							"$options": "i",
						},
					},
					{
						"genre": bson.M{
							"$regex":   query,
							"$options": "i",
						},
					},
				},
			},
		})
	}

	pipeline = append(pipeline, bson.M{
		"$sort": bson.M{
			sortBy: sortOrder,
		},
	})

	pipeline = append(pipeline, bson.M{
		"$skip": skip,
	})

	pipeline = append(pipeline, bson.M{
		"$limit": limit,
	})

	cursor, err := BooksCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return bookDatas, errorHandling.NewAPIError(500, GetAllBooks, err.Error())
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &bookDatas)
	if err == mongo.ErrNilDocument {
		return bookDatas, errorHandling.NewAPIError(404, GetAllBooks, "No books found")
	} else if err != nil {
		return bookDatas, errorHandling.NewAPIError(500, GetAllBooks, err.Error())
	}

	return bookDatas, nil
}

func SearchSuggestions(query string, limit int64) ([]BookDataShort, error) {
	if BooksCollection == nil {
		BooksCollection = config.GetCollection(BooksCollectionName)
	}

	var suggestions []BookDataShort
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{
				"title": bson.M{
					"$regex":   query,
					"$options": "i",
				},
			},
			{
				"genre": bson.M{
					"$regex":   query,
					"$options": "i",
				},
			},
		},
	}

	cursor, err := BooksCollection.Find(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Projection: bson.M{
			"id":            1,
			"title":         1,
			"genre":         1,
			"coverImageUrl": 1,
			"publicId":      1,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &suggestions); err != nil {
		return suggestions, err
	}

	return suggestions, nil
}

func GetRelatedBooks(bookId string, limit int64) (RelatedBooks, error) {
	if BooksCollection == nil {
		BooksCollection = config.GetCollection(BooksCollectionName)
	}

	var relatedBooks RelatedBooks
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	idObj, err := primitive.ObjectIDFromHex(bookId)
	if err == primitive.ErrInvalidHex {
		return relatedBooks, errorHandling.NewAPIError(400, GetRelatedBooks, "Invalid book id")
	} else if err != nil {
		return relatedBooks, errorHandling.NewAPIError(500, GetRelatedBooks, err.Error())
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id": idObj,
			},
		},
		{
			"$project": bson.M{
				"genre": 1,
				"title": 1,
			},
		},
		{
			"$lookup": bson.M{
				"from": "bookdatas",
				"let": bson.M{
					"genres": "$genre",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$ne": []interface{}{"$_id", idObj},
									},
									{
										"$anyElementTrue": bson.M{
											"$map": bson.M{
												"input": "$genre",
												"as":    "genre",
												"in": bson.M{
													"$in": []interface{}{
														bson.M{"$toLower": "$$genre"},
														bson.M{
															"$map": bson.M{
																"input": "$$genres",
																"as":    "g",
																"in":    bson.M{"$toLower": "$$g"},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						"$limit": limit,
					},
				},
				"as": "relatedBooks",
			},
		},
		{
			"$project": bson.M{
				"relatedBooks": 1,
				"_id":          0,
			},
		},
	}

	cursor, err := BooksCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return relatedBooks, errorHandling.NewAPIError(500, GetRelatedBooks, err.Error())
	}
	defer cursor.Close(ctx)

	var result []RelatedBooks
	if err := cursor.All(ctx, &result); err != nil {
		return relatedBooks, errorHandling.NewAPIError(500, GetRelatedBooks, err.Error())
	}
	if len(result) == 0 {
		return relatedBooks, errorHandling.NewAPIError(404, GetRelatedBooks, "No related books found")
	}
	return result[0], nil
}
