package foods

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OffFoodDoc struct {
	Code        string         `bson:"code"`
	ProductName string         `bson:"product_name"`
	Brands      string         `bson:"brands"`
	Popularity  int64          `bson:"popularity_key"`
	ServingSize string         `bson:"serving_size"`
	Quantity    string         `bson:"quantity"`
	Nutriments  map[string]any `bson:"nutriments"`
}

type RepoMongoOFF struct {
	col *mongo.Collection
	searchMode string // "regex" or "text"
}

func NewRepoMongoOFF(client *mongo.Client, dbName, collection, searchMode string) *RepoMongoOFF {
	searchMode = strings.TrimSpace(strings.ToLower(searchMode))
	if searchMode == "" {
		searchMode = "regex"
	}
	return &RepoMongoOFF{
		col:        client.Database(dbName).Collection(collection),
		searchMode: searchMode,
	}
}

func NewMongoClient(ctx context.Context, mongoURI string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Database("admin").RunCommand(ctx, bson.M{"ping": 1}).Err(); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}

// Create a text index for name+brand search. Safe to call at startup.
func (r *RepoMongoOFF) EnsureTextIndex(ctx context.Context) error {
	// Always create the sort helper index (used by regex + cursor).
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "popularity_key", Value: -1}, {Key: "code", Value: 1}},
			Options: options.Index().SetName("idx_popularity_code"),
		},
	}
	if r.searchMode == "text" {
		models = append(models, mongo.IndexModel{
			Keys: bson.D{
				{Key: "product_name", Value: "text"},
				{Key: "brands", Value: "text"},
			},
			Options: options.Index().SetName("idx_text_name_brands").SetDefaultLanguage("none"),
		})
	}
	_, err := r.col.Indexes().CreateMany(ctx, models)
	return err
}

func (r *RepoMongoOFF) ByBarcode(ctx context.Context, code string) (*OffFoodDoc, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, NotFoundError{Msg: "empty barcode"}
	}

	filter := bson.M{
		"code": code,
		"product_name": bson.M{
			"$type": "string",
			"$ne":   "",
		},
	}

	opts := options.FindOne().SetProjection(bson.M{
		"code":         1,
		"product_name": 1,
		"brands":       1,
		"nutriments":   1,
		"serving_size": 1,
		"quantity":     1,
		"_id":          0,
	})

	var doc OffFoodDoc
	err := r.col.FindOne(ctx, filter, opts).Decode(&doc)
	if doc.Nutriments == nil {
		doc.Nutriments = map[string]any{}
	}
	if err == mongo.ErrNoDocuments {
		return nil, NotFoundError{Msg: "off food not found"}
	}
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

// Text search
func (r *RepoMongoOFF) SearchByNameOrBrand(ctx context.Context, q string, limit int, cursor string) ([]OffFoodDoc, *string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []OffFoodDoc{}, nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	var filter bson.M
	var sort bson.D
	var projection bson.M
	projection = bson.M{
		"code":           1,
		"product_name":   1,
		"brands":         1,
		"nutriments":     1,
		"serving_size":   1,
		"quantity":       1,
		"popularity_key": 1,
		"_id":            0,
	}

	if r.searchMode == "text" {
		filter = bson.M{
			"$and": []bson.M{
				{"product_name": bson.M{"$type": "string", "$ne": ""}},
				{"code": bson.M{"$type": "string", "$ne": ""}},
				{"$text": bson.M{"$search": q}},
			},
		}
		// For text search, use score first, then popularity.
		projection["score"] = bson.M{"$meta": "textScore"}
		sort = bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}, {Key: "popularity_key", Value: -1}, {Key: "code", Value: 1}}
	} else {
		pat := regexp.QuoteMeta(q)
		filter = bson.M{
			"$and": []bson.M{
				{"product_name": bson.M{"$type": "string", "$ne": ""}},
				{"code": bson.M{"$type": "string", "$ne": ""}},
				{
					"$or": []bson.M{
						{"product_name": bson.M{"$regex": pat, "$options": "i"}},
						{"brands": bson.M{"$regex": pat, "$options": "i"}},
					},
				},
			},
		}
		sort = bson.D{{Key: "popularity_key", Value: -1}, {Key: "code", Value: 1}}
		// Cursor pagination for regex (and also OK for text): cursor="<popularity>|<code>"
		if p, c, ok := parseCursor(cursor); ok {
			filter["$and"] = append(filter["$and"].([]bson.M), bson.M{
				"$or": []bson.M{
					{"popularity_key": bson.M{"$lt": p}},
					{"popularity_key": p, "code": bson.M{"$gt": c}},
				},
			})
		}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetProjection(projection).
		SetSort(sort)

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make([]OffFoodDoc, 0, limit)
	for cur.Next(ctx) {
		var d OffFoodDoc
		if err := cur.Decode(&d); err != nil {
			continue
		}
		if d.Nutriments == nil {
			d.Nutriments = map[string]any{}
		}
		out = append(out, d)
	}
	if err := cur.Err(); err != nil {
		return nil, nil, err
	}
	var next *string
	if len(out) > 0 {
		last := out[len(out)-1]
		n := makeCursor(last.Popularity, last.Code)
		next = &n
	}
	return out, next, nil
}

func parseCursor(cursor string) (pop int64, code string, ok bool) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return 0, "", false
	}
	parts := strings.SplitN(cursor, "|", 2)
	if len(parts) != 2 {
		return 0, "", false
	}
	p, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", false
	}
	code = strings.TrimSpace(parts[1])
	if code == "" {
		return 0, "", false
	}
	return p, code, true
}

func makeCursor(pop int64, code string) string {
	return strconv.FormatInt(pop, 10) + "|" + code
}
