package eventservice

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	eventmodel "misis_kolhoz/internal/events/model"
)

type Repository interface {
	InitTables(ctx context.Context) error
	ClearEvents(ctx context.Context) error
	AddEvent(ctx context.Context, event eventmodel.EventEmbedding) error
	GetAllEvents(ctx context.Context) ([]eventmodel.EventEmbedding, error)
	GetEventsByMonth(ctx context.Context, year int, month int) ([]eventmodel.EventEmbedding, error)
	GetUpcomingEvents(ctx context.Context, fromDate time.Time, toDate time.Time, limit int) ([]eventmodel.EventEmbedding, error)
	GetEventsByCategory(ctx context.Context, category string) ([]eventmodel.EventEmbedding, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LoadDataFromExcel(ctx context.Context, filePath string) error {
	if err := s.repo.InitTables(ctx); err != nil {
		return fmt.Errorf("eventservice.LoadDataFromExcel init tables: %w", err)
	}

	if err := s.repo.ClearEvents(ctx); err != nil {
		return fmt.Errorf("eventservice.LoadDataFromExcel clear events table: %w", err)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("eventservice.LoadDataFromExcel open file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetMap()
	processedAnySheet := false
	for _, sheetName := range sheets {
		processed, err := s.processSheet(ctx, f, sheetName)
		if err != nil {
			return fmt.Errorf("eventservice.LoadDataFromExcel process sheet: %w", err)
		}
		processedAnySheet = processedAnySheet || processed
	}

	if !processedAnySheet {
		return fmt.Errorf("eventservice.LoadDataFromExcel: no sheets with required headers found")
	}

	return nil
}

func (s *Service) processSheet(ctx context.Context, f *excelize.File, sheetName string) (bool, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return false, fmt.Errorf("eventservice.processSheet get rows: %w", err)
	}

	if len(rows) < 2 {
		return false, nil
	}

	headerMap := make(map[string]int)
	for i, h := range rows[0] {
		headerMap[normalizeHeader(h)] = i
	}

	dateIdx, ok := headerMap[normalizeHeader("Дата")]
	if !ok {
		return false, nil
	}
	holidayIdx, ok := headerMap[normalizeHeader("Праздник / инфоповод")]
	if !ok {
		return false, nil
	}
	categoryIdx, ok := headerMap[normalizeHeader("Категория")]
	if !ok {
		return false, nil
	}
	aboutIdx, ok := headerMap[normalizeHeader("О чём он")]
	if !ok {
		return false, nil
	}
	foodIdx, ok := headerMap[normalizeHeader("Еда / обычаи")]
	if !ok {
		return false, nil
	}

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]

		event := eventmodel.EventEmbedding{Embedding: make([]float32, 384)}
		if dateIdx < len(row) {
			event.EventDate = strings.TrimSpace(row[dateIdx])
		}
		if holidayIdx < len(row) {
			event.HolidayInfo = strings.TrimSpace(row[holidayIdx])
		}
		if categoryIdx < len(row) {
			event.Category = strings.TrimSpace(row[categoryIdx])
		}
		if aboutIdx < len(row) {
			event.About = strings.TrimSpace(row[aboutIdx])
		}
		if foodIdx < len(row) {
			event.FoodCustoms = strings.TrimSpace(row[foodIdx])
		}

		if event.EventDate == "" && event.HolidayInfo == "" && event.Category == "" && event.About == "" && event.FoodCustoms == "" {
			continue
		}
		if event.HolidayInfo == "" && event.About == "" {
			continue
		}

		parsedDate, err := parseEventDate(event.EventDate)
		if err != nil {
			continue
		}
		event.EventDate = parsedDate.Format("2006-01-02")

		if err := s.repo.AddEvent(ctx, event); err != nil {
			return false, fmt.Errorf("eventservice.processSheet add event row %d: %w", rowIdx+1, err)
		}
	}

	return true, nil
}

func normalizeHeader(h string) string {
	replacer := strings.NewReplacer("ё", "е", "Ё", "Е", "\u00a0", " ")
	h = replacer.Replace(h)
	h = strings.TrimSpace(h)
	h = strings.ToLower(h)
	return strings.Join(strings.Fields(h), " ")
}

func (s *Service) GetAllEvents(ctx context.Context) ([]eventmodel.EventEmbedding, error) {
	return s.repo.GetAllEvents(ctx)
}

func (s *Service) GetEventsByMonth(ctx context.Context, year int, month int) ([]eventmodel.EventEmbedding, error) {
	if year < 1900 || year > 3000 {
		return nil, fmt.Errorf("year must be in range 1900..3000")
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month must be in range 1..12")
	}

	return s.repo.GetEventsByMonth(ctx, year, month)
}

func (s *Service) GetUpcomingEvents(ctx context.Context, days int, limit int) ([]eventmodel.EventEmbedding, error) {
	if days < 0 {
		return nil, fmt.Errorf("days must be non-negative")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	from := time.Now().Truncate(24 * time.Hour)
	to := from.AddDate(0, 0, days)
	return s.repo.GetUpcomingEvents(ctx, from, to, limit)
}

func (s *Service) GetEventsByCategory(ctx context.Context, category string) ([]eventmodel.EventEmbedding, error) {
	if strings.TrimSpace(category) == "" {
		return nil, fmt.Errorf("category is required")
	}

	return s.repo.GetEventsByCategory(ctx, category)
}

func parseEventDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	raw = normalizeDateValue(raw)

	layouts := []string{
		"2006-01-02",
		"02.01.2006",
		"02/01/2006",
		"2006/01/02",
		"2.1.2006",
		"2/1/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}

	if num, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64); err == nil {
		if t, err := excelize.ExcelDateToTime(num, false); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format: %q", raw)
}

func normalizeDateValue(raw string) string {
	raw = strings.TrimSpace(raw)

	rangeSeparators := []string{"–", "—", "-"}
	for _, sep := range rangeSeparators {
		if strings.Contains(raw, sep) {
			parts := strings.Split(raw, sep)
			if len(parts) > 0 {
				raw = strings.TrimSpace(parts[0])
			}
			break
		}
	}

	dotParts := strings.Split(raw, ".")
	if len(dotParts) == 2 {
		day := strings.TrimSpace(dotParts[0])
		month := strings.TrimSpace(dotParts[1])
		if day != "" && month != "" {
			raw = fmt.Sprintf("%s.%s.%d", day, month, time.Now().Year())
		}
	}

	slashParts := strings.Split(raw, "/")
	if len(slashParts) == 2 {
		day := strings.TrimSpace(slashParts[0])
		month := strings.TrimSpace(slashParts[1])
		if day != "" && month != "" {
			raw = fmt.Sprintf("%s/%s/%d", day, month, time.Now().Year())
		}
	}

	return raw
}
