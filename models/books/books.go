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
}

type BookDataShort struct {
	Id                 primitive.ObjectID `bson:"_id" json:"id"`
	Title              string             `bson:"title" json:"title"`
	Genre              []string           `bson:"genre" json:"genre"`
	CoverImageUrl      string             `bson:"coverImageUrl" json:"coverImageUrl"`
	CoverImagePublicId string             `bson:"coverImagePublicId" json:"coverImagePublicId"`
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
		return bookData, errorHandling.NewAPIError(500, GetBookById, "Something went wrong")
	}

	err = BooksCollection.FindOne(ctx, bson.M{"_id": idObj}, &options.FindOneOptions{}).Decode(&bookData)
	if err == mongo.ErrNoDocuments {
		return bookData, errorHandling.NewAPIError(404, err, "Book not found")
	} else if err != nil {
		return bookData, errorHandling.NewAPIError(500, GetBookById, "Something went wrong")
	}

	return bookData, nil
}

func GetAllBooks(pageSize int64, page int64, sortBy string, sortOrder int64, query string) ([]BookData, error) {
	if BooksCollection == nil {
		BooksCollection = config.GetCollection(BooksCollectionName)
	}

	var bookDatas []BookData
	ctx, cancel := config.GetDBCtx()
	defer cancel()
	skip := pageSize * (page - 1)

	var filter bson.M

	if query != "" {
		filter = bson.M{
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
	}

	cursor, err := BooksCollection.Find(ctx, filter, &options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
		Sort:  bson.D{{Key: sortBy, Value: sortOrder}},
	})
	if err != nil {
		return bookDatas, errorHandling.NewAPIError(500, GetAllBooks, "Something went wrong")
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &bookDatas)
	if err == mongo.ErrNilDocument {
		return bookDatas, errorHandling.NewAPIError(404, GetAllBooks, "No books found")
	} else if err != nil {
		return bookDatas, errorHandling.NewAPIError(500, GetAllBooks, "Something went wrong")
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
