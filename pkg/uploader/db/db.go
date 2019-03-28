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
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			dlog.Error("While closing result set: %v", err)
		}
	}()

	// determine column names
	names, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	col2index, err := checkColumnNames(names, VisitorColumns)
	if err != nil {
		return nil, err
	}

	// map result set to records
	var res []VisitorRecord

	for rows.Next() {
		var rec VisitorRecord
		err := mapRow(rows, &rec, names, col2index)
		if err != nil {
			return nil, err
		}

		target := make([]interface{}, len(names))
		for i := range target {
			target[i] = new(sql.RawBytes)
		}
		if err := rows.Scan(target...); err != nil {
			return nil, err
		}

		res = append(res, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
		geboortedatum      *time.Time
		opnameAfdeling     sql.NullString
		opnameSpecialisme  sql.NullString
		herkomst           sql.NullString
		ontslagbestemming  sql.NullString
	)

	target[col2index[ColID.Name]] = &id
	target[col2index[ColMutatieID.Name]] = &mutatieID
	target[col2index[ColLocatie.Name]] = &locatie
	target[col2index[ColAangemaakt.Name]] = &aangemaakt
	target[col2index[ColBinnenkomstDatum.Name]] = &binnenkomstDatum
	target[col2index[ColBinnenkomstTijd.Name]] = &binnenkomstTijd
	target[col2index[ColAanvangTriageTijd.Name]] = &aanvangTriageTijd
	target[col2index[ColNaarKamerTijd.Name]] = &naarKamerTijd
	target[col2index[ColEersteContactTijd.Name]] = &eersteContactTijd
	target[col2index[ColAfdelingGebeldTijd.Name]] = &afdelingGebeldTijd
	target[col2index[ColGereedOpnameTijd.Name]] = &gereedOpnameTijd
	target[col2index[ColVertrekTijd.Name]] = &vertrekTijd
	target[col2index[ColKamer.Name]] = &kamer
	target[col2index[ColBed.Name]] = &bed
	target[col2index[ColIngangsklacht.Name]] = &ingangsklacht
	target[col2index[ColSpecialisme.Name]] = &specialisme
	target[col2index[ColTriage.Name]] = &triage
	target[col2index[ColVervoerder.Name]] = &vervoerder
	target[col2index[ColGeboortedatum.Name]] = &geboortedatum
	target[col2index[ColOpnameAfdeling.Name]] = &opnameAfdeling
	target[col2index[ColOpnameSpecialisme.Name]] = &opnameSpecialisme
	target[col2index[ColHerkomst.Name]] = &herkomst
	target[col2index[ColOntslagbestemming.Name]] = &ontslagbestemming

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
		BinnenkomstTijd:    asTime(binnenkomstTijd.String),
		AanvangTriageTijd:  asTime(aanvangTriageTijd.String),
		NaarKamerTijd:      asTime(naarKamerTijd.String),
		EersteContactTijd:  asTime(eersteContactTijd.String),
		AfdelingGebeldTijd: asTime(afdelingGebeldTijd.String),
		GereedOpnameTijd:   asTime(gereedOpnameTijd.String),
		VertrekTijd:        asTime(vertrekTijd.String),
		Kamer:              kamer.String,
		Bed:                bed.String,
		Ingangsklacht:      ingangsklacht.String,
		Specialisme:        specialisme.String,
		Triage:             triage.String,
		Vervoerder:         vervoerder.String,
		Geboortedatum:      geb,
		OpnameAfdeling:     opnameAfdeling.String,
		OpnameSpecialisme:  opnameSpecialisme.String,
		Herkomst:           herkomst.String,
		Ontslagbestemming:  ontslagbestemming.String,
	}

	return nil
}

func asTime(s string) string {
	s = strings.TrimPrefix(s, "0001-01-01T")
	s = strings.TrimSuffix(s, ":00Z")
	return s
}

func checkColumnNames(got []string, want []Column) (map[string]int, error) {
	got2pos := make(map[string]int)
	for i, s := range got {
		got2pos[strings.ToLower(s)] = i
	}

	if len(got) != len(got2pos) {
		return nil, ErrDuplicateColumnNames
	}

	var (
		missing  []string
		want2pos = make(map[string]int)
	)

	for _, w := range want {
		idx, ok := got2pos[strings.ToLower(w.Name)]
		if ok {
			want2pos[w.Name] = idx
		} else {
			missing = append(missing, w.Name)
		}
	}

	if len(missing) != 0 {
		return nil, &SelectionError{Missing: missing}
	}

	return want2pos, nil
}
