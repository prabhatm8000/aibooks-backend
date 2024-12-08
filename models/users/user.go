package users

import (
	"example/aibooks-backend/config"
	"example/aibooks-backend/errorHandling"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	Email         string             `bson:"email" json:"email"`
	EmailVerified bool               `bson:"email_verified" json:"email_verified"`
	FirstName     string             `bson:"first_name" json:"first_name"`
	LastName      string             `bson:"last_name" json:"last_name"`
	Password      string             `bson:"password" json:"password"`
	UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

var UsersCollectionName string = "users"
var UsersCollection *mongo.Collection

func AddUser(user Users) (primitive.ObjectID, error) {
	if UsersCollection == nil {
		UsersCollection = config.GetCollection(UsersCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	var existingUser Users
	err := UsersCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil && existingUser.Email == user.Email {
		_, err = UsersCollection.UpdateByID(ctx, existingUser.Id, bson.M{"$set": bson.M{
			"email":          user.Email,
			"email_verified": user.EmailVerified,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
			"password":       user.Password,
		}})
		if err != nil {
			return primitive.NilObjectID, errorHandling.NewAPIError(500, AddUser, err.Error())
		}
		return existingUser.Id, nil
	}

	user.Id = primitive.NewObjectID()

	doc, err := UsersCollection.InsertOne(ctx, user)
	if err != nil {
		return primitive.NilObjectID, errorHandling.NewAPIError(500, AddUser, err.Error())
	}

	return doc.InsertedID.(primitive.ObjectID), nil
}

func GetUserById(id string) (Users, error) {
	if UsersCollection == nil {
		UsersCollection = config.GetCollection(UsersCollectionName)
	}

	var user Users
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	idObj, err := primitive.ObjectIDFromHex(id)
	if err == primitive.ErrInvalidHex {
		return user, errorHandling.NewAPIError(400, GetUserById, "Invalid user id")
	} else if err != nil {
		return user, errorHandling.NewAPIError(500, GetUserById, err.Error())
	}

	err = UsersCollection.FindOne(ctx, bson.M{"_id": idObj}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return user, errorHandling.NewAPIError(404, err, "User not found")
	} else if err != nil {
		return user, errorHandling.NewAPIError(500, GetUserById, err.Error())
	}

	return user, nil
}

func GetUserByEmail(email string) (Users, error) {
	if UsersCollection == nil {
		UsersCollection = config.GetCollection(UsersCollectionName)
	}

	var user Users
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	err := UsersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return user, errorHandling.NewAPIError(404, err, "User not found")
	} else if err != nil {
		return user, errorHandling.NewAPIError(500, GetUserById, err.Error())
	}

	return user, nil
}
