package otps

import (
	"example/aibooks-backend/config"
	"example/aibooks-backend/errorHandling"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Otp struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Otp       string             `bson:"otp" json:"otp"`
	ExpiresAt primitive.DateTime `bson:"expires_at" json:"expires_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at" json:"updated_at"`
}

var OtpsCollectionName string = "otps"
var OtpsCollection *mongo.Collection

func GenerateAndSaveOtpFor(email string) (string, error) {
	if OtpsCollection == nil {
		OtpsCollection = config.GetCollection(OtpsCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	var existingOtp Otp
	err := OtpsCollection.FindOne(ctx, bson.M{"email": email}).Decode(&existingOtp)
	if err == nil {
		if time.Now().Before(existingOtp.UpdatedAt.Time().Add(60 * time.Second)) {
			return "", errorHandling.ErrTooSoon
		}

		// Update existing OTP
		generatedOtp := generateOtp(6)
		update := bson.M{
			"$set": bson.M{
				"otp":        generatedOtp,
				"expires_at": primitive.NewDateTimeFromTime(time.Now().Add(time.Minute * 30)),
				"updated_at": primitive.NewDateTimeFromTime(time.Now()),
			},
		}
		_, err = OtpsCollection.UpdateOne(ctx, bson.M{"_id": existingOtp.Id}, update)
		if err != nil {
			return "", err
		}
		return generatedOtp, nil
	}

	// Insert new OTP
	generatedOtp := generateOtp(6)
	otp := Otp{
		Id:        primitive.NewObjectID(),
		Email:     email,
		Otp:       generatedOtp,
		ExpiresAt: primitive.NewDateTimeFromTime(time.Now().Add(time.Minute * 30)),
		UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err = OtpsCollection.InsertOne(ctx, otp)
	if err != nil {
		return "", err
	}
	return generatedOtp, nil
}

func generateOtp(length int) string {
	digits := []rune("0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

func CompareOtp(email string, otp string) bool {
	if OtpsCollection == nil {
		OtpsCollection = config.GetCollection(OtpsCollectionName)
	}

	ctx, cancel := config.GetDBCtx()
	defer cancel()

	var o Otp
	err := OtpsCollection.FindOne(ctx, bson.M{"email": email}).Decode(&o)
	if err != nil {
		return false
	}

	if o.Otp != otp {
		return false
	}

	if o.ExpiresAt.Time().Before(time.Now()) {
		return false
	}

	return true
}
