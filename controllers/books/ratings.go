package books

import (
	"example/aibooks-backend/models/books"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddRating(c *gin.Context) {
	var rating books.Rating

	if err := c.ShouldBindJSON(&rating); err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	session := sessions.Default(c)
	userId := session.Get("user_id")
	userIdObj, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	rating.UserId = userIdObj

	ratingId, err := books.AddRating(rating)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, gin.H{"message": "Rating added successfully.", "ratingId": ratingId})
}

func GetRatingsById(c *gin.Context) {
	id := c.Param("bookId")

	rating, err := books.GetRatingsById(id)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, rating)
}

func GetMyRatingForBookId(c *gin.Context) {
	bookId := c.Param("bookId")

	session := sessions.Default(c)
	userId := session.Get("user_id")

	rating, err := books.GetMyRatingForBookId(userId.(string), bookId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, rating)
}

func GetRatingsByBookId(c *gin.Context) {
	bookId := c.Param("bookId")
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder, _ := strconv.ParseInt(c.DefaultQuery("sortOrder", "1"), 10, 64)

	rating, err := books.GetRatingsByBookId(bookId, int(limit), int(page), sortBy, int(sortOrder))
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, rating)
}

func DeleteRatingById(c *gin.Context) {
	id := c.Param("ratingId")
	session := sessions.Default(c)
	userId := session.Get("user_id")

	err := books.DeleteRatingById(id, userId.(string))
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, gin.H{"message": "Rating deleted successfully."})
}
