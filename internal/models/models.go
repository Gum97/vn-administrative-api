package models

import "time"

// Province represents the payload from /pcotinh
type Province struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"tentinh" db:"name"`
	Code      int       `json:"mahc" db:"code"` // Note: API sometimes uses int for code in list
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AdminUnit represents the payload from /ptracuu (Wards)
type AdminUnit struct {
	ID            int       `json:"id" db:"id"`
	ProvinceID    int       `json:"matinh" db:"province_id"`
	Name          string    `json:"tenhc" db:"name"`
	Level         string    `json:"loai" db:"level"`
	Code          string    `json:"ma" db:"code"`
	PreMergerDesc string    `json:"truocsapnhap" db:"pre_merger_desc"`
	Lat           float64   `json:"vido" db:"lat"`
	Long          float64   `json:"kinhdo" db:"long"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
