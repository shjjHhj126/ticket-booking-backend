package venue

type Venue struct {
	ID       int       `db:"id" json:"id,omitempty"`
	Name     string    `db:"name" json:"name"`
	City     string    `db:"city" json:"city"`
	Country  string    `db:"country" json:"country"`
	Sections []Section `json:"sections,omitempty"`
}

type Section struct {
	ID   int    `db:"id" json:"id,omitempty"`
	Name string `db:"name" json:"name"`
	Rows []Row  `json:"rows"`
}

type Row struct {
	ID    int    `db:"id" json:"id,omitempty"`
	Name  string `db:"name" json:"name"`
	Seats []Seat `json:"seats"`
}

type Seat struct {
	ID       int  `db:"id" json:"id,omitempty"`
	Number   int  `db:"seat_number" json:"number"` // Seat number within the row
	Price    int  `db:"price" json:"price"`
	BookedBy *int `db:"booked_by" json:"booked_by,omitempty"` // Null means unbooked, could use int+0(in db) but more vague
}
