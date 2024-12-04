package books

import (
	"example/aibooks-backend/config/imageconfigs"
	"example/aibooks-backend/models/books"
	"fmt"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookDataResponse struct {
	Id            primitive.ObjectID      `bson:"_id" json:"id"`
	Title         string                  `bson:"title" json:"title"`
	Summary       string                  `bson:"summary" json:"summary"`
	TotalChapters int                     `bson:"totalChapters" json:"totalChapters"`
	Genre         []string                `bson:"genre" json:"genre"`
	PdfUrl        string                  `bson:"pdfUrl" json:"pdfUrl"`
	PdfPublicId   string                  `bson:"pdfPublicId" json:"pdfPublicId"`
	CoverImage    imageconfigs.CoverImage `bson:"coverImage" json:"coverImage"`
	CreatedAt     primitive.DateTime      `bson:"createdAt" json:"createdAt"`
	Rating        float64                 `bson:"rating" json:"rating"`
	TotalRatings  int                     `bson:"totalRatings" json:"totalRatings"`
}

type BookDataShortResponse struct {
	Id         primitive.ObjectID      `bson:"_id" json:"id"`
	Title      string                  `bson:"title" json:"title"`
	Genre      []string                `bson:"genre" json:"genre"`
	CoverImage imageconfigs.CoverImage `bson:"coverImage" json:"coverImage"`
}

func GetAllBooks(c *gin.Context) {
	query := c.DefaultQuery("q", "")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	sortBy := c.DefaultQuery("sortBy", "title")
	sortOrder, _ := strconv.ParseInt(c.DefaultQuery("sortOrder", "1"), 10, 64)

	bookDatas, err := books.GetAllBooks(page, limit, query, sortBy, sortOrder)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	responseJson := make([]BookDataResponse, len(bookDatas))

	if len(bookDatas) != 0 {
		for i, bookData := range bookDatas {
			var rating float64
			if bookData.TotalRatings != 0 {
				rating = bookData.SumRatings / float64(bookData.TotalRatings)
			}

			responseJson[i] = BookDataResponse{
				Id:            bookData.Id,
				Title:         bookData.Title,
				Summary:       bookData.Summary,
				TotalChapters: bookData.TotalChapters,
				Genre:         bookData.Genre,
				PdfUrl:        bookData.PdfUrl,
				PdfPublicId:   bookData.PdfPublicId,
				CoverImage: imageconfigs.CoverImage{
					PublicId: bookData.CoverImagePublicId,
					Url:      bookData.CoverImageUrl,
					Width:    imageconfigs.GetDefaultWidth(),
					Height:   imageconfigs.GetDefaultHeight(),
				},
				CreatedAt:    bookData.CreatedAt,
				Rating:       rating,
				TotalRatings: bookData.TotalRatings,
			}
		}
	}

	c.IndentedJSON(200, responseJson)
}

func GetBookById(c *gin.Context) {
	id := c.Param("id")

	bookData, err := books.GetBookById(id)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	var rating float64 = 0
	if bookData.TotalRatings != 0 {
		rating = bookData.SumRatings / float64(bookData.TotalRatings)
	}

	responseJson := BookDataResponse{
		Id:            bookData.Id,
		Title:         bookData.Title,
		Summary:       bookData.Summary,
		TotalChapters: bookData.TotalChapters,
		Genre:         bookData.Genre,
		PdfUrl:        bookData.PdfUrl,
		PdfPublicId:   bookData.PdfPublicId,
		CoverImage: imageconfigs.CoverImage{
			PublicId: bookData.CoverImagePublicId,
			Url:      bookData.CoverImageUrl,
			Width:    imageconfigs.GetDefaultWidth(),
			Height:   imageconfigs.GetDefaultHeight(),
		},
		CreatedAt:    bookData.CreatedAt,
		Rating:       rating,
		TotalRatings: bookData.TotalRatings,
	}

	c.IndentedJSON(200, responseJson)
}

func SearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "5"), 10, 64)

	suggestions, err := books.SearchSuggestions(query, limit)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	responseJson := make([]BookDataShortResponse, len(suggestions))

	if len(suggestions) != 0 {
		for i := 0; i < len(suggestions); i++ {
			responseJson[i] = BookDataShortResponse{
				Id:    suggestions[i].Id,
				Title: suggestions[i].Title,
				Genre: suggestions[i].Genre,
				CoverImage: imageconfigs.CoverImage{
					PublicId: suggestions[i].CoverImagePublicId,
					Url:      suggestions[i].CoverImageUrl,
					Width:    imageconfigs.GetDefaultWidth(),
					Height:   imageconfigs.GetDefaultHeight(),
				},
			}
		}
	}

	c.IndentedJSON(200, responseJson)
}

func GetLatestBooks(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	session := sessions.Default(c)
	fmt.Println("session state in latest books:", session.Get("state"))

	latestBooks, err := books.GetAllBooks(1, limit, "", "createdAt", 1)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	responseJson := make([]BookDataResponse, len(latestBooks))

	if len(latestBooks) != 0 {
		for i, bookData := range latestBooks {

			var rating float64
			if bookData.TotalRatings != 0 {
				rating = bookData.SumRatings / float64(bookData.TotalRatings)
			}

			responseJson[i] = BookDataResponse{
				Id:    bookData.Id,
				Title: bookData.Title,
				Genre: bookData.Genre,
				CoverImage: imageconfigs.CoverImage{
					PublicId: bookData.CoverImagePublicId,
					Url:      bookData.CoverImageUrl,
					Width:    imageconfigs.GetDefaultWidth(),
					Height:   imageconfigs.GetDefaultHeight(),
				},
				Summary:       bookData.Summary,
				TotalChapters: bookData.TotalChapters,
				CreatedAt:     bookData.CreatedAt,
				Rating:        rating,
				TotalRatings:  bookData.TotalRatings,
			}
		}
	}

	c.IndentedJSON(200, responseJson)
}

func GetRelatedBooks(c *gin.Context) {
	id := c.Param("id")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	relatedBooks, err := books.GetRelatedBooks(id, limit)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	responseJson := make([]BookDataResponse, len(relatedBooks.RelatedBooks))

	if len(relatedBooks.RelatedBooks) != 0 {
		for i, bookData := range relatedBooks.RelatedBooks {

			var rating float64
			if bookData.TotalRatings != 0 {
				rating = bookData.SumRatings / float64(bookData.TotalRatings)
			}

			responseJson[i] = BookDataResponse{
				Id:    bookData.Id,
				Title: bookData.Title,
				Genre: bookData.Genre,
				CoverImage: imageconfigs.CoverImage{
					PublicId: bookData.CoverImagePublicId,
					Url:      bookData.CoverImageUrl,
					Width:    imageconfigs.GetDefaultWidth(),
					Height:   imageconfigs.GetDefaultHeight(),
				},
				Summary:       bookData.Summary,
				TotalChapters: bookData.TotalChapters,
				CreatedAt:     bookData.CreatedAt,
				Rating:        rating,
				TotalRatings:  bookData.TotalRatings,
			}
		}
	}

	c.IndentedJSON(200, responseJson)
}
