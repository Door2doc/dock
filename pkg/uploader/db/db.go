package db

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
)

// ExecuteVisitorQuery tries to execute the visitor query and marshal the result into records.
func ExecuteVisitorQuery(ctx context.Context, tx *sql.Tx, query string) ([]VisitorRecord, error) {
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
	var res []VisitorRecord

	for rows.Next() {
		var rec VisitorRecord
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

func mapRow(rows *sql.Rows, rec *VisitorRecord, allColumns []string, col2index map[string]int) error {
	target := make([]interface{}, len(allColumns))
	for i := range target {
		target[i] = new(sql.RawBytes)
	}

	var (
		id                 int
		mutatieID          int
		locatie            sql.NullString
		aangemaakt         *time.Time
		binnenkomstDatum   sql.NullString
		binnenkomstTijd    sql.NullString
		aanvangTriageTijd  sql.NullString
		naarKamerTijd      sql.NullString
		eersteContactTijd  sql.NullString
		afdelingGebeldTijd sql.NullString
		gereedOpnameTijd   sql.NullString
		vertrekTijd        sql.NullString
		kamer              sql.NullString
		bed                sql.NullString
		ingangsklacht      sql.NullString
		specialisme        sql.NullString
		triage             sql.NullString
		vervoerder         sql.NullString
		bestemming         sql.NullString
		geboortedatum      *time.Time
		opnameAfdeling     sql.NullString
		opnameSpecialisme  sql.NullString
		herkomst           sql.NullString
		ontslagbestemming  sql.NullString
	)

	target[col2index[ColID]] = &id
	target[col2index[ColMutatieID]] = &mutatieID
	target[col2index[ColLocatie]] = &locatie
	target[col2index[ColAangemaakt]] = &aangemaakt
	target[col2index[ColBinnenkomstDatum]] = &binnenkomstDatum
	target[col2index[ColBinnenkomstTijd]] = &binnenkomstTijd
	target[col2index[ColAanvangTriageTijd]] = &aanvangTriageTijd
	target[col2index[ColNaarKamerTijd]] = &naarKamerTijd
	target[col2index[ColEersteContactTijd]] = &eersteContactTijd
	target[col2index[ColAfdelingGebeldTijd]] = &afdelingGebeldTijd
	target[col2index[ColGereedOpnameTijd]] = &gereedOpnameTijd
	target[col2index[ColVertrekTijd]] = &vertrekTijd
	target[col2index[ColKamer]] = &kamer
	target[col2index[ColBed]] = &bed
	target[col2index[ColIngangsklacht]] = &ingangsklacht
	target[col2index[ColSpecialisme]] = &specialisme
	target[col2index[ColTriage]] = &triage
	target[col2index[ColVervoerder]] = &vervoerder
	target[col2index[ColBestemming]] = &bestemming
	target[col2index[ColGeboortedatum]] = &geboortedatum
	target[col2index[ColOpnameAfdeling]] = &opnameAfdeling
	target[col2index[ColOpnameSpecialisme]] = &opnameSpecialisme
	target[col2index[ColHerkomst]] = &herkomst
	target[col2index[ColOntslagbestemming]] = &ontslagbestemming

	if err := rows.Scan(target...); err != nil {
		return err
	}

	var geb time.Time
	if geboortedatum != nil {
		geb = *geboortedatum
	}

	var created time.Time
	if aangemaakt != nil {
		created = *aangemaakt
	}

	*rec = VisitorRecord{
		ID:                 id,
		MutatieID:          mutatieID,
		Locatie:            locatie.String,
		Aangemaakt:         created,
		BinnenkomstDatum:   binnenkomstDatum.String,
		BinnenkomstTijd:    binnenkomstTijd.String,
		AanvangTriageTijd:  aanvangTriageTijd.String,
		NaarKamerTijd:      naarKamerTijd.String,
		EersteContactTijd:  eersteContactTijd.String,
		AfdelingGebeldTijd: afdelingGebeldTijd.String,
		GereedOpnameTijd:   gereedOpnameTijd.String,
		VertrekTijd:        vertrekTijd.String,
		Kamer:              kamer.String,
		Bed:                bed.String,
		Ingangsklacht:      ingangsklacht.String,
		Specialisme:        specialisme.String,
		Triage:             triage.String,
		Vervoerder:         vervoerder.String,
		Bestemming:         bestemming.String,
		Geboortedatum:      geb,
		OpnameAfdeling:     opnameAfdeling.String,
		OpnameSpecialisme:  opnameSpecialisme.String,
		Herkomst:           herkomst.String,
		Ontslagbestemming:  ontslagbestemming.String,
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
