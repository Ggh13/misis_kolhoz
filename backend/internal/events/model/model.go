package eventmodel

type EventEmbedding struct {
	ID          int       `json:"id"`
	EventDate   string    `json:"event_date"`
	HolidayInfo string    `json:"holiday_info"`
	Category    string    `json:"category"`
	About       string    `json:"about"`
	FoodCustoms string    `json:"food_customs"`
	Embedding   []float32 `json:"-"`
}
