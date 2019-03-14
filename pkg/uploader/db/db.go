package db

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
)

// ExecuteVisitorQuery tries to execute the visitor query and marshal the result into records.
func ExecuteVisitorQuery(ctx context.Context, tx *sql.Tx, query string) ([]Record, error) {
	// execute query
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, &QueryError{err.Error()}
	}
	defer func() {
		if err := rows.Close(); err != nil {
			dlog.Error("While closing result set: %v", err)
		}
	}()

	// determine column names
	names, err := rows.Columns()
	if err != nil {
		return nil, &QueryError{err.Error()}
	}

	col2index, err := checkColumnNames(names, columns)
	if err != nil {
		return nil, err
	}

	// map result set to records
	var res []Record

	for rows.Next() {
		var rec Record
		err := mapRow(rows, &rec, names, col2index)
		if err != nil {
			return nil, &QueryError{err.Error()}
		}

		target := make([]interface{}, len(names))
		for i := range target {
			target[i] = new(sql.RawBytes)
		}
		if err := rows.Scan(target...); err != nil {
			return nil, &QueryError{err.Error()}
		}

		res = append(res, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, &QueryError{err.Error()}
	}

	return res, nil
}

func mapRow(rows *sql.Rows, rec *Record, allColumns []string, col2index map[string]int) error {
	target := make([]interface{}, len(allColumns))
	for i := range target {
		target[i] = new(sql.RawBytes)
	}

	var (
		id                int
		mutatieID         int
		locatie           sql.NullString
		aanmaakDatum      sql.NullString
		aanmaakTijd       sql.NullString
		aankomstDatum     sql.NullString
		aankomstTijd      sql.NullString
		triageDatum       sql.NullString
		triageTijd        sql.NullString
		behandelTijd      sql.NullString
		gezienTijd        sql.NullString
		gebeldTijd        sql.NullString
		opnameTijd        sql.NullString
		vertrekTijd       sql.NullString
		kamer             sql.NullString
		bed               sql.NullString
		klacht            sql.NullString
		specialisme       sql.NullString
		triage            sql.NullString
		vervoer           sql.NullString
		bestemming        sql.NullString
		geboortedatum     *time.Time
		opnameAfdeling    sql.NullString
		opnameSpecialisme sql.NullString
	)

	target[col2index[ColID]] = &id
	target[col2index[ColMutatieID]] = &mutatieID
	target[col2index[ColLocatie]] = &locatie
	target[col2index[ColAanmaakDatum]] = &aanmaakDatum
	target[col2index[ColAanmaakTijd]] = &aanmaakTijd
	target[col2index[ColAankomstDatum]] = &aankomstDatum
	target[col2index[ColAankomstTijd]] = &aankomstTijd
	target[col2index[ColTriageDatum]] = &triageDatum
	target[col2index[ColTriageTijd]] = &triageTijd
	target[col2index[ColBehandelTijd]] = &behandelTijd
	target[col2index[ColGezienTijd]] = &gezienTijd
	target[col2index[ColGebeldTijd]] = &gebeldTijd
	target[col2index[ColOpnameTijd]] = &opnameTijd
	target[col2index[ColVertrekTijd]] = &vertrekTijd
	target[col2index[ColKamer]] = &kamer
	target[col2index[ColBed]] = &bed
	target[col2index[ColKlacht]] = &klacht
	target[col2index[ColSpecialisme]] = &specialisme
	target[col2index[ColTriage]] = &triage
	target[col2index[ColVervoer]] = &vervoer
	target[col2index[ColBestemming]] = &bestemming
	target[col2index[ColGeboortedatum]] = &geboortedatum
	target[col2index[ColOpnameAfdeling]] = &opnameAfdeling
	target[col2index[ColOpnameSpecialisme]] = &opnameSpecialisme

	if err := rows.Scan(target...); err != nil {
		return err
	}

	var geb time.Time
	if geboortedatum != nil {
		geb = geboortedatum.UTC()
	}

	*rec = Record{
		ID:                id,
		MutatieID:         mutatieID,
		Locatie:           locatie.String,
		AanmaakDatum:      aanmaakDatum.String,
		AanmaakTijd:       aanmaakTijd.String,
		AankomstDatum:     aankomstDatum.String,
		AankomstTijd:      aankomstTijd.String,
		TriageDatum:       triageDatum.String,
		TriageTijd:        triageTijd.String,
		BehandelTijd:      behandelTijd.String,
		GezienTijd:        gezienTijd.String,
		GebeldTijd:        gebeldTijd.String,
		OpnameTijd:        opnameTijd.String,
		VertrekTijd:       vertrekTijd.String,
		Kamer:             kamer.String,
		Bed:               bed.String,
		Klacht:            klacht.String,
		Specialisme:       specialisme.String,
		Triage:            triage.String,
		Vervoer:           vervoer.String,
		Bestemming:        bestemming.String,
		Geboortedatum:     geb,
		OpnameAfdeling:    opnameAfdeling.String,
		OpnameSpecialisme: opnameSpecialisme.String,
	}

	return nil
}

func checkColumnNames(got, want []string) (map[string]int, error) {
	got2pos := make(map[string]int)
	for i, s := range got {
		got2pos[strings.ToLower(s)] = i
	}

	if len(got) != len(got2pos) {
		return nil, &QueryError{"query contains duplicate column names"}
	}

	var (
		missing  []string
		want2pos = make(map[string]int)
	)

	for _, w := range want {
		idx, ok := got2pos[strings.ToLower(w)]
		if ok {
			want2pos[w] = idx
		} else {
			missing = append(missing, w)
		}
	}

	if len(missing) != 0 {
		return nil, &SelectionError{Missing: missing}
	}

	return want2pos, nil
}
