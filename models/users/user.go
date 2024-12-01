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
	Aud           string             `bson:"aud" json:"aud"`
	Email         string             `bson:"email" json:"email"`
	EmailVerified bool               `bson:"email_verified" json:"email_verified"`
	Exp           int64              `bson:"exp" json:"exp"`
	FamilyName    string             `bson:"family_name" json:"family_name"`
	GivenName     string             `bson:"given_name" json:"given_name"`
	Iat           int64              `bson:"iat" json:"iat"`
	Iss           string             `bson:"iss" json:"iss"`
	Name          string             `bson:"name" json:"name"`
	Nickname      string             `bson:"nickname" json:"nickname"`
	Picture       string             `bson:"picture" json:"picture"`
	Sid           string             `bson:"sid" json:"sid"`
	Sub           string             `bson:"sub" json:"sub"`
	UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

type UserShort struct {
	Id         primitive.ObjectID `bson:"_id" json:"id"`
	Name       string             `bson:"name" json:"name"`
	GivenName  string             `bson:"given_name" json:"given_name"`
	FamilyName string             `bson:"family_name" json:"family_name"`
	Picture    string             `bson:"picture" json:"picture"`
	Nickname   string             `bson:"nickname" json:"nickname"`
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
	err := UsersCollection.FindOne(ctx, bson.M{"sub": user.Sub, "email": user.Email}).Decode(&existingUser)
	if err == nil && existingUser.Sub == user.Sub && existingUser.Email == user.Email {
		_, err = UsersCollection.UpdateByID(ctx, existingUser.Id, bson.M{"$set": bson.M{
			"aud":            user.Aud,
			"email_verified": user.EmailVerified,
			"exp":            user.Exp,
			"family_name":    user.FamilyName,
			"given_name":     user.GivenName,
			"iat":            user.Iat,
			"iss":            user.Iss,
			"name":           user.Name,
			"nickname":       user.Nickname,
			"picture":        user.Picture,
			"sid":            user.Sid,
			"sub":            user.Sub,
			"updated_at":     user.UpdatedAt,
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

func GetUserBySub(sub string) (Users, error) {
	if UsersCollection == nil {
		UsersCollection = config.GetCollection(UsersCollectionName)
	}

	var user Users
	ctx, cancel := config.GetDBCtx()
	defer cancel()

	err := UsersCollection.FindOne(ctx, bson.M{"sub": sub}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return user, errorHandling.NewAPIError(404, err, "User not found")
	} else if err != nil {
		return user, errorHandling.NewAPIError(500, GetUserBySub, err.Error())
	}

	return user, nil
}
