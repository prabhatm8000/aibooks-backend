package userlibrarys

import (
	"example/aibooks-backend/config/imageconfigs"
	"example/aibooks-backend/controllers/books"
	"example/aibooks-backend/models/userlibrarys"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddBookToLibrary(c *gin.Context) {
	bookId := c.Param("bookId")

	userId := c.GetString("user_id")

	bookIdObj, err := primitive.ObjectIDFromHex(bookId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Invalid book id."})
		return
	}

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	err = userlibrarys.AddBookToLibrary(userIdObj, bookIdObj)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, gin.H{"message": "Book added to library."})
}

func RemoveBookFromLibrary(c *gin.Context) {
	bookId := c.Param("bookId")

	userId := c.GetString("user_id")

	bookIdObj, err := primitive.ObjectIDFromHex(bookId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Invalid book id."})
		return
	}

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	err = userlibrarys.RemoveBookFromLibrary(userIdObj, bookIdObj)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, gin.H{"message": "Book removed from library."})
}

func GetMyLibrary(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	userId := c.GetString("user_id")

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	library, err := userlibrarys.GetLibraryByUserId(userIdObj, page, limit)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	fmtBooks := make([]books.BookDataResponse, len(library.Books))

	if len(library.Books) != 0 {
		for i, bookData := range library.Books {
			var rating float64
			if bookData.TotalRatings != 0 {
				rating = bookData.SumRatings / float64(bookData.TotalRatings)
			}

			fmtBooks[i] = books.BookDataResponse{
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

	c.IndentedJSON(200, gin.H{
		"books":      fmtBooks,
		"totalBooks": library.TotalBooks,
		"id":         library.Id,
	})
}

func IsBookInLibrary(c *gin.Context) {
	bookId := c.Param("bookId")

	userId := c.GetString("user_id")

	bookIdObj, err := primitive.ObjectIDFromHex(bookId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Invalid book id."})
		return
	}

	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	isInLibrary, err := userlibrarys.IsBookInLibrary(userIdObj, bookIdObj, nil)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, gin.H{"isInLibrary": isInLibrary})
}
