package product

import (
	"context"
	"errors"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"github.com/ringbrew/newaim/productsearch/internal/domain/embedding"
	"log"
	"strings"
)

type MilvusStore struct {
	ctx    *domain.UseCaseContext
	client client.Client
}

type QueryVectorRequest struct {
	Input embedding.Vector
	Top   int
}

type QueryVectorRes struct {
	Id    string
	Score float32
}

type QueryVectorResponse struct {
	Data []QueryVectorRes
}

func newMilvusStore(ctx *domain.UseCaseContext) (*MilvusStore, error) {
	mc, err := client.NewClient(context.Background(), client.Config{
		Address:  ctx.Config.Miluvs.Endpoint,
		Username: ctx.Config.Miluvs.Username,
		Password: ctx.Config.Miluvs.Password,
	})
	if err != nil {
		return nil, err
	}

	_, err = mc.DescribeDatabase(context.Background(), ctx.Config.Miluvs.DB)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if err := mc.CreateDatabase(context.Background(), ctx.Config.Miluvs.DB); err != nil {
				log.Fatal(err.Error())
			}
		} else {
			return nil, err
		}
	}

	if err := mc.UsingDatabase(context.Background(), ctx.Config.Miluvs.DB); err != nil {
		log.Fatal(err.Error())
	}

	ms := &MilvusStore{
		ctx:    ctx,
		client: mc,
	}

	return ms, nil
}

const (
	PARTITION   = ""
	VectorField = "vector"
)

func (ms *MilvusStore) BatchCreate(ctx context.Context, ds []*Product) error {
	if len(ds) == 0 {
		return errors.New("empty ds")
	}

	dim := len(ds[0].Vector)

	c := len(ds)
	id := make([]string, 0, c)
	vector := make([][]float32, 0, c)

	for _, v := range ds {
		id = append(id, v.Id)
		vector = append(vector, v.Vector)
	}

	idCol := entity.NewColumnVarChar("id", id)
	vectorCol := entity.NewColumnFloatVector(VectorField, dim, vector)

	col, err := ms.collection(ctx, dim)
	if err != nil {
		return err
	}
	if _, err := ms.client.Insert(
		ctx, col, PARTITION,
		idCol,
		vectorCol,
	); err != nil {
		return err
	}
	return nil
}

func (ms *MilvusStore) Create(ctx context.Context, p Product) error {
	return ms.BatchCreate(ctx, []*Product{&p})
}

func (ms *MilvusStore) Update(ctx context.Context, p Product) error {
	if err := ms.Delete(ctx, p); err != nil {
		return err
	}

	if err := ms.Create(ctx, p); err != nil {
		return err
	}
	return nil
}

func (ms *MilvusStore) Delete(ctx context.Context, p Product) error {
	dim := len(p.Vector)
	col, err := ms.collection(ctx, dim)
	if err != nil {
		return err
	}
	expr := fmt.Sprintf("id in [%s]", sliceToStr([]string{p.Id}))
	return ms.client.Delete(ctx, col, PARTITION, expr)
}

func (ms *MilvusStore) Query(ctx context.Context, request QueryVectorRequest) (QueryVectorResponse, error) {
	dim := len(request.Input)

	col, err := ms.collection(ctx, dim)
	if err != nil {
		return QueryVectorResponse{}, err
	}

	if err := ms.client.LoadCollection(ctx, col, false); err != nil {
		return QueryVectorResponse{}, err
	}

	of := []string{"id"}

	vector := []entity.Vector{
		entity.FloatVector(request.Input),
	}

	sp, err := entity.NewIndexFlatSearchParam()
	rs, err := ms.client.Search(ctx, col, nil, "", of, vector, VectorField, entity.L2, request.Top, sp, client.WithSearchQueryConsistencyLevel(entity.ClStrong))
	if err != nil {
		return QueryVectorResponse{}, err
	}

	if rs == nil {
		return QueryVectorResponse{}, nil
	}

	data := make([]QueryVectorRes, 0)
	for _, sr := range rs {
		for i, s := range sr.Scores {
			id, err := sr.IDs.GetAsString(i)
			if err != nil {
				return QueryVectorResponse{}, err
			}

			qre := QueryVectorRes{
				Id:    id,
				Score: s,
			}

			data = append(data, qre)
		}
	}

	return QueryVectorResponse{
		Data: data,
	}, nil
}

func (ms *MilvusStore) collection(ctx context.Context, dim int) (string, error) {
	colName := fmt.Sprintf("product_vector_%d", dim)

	if exist, err := ms.client.HasCollection(ctx, colName); err != nil {
		return "", err
	} else if !exist {
		if err := ms.createCollection(ctx, colName, dim); err != nil {
			return "", err
		}
	}

	return colName, nil
}

func (ms *MilvusStore) createCollection(ctx context.Context, colName string, dim int) error {
	exist, err := ms.client.HasCollection(ctx, colName)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	schema := &entity.Schema{
		CollectionName: colName,
		Fields: []*entity.Field{
			entity.NewField().WithName("id").WithDataType(entity.FieldTypeVarChar).WithIsPrimaryKey(true).WithMaxLength(32).WithDescription("id"),
			entity.NewField().WithName(VectorField).WithDataType(entity.FieldTypeFloatVector).WithDim(int64(dim)).WithDescription(VectorField),
		},
		EnableDynamicField: true,
	}

	if err := ms.client.CreateCollection(ctx, schema, 2); err != nil {
		return err
	}

	_index, err := entity.NewIndexHNSW(entity.L2, 8, 64)
	if err != nil {
		return err
	}

	return ms.client.CreateIndex(ctx, colName, VectorField, _index, false)
}

func sliceToStr(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return "'" + strings.Join(arr, "', '") + "'"
}
