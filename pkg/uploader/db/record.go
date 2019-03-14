package db

import "time"

const (
	ColID                = "SEHID"
	ColMutatieID         = "SEHMUTID"
	ColLocatie           = "LOCATIECOD"
	ColAanmaakDatum      = "AANMAAKDAT"
	ColAanmaakTijd       = "AANMAAKTIJD"
	ColAankomstDatum     = "AANKSDATUM"
	ColAankomstTijd      = "AANKSTIJD"
	ColTriageDatum       = "TRIADATUM"
	ColTriageTijd        = "TRIAGETIJD"
	ColBehandelTijd      = "ARTSBHTIJD"
	ColGezienTijd        = "PATGEZT"
	ColGebeldTijd        = "GEBELD"
	ColOpnameTijd        = "INSCHRTIJD"
	ColVertrekTijd       = "ARBEHETIJD"
	ColKamer             = "BEHKAMERCO"
	ColBed               = "BEDNR"
	ColKlacht            = "KLACHT"
	ColSpecialisme       = "SPECIALISM"
	ColTriage            = "TRIANIVCOD"
	ColVervoer           = "VERVOERTYP"
	ColBestemming        = "BESTEMMING"
	ColGeboortedatum     = "GEBDAT"
	ColOpnameAfdeling    = "OPNAMEAFD"
	ColOpnameSpecialisme = "OPNAMESPEC"
)

var columns = []string{
	ColID,
	ColMutatieID,
	ColLocatie,
	ColAanmaakDatum,
	ColAanmaakTijd,
	ColAankomstDatum,
	ColAankomstTijd,
	ColTriageDatum,
	ColTriageTijd,
	ColBehandelTijd,
	ColGezienTijd,
	ColGebeldTijd,
	ColOpnameTijd,
	ColVertrekTijd,
	ColKamer,
	ColBed,
	ColKlacht,
	ColSpecialisme,
	ColTriage,
	ColVervoer,
	ColBestemming,
	ColGeboortedatum,
	ColOpnameAfdeling,
	ColOpnameSpecialisme,
}

type Record struct {
	ID            int
	MutatieID     int
	Locatie       string
	AanmaakDatum  string
	AanmaakTijd   string
	AankomstDatum string
	AankomstTijd  string

	// Deprecated
	TriageDatum       string
	TriageTijd        string
	BehandelTijd      string
	GezienTijd        string
	GebeldTijd        string
	OpnameTijd        string
	VertrekTijd       string
	Kamer             string
	Bed               string
	Klacht            string
	Specialisme       string
	Triage            string
	Vervoer           string
	Bestemming        string
	Geboortedatum     time.Time
	OpnameAfdeling    string
	OpnameSpecialisme string

	// TODO
	Herkomst          string
	Ontslagbestemming string
}
