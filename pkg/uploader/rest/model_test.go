package rest

import (
	"reflect"
	"testing"
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/db"
)

func tm(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	t = t.In(time.UTC)
	return &t
}

func u(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	tu := t.In(time.UTC)
	return &tu
}

func TestVisitorRecordFromDB(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Fatal(err)
	}

	for name, test := range map[string]struct {
		Given *db.Record
		Want  *VisitorRecord
	}{
		"minimal": {
			Given: &db.Record{ID: 12, MutatieID: 100, Locatie: "qqq"},
			Want:  &VisitorRecord{ID: 12, MutatieID: 100, Locatie: "qqq"},
		},
		"Aangemeld": {},
		"Binnenkomst": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
			},
			Want: &VisitorRecord{
				Binnenkomst: tm("2017-01-01T00:20:00+01:00"),
			},
		},
		"AanvangTriage": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
				TriageTijd:    "00:24",
			},
			Want: &VisitorRecord{
				Binnenkomst:   tm("2017-01-01T00:20:00+01:00"),
				AanvangTriage: tm("2017-01-01T00:24:00+01:00"),
			},
		},
		"AanvangTriage, geen aankomst": {
			Given: &db.Record{
				TriageTijd: "00:24",
			},
			Want: &VisitorRecord{},
		},
		"AanvangTriage, volgende dag": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "23:20:00",
				TriageTijd:    "00:24",
			},
			Want: &VisitorRecord{
				Binnenkomst:   tm("2017-01-01T23:20:00+01:00"),
				AanvangTriage: tm("2017-01-02T00:24:00+01:00"),
			},
		},
		"NaarKamer": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
				BehandelTijd:  "00:20",
			},
			Want: &VisitorRecord{
				Binnenkomst: tm("2017-01-01T00:20:00+01:00"),
				NaarKamer:   tm("2017-01-01T00:20:00+01:00"),
			},
		},
		"EersteContact": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
				GezienTijd:    "03:06",
			},
			Want: &VisitorRecord{
				Binnenkomst:   tm("2017-01-01T00:20:00+01:00"),
				EersteContact: tm("2017-01-01T03:06:00+01:00"),
			},
		},
		"AfdelingGebeld": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
				GebeldTijd:    "02:23",
			},
			Want: &VisitorRecord{
				Binnenkomst:    tm("2017-01-01T00:20:00+01:00"),
				AfdelingGebeld: tm("2017-01-01T02:23:00+01:00"),
			},
		},
		"GereedOpname": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
				OpnameTijd:    "04:52",
			},
			Want: &VisitorRecord{
				Binnenkomst:  tm("2017-01-01T00:20:00+01:00"),
				GereedOpname: tm("2017-01-01T04:52:00+01:00"),
			},
		},
		"Vertrek": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
				VertrekTijd:   "07:21",
			},
			Want: &VisitorRecord{
				Binnenkomst: tm("2017-01-01T00:20:00+01:00"),
				Vertrek:     tm("2017-01-01T07:21:00+01:00"),
			},
		},
		"Kamer": {
			Given: &db.Record{
				Kamer: "D12",
			},
			Want: &VisitorRecord{
				Kamer: "D12",
			},
		},
		"Bed": {
			Given: &db.Record{
				Bed: "01",
			},
			Want: &VisitorRecord{
				Bed: "01",
			},
		},
		"Ingangsklacht": {
			Given: &db.Record{
				Klacht: "kortademigheid volwassene",
			},
			Want: &VisitorRecord{
				Ingangsklacht: "kortademigheid volwassene",
			},
		},
		"Specialisme": {
			Given: &db.Record{
				Specialisme: "INT",
			},
			Want: &VisitorRecord{
				Specialisme: "INT",
			},
		},
		"Triage": {
			Given: &db.Record{
				Triage: "04",
			},
			Want: &VisitorRecord{
				Triage: "04",
			},
		},
		"Herkomst": {
			Given: &db.Record{
				Herkomst: "EIG",
			},
			Want: &VisitorRecord{
				Herkomst: "EIG",
			},
		},
		"Vervoerder": {
			Given: &db.Record{
				Vervoer: "AMB",
			},
			Want: &VisitorRecord{
				Vervoerder: "AMB",
			},
		},
		"Ontslagbestemming": {
			Given: &db.Record{
				Ontslagbestemming: "NH",
			},
			Want: &VisitorRecord{
				Ontslagbestemming: "NH",
			},
		},
		"OpnameAfdeling": {
			Given: &db.Record{
				OpnameAfdeling: "HAOA",
			},
			Want: &VisitorRecord{
				OpnameAfdeling: "HAOA",
			},
		},
		"OpnameSpecialisme": {
			Given: &db.Record{
				OpnameSpecialisme: "INT",
			},
			Want: &VisitorRecord{
				OpnameSpecialisme: "INT",
			},
		},
		"Leeftijd": {
			Given: &db.Record{
				AankomstDatum: "2017-01-01",
				AankomstTijd:  "00:20:00",
				Geboortedatum: time.Date(1977, time.July, 24, 12, 0, 0, 0, time.UTC),
			},
			Want: &VisitorRecord{
				Binnenkomst: tm("2017-01-01T00:20:00+01:00"),
				Leeftijd:    "3",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if test.Want == nil {
				t.Skip("not defined")
			}

			got, err := VisitorRecordFromDB(test.Given, loc)
			if err != nil {
				t.Fatal(err)
			}

			got.Binnenkomst = u(got.Binnenkomst)
			got.AanvangTriage = u(got.AanvangTriage)
			got.NaarKamer = u(got.NaarKamer)
			got.EersteContact = u(got.EersteContact)
			got.AfdelingGebeld = u(got.AfdelingGebeld)
			got.Vertrek = u(got.Vertrek)
			got.Aangemeld = u(got.Aangemeld)
			got.GereedOpname = u(got.GereedOpname)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("VisitorRecordFromDB(\n\t%#v) == \n\t%v, got \n\t%v", test.Given, test.Want, got)
			}
		})
	}
}
