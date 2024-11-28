package books

import (
	"example/aibooks-backend/models/books"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImageWidthHeight struct {
	Small  int64 `bson:"small" json:"small"`
	Medium int64 `bson:"medium" json:"medium"`
	Large  int64 `bson:"large" json:"large"`
}

var defaultWidth = ImageWidthHeight{Small: 65, Medium: 130, Large: 260}
var defaultHeight = ImageWidthHeight{Small: 95, Medium: 190, Large: 380}

type CoverImage struct {
	PublicId string           `bson:"publicId" json:"publicId"`
	Url      string           `bson:"url" json:"url"`
	Width    ImageWidthHeight `bson:"width" json:"width"`
	Height   ImageWidthHeight `bson:"height" json:"height"`
}

type BookDataResponse struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Summary       string             `bson:"summary" json:"summary"`
	TotalChapters int                `bson:"totalChapters" json:"totalChapters"`
	Genre         []string           `bson:"genre" json:"genre"`
	PdfUrl        string             `bson:"pdfUrl" json:"pdfUrl"`
	PdfPublicId   string             `bson:"pdfPublicId" json:"pdfPublicId"`
	CoverImage    CoverImage         `bson:"coverImage" json:"coverImage"`
	CreatedAt     primitive.DateTime `bson:"createdAt" json:"createdAt"`
	Rating        interface{}        `bson:"rating" json:"rating"`
}

type BookDataShortResponse struct {
	Id         primitive.ObjectID `bson:"_id" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Genre      []string           `bson:"genre" json:"genre"`
	CoverImage CoverImage         `bson:"coverImage" json:"coverImage"`
}

func GetAllBooks(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	sortBy := c.DefaultQuery("sortBy", "title")
	sortOrder, _ := strconv.ParseInt(c.DefaultQuery("sortOrder", "1"), 10, 64)

	bookDatas, err := books.GetAllBooks(pageSize, page, sortBy, sortOrder, query)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": err.Error()})
		return
	}

	responseJson := make([]BookDataResponse, len(bookDatas))

	if len(bookDatas) != 0 {
		for i, bookData := range bookDatas {
			responseJson[i] = BookDataResponse{
				Id:    bookData.Id,
				Title: bookData.Title,
				Genre: bookData.Genre,
				CoverImage: CoverImage{
					PublicId: bookData.CoverImagePublicId,
					Url:      bookData.CoverImageUrl,
					Width:    defaultWidth,
					Height:   defaultHeight,
				},
				Summary:       bookData.Summary,
				TotalChapters: bookData.TotalChapters,
				CreatedAt:     bookData.CreatedAt,
			}
		}
	}

	c.IndentedJSON(200, gin.H{
		"pageInfo": gin.H{
			"pageSize":  pageSize,
			"page":      page,
			"sortBy":    sortBy,
			"sortOrder": sortOrder,
		},
		"data": bookDatas,
	})
}

func GetBookById(c *gin.Context) {
	id := c.Param("id")

	bookData, err := books.GetBookById(id)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": err.Error()})
		return
	}

	var bookRating interface{}
	r, e := books.GetBookRatingSummary(id)
	if e == nil {
		bookRating = r
	}

	responseJson := BookDataResponse{
		Id:            bookData.Id,
		Title:         bookData.Title,
		Summary:       bookData.Summary,
		TotalChapters: bookData.TotalChapters,
		Genre:         bookData.Genre,
		PdfUrl:        bookData.PdfUrl,
		PdfPublicId:   bookData.PdfPublicId,
		CoverImage: CoverImage{
			PublicId: bookData.CoverImagePublicId,
			Url:      bookData.CoverImageUrl,
			Width:    defaultWidth,
			Height:   defaultHeight,
		},
		CreatedAt: bookData.CreatedAt,
		Rating:    bookRating,
	}

	c.IndentedJSON(200, responseJson)
}

func SearchSuggestions(c *gin.Context) {
	query := c.Query("query")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "5"), 10, 64)

	suggestions, err := books.SearchSuggestions(query, limit)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": err.Error()})
		return
	}

	responseJson := make([]BookDataShortResponse, len(suggestions))

	if len(suggestions) != 0 {
		for i := 0; i < len(suggestions); i++ {
			responseJson[i] = BookDataShortResponse{
				Id:    suggestions[i].Id,
				Title: suggestions[i].Title,
				Genre: suggestions[i].Genre,
				CoverImage: CoverImage{
					PublicId: suggestions[i].CoverImagePublicId,
					Url:      suggestions[i].CoverImageUrl,
					Width:    defaultWidth,
					Height:   defaultHeight,
				},
			}
		}
	}

	c.IndentedJSON(200, responseJson)
}

func GetLatestBooks(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)

	latestBooks, err := books.GetAllBooks(limit, 1, "createdAt", -1, "")
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": err.Error()})
		return
	}

	responseJson := make([]BookDataResponse, len(latestBooks))

	if len(latestBooks) != 0 {
		for i, bookData := range latestBooks {
			responseJson[i] = BookDataResponse{
				Id:    bookData.Id,
				Title: bookData.Title,
				Genre: bookData.Genre,
				CoverImage: CoverImage{
					PublicId: bookData.CoverImagePublicId,
					Url:      bookData.CoverImageUrl,
					Width:    defaultWidth,
					Height:   defaultHeight,
				},
				Summary:       bookData.Summary,
				TotalChapters: bookData.TotalChapters,
				CreatedAt:     bookData.CreatedAt,
			}
		}
	}

	c.IndentedJSON(200, responseJson)
}
