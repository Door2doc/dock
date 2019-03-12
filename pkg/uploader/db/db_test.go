package db

import (
	"context"
	"database/sql"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
)

const (
	TestDSN = "postgres://pguser:pwd@localhost:5436/pgdb?sslmode=disable"
)

var (
	_ pq.Driver
)

func TestExecuteQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for name, test := range map[string]struct {
		Query string
		Want  []Record
		Err   error
	}{
		"correct": {
			Query: `select * from correct where id = 1`,
			Want: []Record{{
				ID:                328996,
				MutatieID:         1091568,
				Locatie:           "A",
				AanmaakDatum:      "",
				AanmaakTijd:       "",
				AankomstDatum:     "2017-07-13 00:00:00.000",
				AankomstTijd:      "23:18",
				TriageDatum:       "",
				TriageTijd:        "",
				BehandelTijd:      "23:18",
				GezienTijd:        "02:40",
				GebeldTijd:        "",
				OpnameTijd:        "02:06",
				VertrekTijd:       "04:34",
				Kamer:             "",
				Bed:               "",
				Klacht:            "Pneumonie",
				Specialisme:       "04",
				Triage:            "",
				Vervoer:           "2",
				Bestemming:        "A",
				Geboortedatum:     time.Date(1977, time.July, 24, 12, 0, 0, 0, time.UTC),
				OpnameAfdeling:    "",
				OpnameSpecialisme: "",
			}},
		},
		"check all columns": {
			Query: `select * from correct where id = 2`,
			Want: []Record{{
				ID:                1,
				MutatieID:         2,
				Locatie:           "a",
				AanmaakDatum:      "b",
				AanmaakTijd:       "c",
				AankomstDatum:     "d",
				AankomstTijd:      "e",
				TriageDatum:       "f",
				TriageTijd:        "g",
				BehandelTijd:      "h",
				GezienTijd:        "i",
				GebeldTijd:        "j",
				OpnameTijd:        "k",
				VertrekTijd:       "l",
				Kamer:             "m",
				Bed:               "n",
				Klacht:            "o",
				Specialisme:       "p",
				Triage:            "q",
				Vervoer:           "r",
				Bestemming:        "s",
				Geboortedatum:     time.Time{},
				OpnameAfdeling:    "u",
				OpnameSpecialisme: "v",
			}},
		},
		"missing columns": {
			Query: `select null as hello`,
			Err:   &SelectionError{Missing: []string{"SEHID", "SEHMUTID", "LOCATIECOD", "AANMAAKDAT", "AANMAAKTIJD", "AANKSDATUM", "AANKSTIJD", "TRIADATUM", "TRIAGETIJD", "ARTSBHTIJD", "PATGEZT", "GEBELD", "INSCHRTIJD", "ARBEHETIJD", "BEHKAMERCO", "BEDNR", "KLACHT", "SPECIALISM", "TRIANIVCOD", "VERVOERTYP", "BESTEMMING", "GEBDAT", "OPNAMEAFD", "OPNAMESPEC"}},
		},
		"some missing columns": {
			Query: `select 1 as sehid, 2 as sehmutid, '' as locatiecod, '' as aanmaakdat, null as opnameafd, null as opnamespec`,
			Err:   &SelectionError{Missing: []string{"AANMAAKTIJD", "AANKSDATUM", "AANKSTIJD", "TRIADATUM", "TRIAGETIJD", "ARTSBHTIJD", "PATGEZT", "GEBELD", "INSCHRTIJD", "ARBEHETIJD", "BEHKAMERCO", "BEDNR", "KLACHT", "SPECIALISM", "TRIANIVCOD", "VERVOERTYP", "BESTEMMING", "GEBDAT"}},
		},
		"duplicate columns": {
			Query: `select null as hello, null as hello`,
			Err:   &QueryError{Cause: "query contains duplicate column names"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			tx, cancel := setup(ctx, t)
			defer cancel()

			got, err := ExecuteQuery(ctx, tx, test.Query)
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("ExecuteQuery() == \n\t%#v, got \n\t%#v", test.Want, got)
			}
			if !reflect.DeepEqual(err, test.Err) {
				t.Errorf("ExecuteQuery() == \n\t_, %#v; got \n\t_, %#v", test.Err, err)
			}
		})
	}
}

func TestExecuteQueryPermutations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, cancel := setup(ctx, t)
	defer cancel()

	parts := []string{
		"1 as sehid",
		"2 as sehmutid",
		"'a' as locatiecod",
		"'b' as aanmaakdat",
		"'c' as aanmaaktijd",
		"'d' as aanksdatum",
		"'e' as aankstijd",
		"'f' as triadatum",
		"'g' as triagetijd",
		"'h' as artsbhtijd",
		"'i' as patgezt",
		"'j' as gebeld",
		"'k' as inschrtijd",
		"'l' as arbehetijd",
		"'m' as behkamerco",
		"'n' as bednr",
		"'o' as klacht",
		"'p' as specialism",
		"'q' as trianivcod",
		"'r' as vervoertyp",
		"'s' as bestemming",
		"NULL as gebdat",
		"'u' as opnameafd",
		"'v' as opnamespec",
		"'x' as ignoreme_1",
		"'x' as ignoreme_2",
		"'x' as ignoreme_3",
		"2 as ignoreme_4",
		"3 as ignoreme_5",
		"4 as ignoreme_6",
		"'x' as ignoreme_7",
		"'x' as ignoreme_8",
		"'x' as ignoreme_9",
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(len(parts), func(i, j int) {
		parts[i], parts[j] = parts[j], parts[i]
	})
	query := "select " + strings.Join(parts, ",")
	got, err := ExecuteQuery(ctx, tx, query)
	if err != nil {
		t.Fatal(err)
	}

	want := []Record{{
		ID:                1,
		MutatieID:         2,
		Locatie:           "a",
		AanmaakDatum:      "b",
		AanmaakTijd:       "c",
		AankomstDatum:     "d",
		AankomstTijd:      "e",
		TriageDatum:       "f",
		TriageTijd:        "g",
		BehandelTijd:      "h",
		GezienTijd:        "i",
		GebeldTijd:        "j",
		OpnameTijd:        "k",
		VertrekTijd:       "l",
		Kamer:             "m",
		Bed:               "n",
		Klacht:            "o",
		Specialisme:       "p",
		Triage:            "q",
		Vervoer:           "r",
		Bestemming:        "s",
		Geboortedatum:     time.Time{},
		OpnameAfdeling:    "u",
		OpnameSpecialisme: "v",
	}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExecuteQuery() == \n\t%#v, got \n\t%#v", want, got)
	}
}

func setup(ctx context.Context, t *testing.T) (*sql.Tx, context.CancelFunc) {
	db, err := sql.Open("postgres", TestDSN)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	return tx, func() {
		if err := tx.Commit(); err != nil {
			t.Error(err)
		}
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}
}
