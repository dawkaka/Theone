package repository

import (
	"context"

	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/mongo"
)

type Reports struct {
	col *mongo.Collection
}

func (r *Reports) ReportPost(report entity.ReportPost) error {
	_, err := r.col.InsertOne(context.TODO(), report)
	return err
}

func (r *Reports) ReportCouple(report entity.ReportCouple) error {
	_, err := r.col.InsertOne(context.TODO(), report)
	return err
}

func NewReportRepo(col *mongo.Collection) Reports {
	return Reports{
		col: col,
	}
}
