package rest

import (
	"fmt"
	"regexp"
	"strconv"
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
func VisitorRecordFromDB(r *db.Record, loc *time.Location) (*VisitorRecord, error) {
	var err error
	res := &VisitorRecord{
		ID:                r.ID,
		MutatieID:         r.MutatieID,
		Locatie:           r.Locatie,
		Kamer:             r.Kamer,
		Bed:               r.Bed,
		Ingangsklacht:     r.Klacht,
		Specialisme:       r.Specialisme,
		Triage:            r.Triage,
		Herkomst:          r.Herkomst,
		Vervoerder:        r.Vervoer,
		Ontslagbestemming: r.Ontslagbestemming,
		OpnameAfdeling:    r.OpnameAfdeling,
		OpnameSpecialisme: r.OpnameSpecialisme,
	}
	res.Binnenkomst, err = datumTijd(r.AankomstDatum, r.AankomstTijd, loc)
	if err != nil {
		return nil, err
	}
	res.AanvangTriage, err = datumTijdRef(res.Binnenkomst, r.TriageTijd, loc)
	if err != nil {
		return nil, err
	}
	res.NaarKamer, err = datumTijdRef(res.Binnenkomst, r.BehandelTijd, loc)
	if err != nil {
		return nil, err
	}
	res.EersteContact, err = datumTijdRef(res.Binnenkomst, r.GezienTijd, loc)
	if err != nil {
		return nil, err
	}
	res.AfdelingGebeld, err = datumTijdRef(res.Binnenkomst, r.GebeldTijd, loc)
	if err != nil {
		return nil, err
	}
	res.GereedOpname, err = datumTijdRef(res.Binnenkomst, r.OpnameTijd, loc)
	if err != nil {
		return nil, err
	}
	res.Vertrek, err = datumTijdRef(res.Binnenkomst, r.VertrekTijd, loc)
	if err != nil {
		return nil, err
	}

	if res.Binnenkomst != nil && !r.Geboortedatum.IsZero() {
		leeftijd := age(res.Binnenkomst, r.Geboortedatum)
		leeftijd = leeftijd / 10
		res.Leeftijd = strconv.Itoa(leeftijd)
	}

	return res, nil
}

func age(now *time.Time, birthday time.Time) int {
	years := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		years--
	}
	return years
}

var (
	reParseTijd = regexp.MustCompile(`^(\d\d:\d\d)(:\d\d)?$`)
)

func normalizeTijd(t string) (string, error) {
	parsedTijd := reParseTijd.FindAllStringSubmatch(t, 1)
	if len(parsedTijd) != 1 || len(parsedTijd[0]) < 2 {
		return "", fmt.Errorf("unrecognized time format: %q", t)
	}
	return parsedTijd[0][1], nil
}

func datumTijd(datum, tijd string, location *time.Location) (*time.Time, error) {
	if datum == "" || tijd == "" {
		return nil, nil
	}

	nt, err := normalizeTijd(tijd)
	if err != nil {
		return nil, err
	}

	t, err := time.ParseInLocation("2006-01-02 15:04", datum+" "+nt, location)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func datumTijdRef(ref *time.Time, tijd string, location *time.Location) (*time.Time, error) {
	if ref == nil {
		return nil, nil
	}
	if tijd == "" {
		return nil, nil
	}

	nt, err := normalizeTijd(tijd)
	if err != nil {
		return nil, err
	}

	dt := fmt.Sprintf("%d-%02d-%02d %s", ref.Year(), ref.Month(), ref.Day(), nt)
	t, err := time.ParseInLocation("2006-01-02 15:04", dt, location)
	if err != nil {
		return nil, err
	}

	if ref.After(t) {
		t = t.AddDate(0, 0, 1)
	}

	return &t, err
}