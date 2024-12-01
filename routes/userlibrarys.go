package routes

import (
	"example/aibooks-backend/controllers/userlibrarys"
	"example/aibooks-backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterLibraryRoutes(r *gin.RouterGroup) {
	libraryGroup := r.Group("/myLibrary")
	libraryGroup.Use(middleware.IsAuthenticated)
	{
		libraryGroup.GET("/getBooks", userlibrarys.GetMyLibrary)
		libraryGroup.PUT("/addBook/:bookId", userlibrarys.AddBookToLibrary)
		libraryGroup.DELETE("/removeBook/:bookId", userlibrarys.RemoveBookFromLibrary)
		libraryGroup.GET("/isBookInLibrary/:bookId", userlibrarys.IsBookInLibrary)
	}
}
