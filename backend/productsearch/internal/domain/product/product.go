package product

import (
	"github.com/ringbrew/newaim/productsearch/internal/domain/embedding"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Product struct {
	Id          string           `bson:"id" json:"id"`
	CreateTime  time.Time        `bson:"createTime" json:"createTime"`
	UpdateTime  time.Time        `bson:"updateTime" json:"updateTime"`
	SKU         string           `bson:"sku" json:"sku"`
	Title       string           `bson:"title" json:"title"`
	Description string           `bson:"description" json:"description"`
	Vector      embedding.Vector `bson:"vector" json:"vector,omitempty"`
	Score       float64          `bson:"-" json:"score,omitempty"`
}

func (p *Product) GetId() string {
	return p.Id
}

func (p *Product) SetId(id string) {
	p.Id = id
}

func GetColName() string {
	return "Product"
}

type IdGenerator interface {
	NewId() string
}

type BsonIdGenerator struct {
}

func NewIdGenerator() IdGenerator {
	return &BsonIdGenerator{}
}

func (g *BsonIdGenerator) NewId() string {
	return primitive.NewObjectID().Hex()
}
