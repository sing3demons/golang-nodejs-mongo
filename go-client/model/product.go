package model

import "time"

type Products struct {
	Type            string    `json:"@type,omitempty" bson:"@type,omitempty"`
	ID              string    `json:"_id,omitempty" bson:"_id,omitempty"`
	Name            string    `json:"name,omitempty" bson:"name,omitempty"`
	Href            string    `json:"href,omitempty" bson:"href,omitempty"`
	LifecycleStatus string    `json:"lifecycleStatus,omitempty" bson:"lifecycleStatus,omitempty"`
	Version         string    `json:"version,omitempty" bson:"version,omitempty"`
	LastUpdate      time.Time `json:"lastUpdate,omitempty" bson:"lastUpdate,omitempty"`
	ValidFor        *ValidFor `json:"validFor,omitempty" bson:"validFor,omitempty"`

	ProductPrice *ProductPrice `json:"price,omitempty" bson:"productPrice,omitempty"`
}

type ProductPrice struct {
	Name  string  `json:"name,omitempty" bson:"name,omitempty"`
	Value float64 `json:"value,omitempty" bson:"value,omitempty"`
	Unit  string  `json:"unit,omitempty" bson:"unit,omitempty"`
}

type ValidFor struct {
	StartDateTime time.Time `json:"startDateTime,omitempty" bson:"startDateTime,omitempty"`
	EndDateTime   time.Time `json:"endDateTime,omitempty" bson:"endDateTime,omitempty"`
}
