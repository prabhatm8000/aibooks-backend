package routes

import (
	"example/aibooks-backend/controllers/books"
	"example/aibooks-backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterBookdataRoutes(r *gin.RouterGroup) {
	booksGroup := r.Group("/books")
	{
		booksGroup.GET("/searchSuggestions", books.SearchSuggestions)
		booksGroup.GET("/", books.GetAllBooks)
		booksGroup.GET("/byId/:id", books.GetBookById)
		booksGroup.GET("/latest", books.GetLatestBooks)
	}

	ratingsGroup := booksGroup.Group("/ratings")
	{
		ratingsGroup.GET("/:bookId", books.GetRatingsByBookId)
		ratingsGroup.Use(middleware.IsAuthenticated).POST("/add", books.AddRating)
		ratingsGroup.Use(middleware.IsAuthenticated).GET("/myRatingFor/:bookId", books.GetRatingByUserIdBookId)
		ratingsGroup.Use(middleware.IsAuthenticated).DELETE("/delete/:ratingId", books.DeleteRatingById)
	}
}
