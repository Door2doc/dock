package db

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/door2doc/d2d-uploader/pkg/uploader/dlog"
)

// ExecuteVisitorQuery tries to execute the visitor query and marshal the result into records.
func ExecuteVisitorQuery(ctx context.Context, tx *sql.Tx, query string) (VisitorRecords, error) {
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
		err := mapVisitorRow(rows, &rec, names, col2index)
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

func mapVisitorRow(rows *sql.Rows, rec *VisitorRecord, allColumns []string, col2index map[string]int) error {
	target := make([]interface{}, len(allColumns))
	for i := range target {
		target[i] = new(sql.RawBytes)
	}

	var (
		bezoeknummer      int
		mutatieID         int
		locatie           sql.NullString
		afdeling          sql.NullString
		aangemaakt        *time.Time
		binnenkomstDatum  sql.NullString
		binnenkomstTijd   sql.NullString
		aanvangTriageTijd sql.NullString
		naarKamerTijd     sql.NullString
		bijArtsTijd       sql.NullString
		artsKlaarTijd     sql.NullString
		gereedOpnameTijd  sql.NullString
		vertrekTijd       sql.NullString
		eindTijd          sql.NullString
		mutatieEindTijd   sql.NullString
		mutatieStatus     sql.NullString
		kamer             sql.NullString
		bed               sql.NullString
		ingangsklacht     sql.NullString
		specialisme       sql.NullString
		urgentie          sql.NullString
		vervoerder        sql.NullString
		geboortedatum     *time.Time
		opnameAfdeling    sql.NullString
		opnameSpecialisme sql.NullString
		herkomst          sql.NullString
		ontslagbestemming sql.NullString
		vervallen         bool
	)

	target[col2index[ColBezoeknummer.Name]] = &bezoeknummer
	target[col2index[ColMutatieID.Name]] = &mutatieID
	target[col2index[ColLocatie.Name]] = &locatie
	target[col2index[ColAfdeling.Name]] = &afdeling
	target[col2index[ColAangemeld.Name]] = &aangemaakt
	target[col2index[ColBinnenkomstDatum.Name]] = &binnenkomstDatum
	target[col2index[ColBinnenkomstTijd.Name]] = &binnenkomstTijd
	target[col2index[ColTriageTijd.Name]] = &aanvangTriageTijd
	target[col2index[ColNaarKamerTijd.Name]] = &naarKamerTijd
	target[col2index[ColBijArtsTijd.Name]] = &bijArtsTijd
	target[col2index[ColArtsKlaarTijd.Name]] = &artsKlaarTijd
	target[col2index[ColGereedOpnameTijd.Name]] = &gereedOpnameTijd
	target[col2index[ColVertrekTijd.Name]] = &vertrekTijd
	target[col2index[ColEindTijd.Name]] = &eindTijd
	target[col2index[ColKamer.Name]] = &kamer
	target[col2index[ColBed.Name]] = &bed
	target[col2index[ColIngangsklacht.Name]] = &ingangsklacht
	target[col2index[ColSpecialisme.Name]] = &specialisme
	target[col2index[ColUrgentie.Name]] = &urgentie
	target[col2index[ColVervoerder.Name]] = &vervoerder
	target[col2index[ColGeboortedatum.Name]] = &geboortedatum
	target[col2index[ColOpnameAfdeling.Name]] = &opnameAfdeling
	target[col2index[ColOpnameSpecialisme.Name]] = &opnameSpecialisme
	target[col2index[ColHerkomst.Name]] = &herkomst
	target[col2index[ColOntslagbestemming.Name]] = &ontslagbestemming
	target[col2index[ColVervallen.Name]] = &vervallen
	target[col2index[ColMutatieEindTijd.Name]] = &mutatieEindTijd
	target[col2index[ColMutatieStatus.Name]] = &mutatieStatus

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
		Bezoeknummer:      bezoeknummer,
		MutatieID:         mutatieID,
		Locatie:           locatie.String,
		Afdeling:          afdeling.String,
		Aangemeld:         created,
		BinnenkomstDatum:  binnenkomstDatum.String,
		BinnenkomstTijd:   asTime(binnenkomstTijd.String),
		TriageTijd:        asTime(aanvangTriageTijd.String),
		NaarKamerTijd:     asTime(naarKamerTijd.String),
		BijArtsTijd:       asTime(bijArtsTijd.String),
		ArtsKlaarTijd:     asTime(artsKlaarTijd.String),
		GereedOpnameTijd:  asTime(gereedOpnameTijd.String),
		VertrekTijd:       asTime(vertrekTijd.String),
		EindTijd:          asTime(eindTijd.String),
		MutatieEindTijd:   asTime(mutatieEindTijd.String),
		Mutatiestatus:     mutatieStatus.String,
		Kamer:             kamer.String,
		Bed:               bed.String,
		Ingangsklacht:     ingangsklacht.String,
		Specialisme:       specialisme.String,
		Urgentie:          urgentie.String,
		Vervoerder:        vervoerder.String,
		Geboortedatum:     geb,
		OpnameAfdeling:    opnameAfdeling.String,
		OpnameSpecialisme: opnameSpecialisme.String,
		Herkomst:          herkomst.String,
		Ontslagbestemming: ontslagbestemming.String,
		Vervallen:         vervallen,
	}

	return nil
}

