package imageconfigs

type ImageWidthHeight struct {
	Small  int64 `bson:"small" json:"small"`
	Medium int64 `bson:"medium" json:"medium"`
	Large  int64 `bson:"large" json:"large"`
}

type CoverImage struct {
	PublicId string           `bson:"publicId" json:"publicId"`
	Url      string           `bson:"url" json:"url"`
	Width    ImageWidthHeight `bson:"width" json:"width"`
	Height   ImageWidthHeight `bson:"height" json:"height"`
}

var defaultWidth = ImageWidthHeight{Small: 65, Medium: 130, Large: 260}
var defaultHeight = ImageWidthHeight{Small: 95, Medium: 190, Large: 380}

func GetDefaultWidth() ImageWidthHeight {
	return defaultWidth
}

func GetDefaultHeight() ImageWidthHeight {
	return defaultHeight
}
