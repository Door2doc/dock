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

func u(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(time.UTC)
}

func TestExecuteQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for name, test := range map[string]struct {
		Query string
		Want  []VisitorRecord
		Err   error
	}{
		"correct": {
			Query: `select * from correct where id = 1`,
			Want: []VisitorRecord{{
				ID:                 328996,
				MutatieID:          1091568,
				Locatie:            "A",
				Aangemaakt:         time.Date(2017, time.July, 13, 13, 0, 0, 0, time.UTC),
				BinnenkomstDatum:   "2017-07-13",
				BinnenkomstTijd:    "23:18",
				AanvangTriageTijd:  "",
				NaarKamerTijd:      "23:18",
				EersteContactTijd:  "02:40",
				AfdelingGebeldTijd: "",
				GereedOpnameTijd:   "02:06",
				VertrekTijd:        "04:34",
				Kamer:              "",
				Bed:                "",
				Ingangsklacht:      "Pneumonie",
				Specialisme:        "04",
				Triage:             "",
				Vervoerder:         "2",
				Geboortedatum:      time.Date(1977, time.July, 24, 12, 0, 0, 0, time.UTC),
				OpnameAfdeling:     "",
				OpnameSpecialisme:  "",
			}},
		},
		"check all columns": {
			Query: `select * from correct where id = 2`,
			Want: []VisitorRecord{{
				ID:                 1,
				MutatieID:          2,
				Aangemaakt:         time.Date(2018, time.July, 4, 12, 4, 0, 0, time.UTC),
				Locatie:            "locatie",
				BinnenkomstDatum:   "binnenkomstdatum",
				BinnenkomstTijd:    "binnenkomsttijd",
				AanvangTriageTijd:  "aanvangtriagetijd",
				NaarKamerTijd:      "naarkamertijd",
				EersteContactTijd:  "eerstecontacttijd",
				AfdelingGebeldTijd: "afdelinggebeldtijd",
				GereedOpnameTijd:   "gereedopnametijd",
				VertrekTijd:        "vertrektijd",
				Kamer:              "kamer",
				Bed:                "bed",
				Ingangsklacht:      "ingangsklacht",
				Specialisme:        "specialisme",
				Triage:             "triage",
				Vervoerder:         "vervoerder",
				Geboortedatum:      time.Date(1977, time.July, 24, 12, 0, 0, 0, time.UTC),
				OpnameAfdeling:     "opnameafdeling",
				OpnameSpecialisme:  "opnamespecialisme",
				Herkomst:           "herkomst",
				Ontslagbestemming:  "ontslagbestemming",
			}},
		},
		"missing columns": {
			Query: `select null as hello`,
			Err: &SelectionError{Missing: []string{
				ColID.Name,
				ColMutatieID.Name,
				ColLocatie.Name,
				ColAangemaakt.Name,
				ColBinnenkomstDatum.Name,
				ColBinnenkomstTijd.Name,
				ColAanvangTriageTijd.Name,
				ColNaarKamerTijd.Name,
				ColEersteContactTijd.Name,
				ColAfdelingGebeldTijd.Name,
				ColGereedOpnameTijd.Name,
				ColVertrekTijd.Name,
				ColKamer.Name,
				ColBed.Name,
				ColIngangsklacht.Name,
				ColSpecialisme.Name,
				ColTriage.Name,
				ColVervoerder.Name,
				ColGeboortedatum.Name,
				ColOpnameAfdeling.Name,
				ColOpnameSpecialisme.Name,
				ColHerkomst.Name,
				ColOntslagbestemming.Name,
			}},
		},
		"duplicate columns": {
			Query: `select null as hello, null as hello`,
			Err:   &QueryError{Cause: "query contains duplicate column names"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			tx, cancel := setup(ctx, t)
			defer cancel()

			got, err := ExecuteVisitorQuery(ctx, tx, test.Query)

			for i := range got {
				got[i].Aangemaakt = u(got[i].Aangemaakt)
				got[i].Geboortedatum = u(got[i].Geboortedatum)
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("ExecuteVisitorQuery() == \n\t%v, got \n\t%v", test.Want, got)
			}
			if !reflect.DeepEqual(err, test.Err) {
				t.Errorf("ExecuteVisitorQuery() == \n\t_, %#v; got \n\t_, %#v", test.Err, err)
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
		"1 AS sehid",
		"2 AS sehmutid",
		"'locatie' AS locatie",
		"NULL AS aangemaakt",
		"'binnenkomstdatum' AS binnenkomstdatum",
		"'binnenkomsttijd' AS binnenkomsttijd",
		"'aanvangtriagetijd' AS aanvangtriagetijd",
		"'naarkamertijd' AS naarkamertijd",
		"'eerstecontacttijd' AS eerstecontacttijd",
		"'afdelinggebeldtijd' AS afdelinggebeldtijd",
		"'gereedopnametijd' AS gereedopnametijd",
		"'vertrektijd' AS vertrektijd",
		"'kamer' AS kamer",
		"'bed' AS bed",
		"'ingangsklacht' AS ingangsklacht",
		"'specialisme' AS specialisme",
		"'triage' AS triage",
		"'vervoerder' AS vervoerder",
		"'bestemming' AS bestemming",
		"NULL AS geboortedatum",
		"'opnameafdeling' AS opnameafdeling",
		"'opnamespecialisme' AS opnamespecialisme",
		"'herkomst' AS herkomst",
		"'ontslagbestemming' AS ontslagbestemming",
		"'x' AS ignoreme_1",
		"'x' AS ignoreme_2",
		"'x' AS ignoreme_3",
		"2 AS ignoreme_4",
		"3 AS ignoreme_5",
		"4 AS ignoreme_6",
		"'x' AS ignoreme_7",
		"'x' AS ignoreme_8",
		"'x' AS ignoreme_9",
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(len(parts), func(i, j int) {
		parts[i], parts[j] = parts[j], parts[i]
	})
	query := "select " + strings.Join(parts, ",")
	got, err := ExecuteVisitorQuery(ctx, tx, query)
	if err != nil {
		t.Fatal(err)
	}

	want := []VisitorRecord{{
		ID:                 1,
		MutatieID:          2,
		Locatie:            "locatie",
		BinnenkomstDatum:   "binnenkomstdatum",
		BinnenkomstTijd:    "binnenkomsttijd",
		AanvangTriageTijd:  "aanvangtriagetijd",
		NaarKamerTijd:      "naarkamertijd",
		EersteContactTijd:  "eerstecontacttijd",
		AfdelingGebeldTijd: "afdelinggebeldtijd",
		GereedOpnameTijd:   "gereedopnametijd",
		VertrekTijd:        "vertrektijd",
		Kamer:              "kamer",
		Bed:                "bed",
		Ingangsklacht:      "ingangsklacht",
		Specialisme:        "specialisme",
		Triage:             "triage",
		Vervoerder:         "vervoerder",
		Geboortedatum:      time.Time{},
		OpnameAfdeling:     "opnameafdeling",
		OpnameSpecialisme:  "opnamespecialisme",
		Herkomst:           "herkomst",
		Ontslagbestemming:  "ontslagbestemming",
	}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExecuteVisitorQuery() == \n\t%#v, got \n\t%#v", want, got)
	}
}

func setup(ctx context.Context, t *testing.T) (*sql.Tx, context.CancelFunc) {
	if testing.Short() {
		t.Skip("uses database")
	}

	db, err := sql.Open("postgres", TestDSN)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}
}