// ExecuteRadiologieQuery tries to execute the visitor query and marshal the result into records.
func ExecuteRadiologieQuery(ctx context.Context, tx *sql.Tx, query string) (RadiologieOrders, error) {
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

	col2index, err := checkColumnNames(names, RadiologieColumns)
	if err != nil {
		return nil, err
	}

	// map result set to records
	var res []RadiologieOrder

	for rows.Next() {
		var rec RadiologieOrder
		err := mapRadiologieRow(rows, &rec, names, col2index)
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

func mapRadiologieRow(rows *sql.Rows, rec *RadiologieOrder, allColumns []string, col2index map[string]int) error {
	target := make([]interface{}, len(allColumns))
	for i := range target {
		target[i] = new(sql.RawBytes)
	}

	var (
		bezoeknummer int
		orderNummer  int
		status       sql.NullString
		start        sql.NullTime
		eind         sql.NullTime
		module       sql.NullString
	)

	target[col2index[ColBezoeknummer.Name]] = &bezoeknummer
	target[col2index[ColOrderNummer.Name]] = &orderNummer
	target[col2index[ColOrderStatus.Name]] = &status
	target[col2index[ColOrderStart.Name]] = &start
	target[col2index[ColOrderEind.Name]] = &eind
	target[col2index[ColOrderModule.Name]] = &module

	if err := rows.Scan(target...); err != nil {
		return err
	}

	*rec = RadiologieOrder{
		Bezoeknummer: bezoeknummer,
		Ordernummer:  orderNummer,
		Status:       status.String,
		Start:        asTimeRef(start),
		Eind:         asTimeRef(eind),
		Module:       module.String,
	}

	return nil
}

// ExecuteLabQuery tries to execute the visitor query and marshal the result into records.
func ExecuteLabQuery(ctx context.Context, tx *sql.Tx, query string) (LabOrders, error) {
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

	col2index, err := checkColumnNames(names, LabColumns)
	if err != nil {
		return nil, err
	}

	// map result set to records
	var res []LabOrder

	for rows.Next() {
		var rec LabOrder
		err := mapLabRow(rows, &rec, names, col2index)
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

func mapLabRow(rows *sql.Rows, rec *LabOrder, allColumns []string, col2index map[string]int) error {
	target := make([]interface{}, len(allColumns))
	for i := range target {
		target[i] = new(sql.RawBytes)
	}

	var (
		bezoeknummer int
		orderNummer  int
		status       sql.NullString
		start        sql.NullTime
		eind         sql.NullTime
	)

	target[col2index[ColBezoeknummer.Name]] = &bezoeknummer
	target[col2index[ColOrderNummer.Name]] = &orderNummer
	target[col2index[ColOrderStatus.Name]] = &status
	target[col2index[ColOrderStart.Name]] = &start
	target[col2index[ColOrderEind.Name]] = &eind

	if err := rows.Scan(target...); err != nil {
		return err
	}

	*rec = LabOrder{
		Bezoeknummer: bezoeknummer,
		Ordernummer:  orderNummer,
		Status:       status.String,
		Start:        asTimeRef(start),
		Eind:         asTimeRef(eind),
	}

	return nil
}

// ExecuteConsultQuery tries to execute the visitor query and marshal the result into records.
func ExecuteConsultQuery(ctx context.Context, tx *sql.Tx, query string) (ConsultOrders, error) {
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

	col2index, err := checkColumnNames(names, ConsultColumns)
	if err != nil {
		return nil, err
	}

	// map result set to records
	var res []ConsultOrder

	for rows.Next() {
		var rec ConsultOrder
		err := mapConsultRow(rows, &rec, names, col2index)
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

func mapConsultRow(rows *sql.Rows, rec *ConsultOrder, allColumns []string, col2index map[string]int) error {
	target := make([]interface{}, len(allColumns))
	for i := range target {
		target[i] = new(sql.RawBytes)
	}

	var (
		bezoeknummer int
		orderNummer  int
		status       sql.NullString
		start        sql.NullTime
		eind         sql.NullTime
		specialisme  sql.NullString
	)

	target[col2index[ColBezoeknummer.Name]] = &bezoeknummer
	target[col2index[ColOrderNummer.Name]] = &orderNummer
	target[col2index[ColOrderStatus.Name]] = &status
	target[col2index[ColOrderStart.Name]] = &start
	target[col2index[ColOrderEind.Name]] = &eind
	target[col2index[ColOrderSpecialisme.Name]] = &specialisme

	if err := rows.Scan(target...); err != nil {
		return err
	}

	*rec = ConsultOrder{
		Bezoeknummer: bezoeknummer,
		Ordernummer:  orderNummer,
		Status:       status.String,
		Start:        asTimeRef(start),
		Eind:         asTimeRef(eind),
		Specialisme:  specialisme.String,
	}

	return nil
}

func asTimeRef(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
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
		return nil, &SelectionError{Missing: missing, Got: got}
	}

	return want2pos, nil
}
