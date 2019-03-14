package rest

import (
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/db"
)

// VisitorRecord defines a single mutation on a visit.
type VisitorRecord struct {
	// Required fields
	ID        int    `json:"bezoek_id"`
	MutatieID int    `json:"mutatie_id"`
	Locatie   string `json:"code_locatie"`

	Aangemeld         *time.Time `json:"dt_aangemeld,omitempty"`
	Binnenkomst       *time.Time `json:"dt_binnenkomst,omitempty"`
	AanvangTriage     *time.Time `json:"dt_aanvang_triage,omitempty"`
	NaarKamer         *time.Time `json:"dt_naar_behandelkamer,omitempty"`
	EersteContact     *time.Time `json:"dt_eerste_contact_arts,omitempty"`
	AfdelingGebeld    *time.Time `json:"dt_afdeling_gebeld,omitempty"`
	GereedOpname      *time.Time `json:"dt_gereed_opname,omitempty"`
	Vertrek           *time.Time `json:"dt_vertrek,omitempty"`
	Kamer             string     `json:"behandelkamer,omitempty"`
	Bed               string     `json:"bed,omitempty"`
	Ingangsklacht     string     `json:"code_ingangsklacht,omitempty"`
	Specialisme       string     `json:"code_specialisme,omitempty"`
	Triage            string     `json:"code_triage,omitempty"`
	Herkomst          string     `json:"code_herkomst,omitempty"`
	Vervoerder        string     `json:"code_vervoerder,omitempty"`
	Ontslagbestemming string     `json:"code_ontslagbestemming,omitempty"`
	OpnameAfdeling    string     `json:"code_opnameafdeling,omitempty"`
	OpnameSpecialisme string     `json:"code_opnamespecialisme,omitempty"`
	Leeftijd          string     `json:"cat_leeftijd,omitempty"`
}

// VisitorRecordFromDB converts a database record to a visitor record.
func VisitorRecordFromDB(*db.Record) *VisitorRecord {
	return nil
}
