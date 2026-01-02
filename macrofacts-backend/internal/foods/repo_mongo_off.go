package foods

import (
	"context"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RepoMongoOFF struct {
	col        *mongo.Collection
	searchMode string // "keywords" | "regex" | "text"
}

type OffFoodDoc struct {
	ID any `bson:"_id"`

	// Optional in some dumps, but often empty because the barcode is stored in _id.
	Code string `bson:"code"`

	ProductName string `bson:"product_name"`
	Brands      string `bson:"brands"`

	ServingSize string `bson:"serving_size"`
	Quantity    string `bson:"quantity"`

	Popularity  int64 `bson:"popularity_key"`
	UniqueScans int64 `bson:"unique_scans_n"`

	Keywords []string `bson:"_keywords"`

	Nutriments map[string]any `bson:"nutriments"`
}

func (d OffFoodDoc) IDString() string {
	switch v := d.ID.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func NewRepoMongoOFF(client *mongo.Client, dbName, collection, searchMode string) *RepoMongoOFF {
	searchMode = strings.TrimSpace(strings.ToLower(searchMode))
	if searchMode == "" {
		// Default to keyword search because OFF dumps commonly include `_keywords`
		// and many datasets are already near the index limit (so adding new indexes fails).
		searchMode = "keywords"
	}
	return &RepoMongoOFF{
		col:        client.Database(dbName).Collection(collection),
		searchMode: searchMode,
	}
}

// Only needed if you explicitly use the "text" mode.
// In "keywords" mode, avoid attempting to create extra indexes (many dumps hit index limits).
func (r *RepoMongoOFF) EnsureTextIndex(ctx context.Context) error {
	if r.searchMode != "text" {
		return nil
	}

	models := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "product_name", Value: "text"},
				{Key: "brands", Value: "text"},
			},
			Options: options.Index().SetName("idx_name_brand_text").SetDefaultLanguage("none"),
		},
	}

	_, err := r.col.Indexes().CreateMany(ctx, models)
	return err
}

func (r *RepoMongoOFF) ByBarcode(ctx context.Context, code string) (*OffFoodDoc, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}

	// 1) Try "code" first (works on standard OFF dumps)
	filter := bson.M{
		"code":         code,
		"product_name": bson.M{"$type": "string", "$ne": ""},
	}

	projection := bson.M{
		"_id":            1, // safe now because ID is `any`
		"code":           1,
		"product_name":   1,
		"brands":         1,
		"nutriments":     1,
		"serving_size":   1,
		"quantity":       1,
		"unique_scans_n": 1,
	}

	var out OffFoodDoc
	err := r.col.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&out)
	if err == nil {
		return &out, nil
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 2) Fallback: some dumps store barcode in _id as a STRING.
	// Querying _id is okay; decoding is safe because ID is `any`.
	filter = bson.M{
		"_id":          code,
		"product_name": bson.M{"$type": "string", "$ne": ""},
	}

	err = r.col.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&out)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RepoMongoOFF) SearchByNameOrBrand(
	ctx context.Context,
	q string,
	limit int,
	cursor string,
) ([]OffFoodDoc, *string, error) {

	tokens := keywordize(q)
	if len(tokens) == 0 {
		return []OffFoodDoc{}, nil, nil
	}

	filter := bson.M{
		"_keywords": bson.M{"$in": tokens},
		"product_name": bson.M{
			"$type": "string",
			"$ne":   "",
		},
		"code": bson.M{
			"$type": "string",
			"$ne":   "",
		},
	}

	// Cursor for keywords: "<unique_scans_n>|<code>"
	if cursor != "" {
		if u, code, ok := parseKeywordCursor(cursor); ok {
			filter = bson.M{
				"$and": []bson.M{
					filter,
					{
						"$or": []bson.M{
							{"unique_scans_n": bson.M{"$lt": u}},
							{
								"unique_scans_n": u,
								"code":           bson.M{"$gt": code},
							},
						},
					},
				},
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{
			{Key: "unique_scans_n", Value: -1},
			{Key: "code", Value: 1},
		}).
		SetProjection(bson.M{
			"_id":            0, // ⬅️ CRITICAL FIX
			"code":           1,
			"product_name":   1,
			"brands":         1,
			"nutriments":     1,
			"serving_size":   1,
			"quantity":       1,
			"unique_scans_n": 1,
		})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, nil, err
	}
	defer cur.Close(ctx)

	var out []OffFoodDoc
	if err := cur.All(ctx, &out); err != nil {
		return nil, nil, err
	}

	var next *string
	if len(out) > 0 {
		last := out[len(out)-1]
		c := makeKeywordCursor(last.UniqueScans, last.Code)
		next = &c
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
	return p, parts[1], true
}

func makeCursor(pop int64, code string) string {
	return strconv.FormatInt(pop, 10) + "|" + code
}

// keywordize turns a free-form query into OFF `_keywords` tokens (lowercase words).
func keywordize(q string) []string {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}

	// Replace anything that's not letter/number with spaces.
	var b strings.Builder
	b.Grow(len(q))
	for _, r := range q {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}

	raw := strings.Fields(b.String())
	if len(raw) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if len(t) < 2 {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
		if len(out) >= 6 {
			break
		}
	}
	return out
}

func parseKeywordCursor(cursor string) (unique int64, id string, ok bool) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return 0, "", false
	}
	parts := strings.SplitN(cursor, "|", 2)
	if len(parts) != 2 {
		return 0, "", false
	}
	u, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", false
	}
	id = strings.TrimSpace(parts[1])
	if id == "" {
		return 0, "", false
	}
	return u, id, true
}

func makeKeywordCursor(unique int64, id string) string {
	return strconv.FormatInt(unique, 10) + "|" + id
}
