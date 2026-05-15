package eventrepository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	eventmodel "misis_kolhoz/internal/events/model"
)

const (
	createExtensionVector = `CREATE EXTENSION IF NOT EXISTS vector`

	createEventsTable = `CREATE TABLE IF NOT EXISTS public.event_embeddings (
		id SERIAL PRIMARY KEY,
		event_date TEXT,
		holiday_info TEXT,
		category TEXT,
		about TEXT,
		food_customs TEXT,
		embedding vector(384)
	)`

	insertEventEmbedding = `INSERT INTO public.event_embeddings
		(event_date, holiday_info, category, about, food_customs, embedding)
		VALUES ($1, $2, $3, $4, $5, $6::vector)`

	truncateEventsTable = `TRUNCATE TABLE public.event_embeddings RESTART IDENTITY`

	selectAllEvents = `SELECT id, event_date, holiday_info, category, about, food_customs
		FROM public.event_embeddings
		ORDER BY event_date::date ASC, id ASC`

	selectEventsByMonth = `SELECT id, event_date, holiday_info, category, about, food_customs
		FROM public.event_embeddings
		WHERE EXTRACT(YEAR FROM event_date::date) = $1
		  AND EXTRACT(MONTH FROM event_date::date) = $2
		ORDER BY event_date::date ASC, id ASC`

	selectUpcomingEvents = `SELECT id, event_date, holiday_info, category, about, food_customs
		FROM public.event_embeddings
		WHERE event_date::date >= $1::date
		  AND event_date::date <= $2::date
		ORDER BY event_date::date ASC, id ASC
		LIMIT $3`

	selectEventsByCategory = `SELECT id, event_date, holiday_info, category, about, food_customs
		FROM public.event_embeddings
		WHERE LOWER(category) = LOWER($1)
		ORDER BY event_date::date ASC, id ASC`
)

type Repository struct {
	pgDB *pgxpool.Pool
}

func NewRepository(pgDB *pgxpool.Pool) *Repository {
	return &Repository{pgDB: pgDB}
}

func (r *Repository) InitTables(ctx context.Context) error {
	_, err := r.pgDB.Exec(ctx, createExtensionVector)
	if err != nil {
		return fmt.Errorf("eventrepository.createExtensionVector: %w", err)
	}

	_, err = r.pgDB.Exec(ctx, createEventsTable)
	if err != nil {
		return fmt.Errorf("eventrepository.createEventsTable: %w", err)
	}

	return nil
}

func (r *Repository) AddEvent(ctx context.Context, event eventmodel.EventEmbedding) error {
	vecStr := formatVectorForSQL(event.Embedding)

	_, err := r.pgDB.Exec(ctx, insertEventEmbedding,
		event.EventDate,
		event.HolidayInfo,
		event.Category,
		event.About,
		event.FoodCustoms,
		vecStr,
	)
	if err != nil {
		return fmt.Errorf("eventrepository.AddEvent: %w", err)
	}

	return nil
}

func (r *Repository) ClearEvents(ctx context.Context) error {
	_, err := r.pgDB.Exec(ctx, truncateEventsTable)
	if err != nil {
		return fmt.Errorf("eventrepository.ClearEvents: %w", err)
	}

	return nil
}

func (r *Repository) GetAllEvents(ctx context.Context) ([]eventmodel.EventEmbedding, error) {
	return r.queryEvents(ctx, selectAllEvents)
}

func (r *Repository) GetEventsByMonth(ctx context.Context, year int, month int) ([]eventmodel.EventEmbedding, error) {
	return r.queryEvents(ctx, selectEventsByMonth, year, month)
}

func (r *Repository) GetUpcomingEvents(ctx context.Context, fromDate time.Time, toDate time.Time, limit int) ([]eventmodel.EventEmbedding, error) {
	return r.queryEvents(ctx, selectUpcomingEvents, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"), limit)
}

func (r *Repository) GetEventsByCategory(ctx context.Context, category string) ([]eventmodel.EventEmbedding, error) {
	return r.queryEvents(ctx, selectEventsByCategory, strings.TrimSpace(category))
}

func (r *Repository) queryEvents(ctx context.Context, query string, args ...any) ([]eventmodel.EventEmbedding, error) {
	rows, err := r.pgDB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("eventrepository.queryEvents: %w", err)
	}
	defer rows.Close()

	result := make([]eventmodel.EventEmbedding, 0)
	for rows.Next() {
		var item eventmodel.EventEmbedding
		if err := rows.Scan(&item.ID, &item.EventDate, &item.HolidayInfo, &item.Category, &item.About, &item.FoodCustoms); err != nil {
			return nil, fmt.Errorf("eventrepository.queryEvents scan: %w", err)
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("eventrepository.queryEvents rows: %w", err)
	}

	return result, nil
}

func formatVectorForSQL(emb []float32) string {
	parts := make([]string, len(emb))
	for i, v := range emb {
		parts[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}

	return "[" + strings.Join(parts, ",") + "]"
}
