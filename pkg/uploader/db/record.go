package db

import "time"

const (
	ColID                 = "SEHID"
	ColMutatieID          = "SEHMUTID"
	ColLocatie            = "Locatie"
	ColAangemaakt         = "Aangemaakt"
	ColBinnenkomstDatum   = "BinnenkomstDatum"
	ColBinnenkomstTijd    = "BinnenkomstTijd"
	ColAanvangTriageTijd  = "AanvangTriageTijd"
	ColNaarKamerTijd      = "NaarKamerTijd"
	ColEersteContactTijd  = "EersteContactTijd"
	ColAfdelingGebeldTijd = "AfdelingGebeldTijd"
	ColGereedOpnameTijd   = "GereedOpnameTijd"
	ColVertrekTijd        = "VertrekTijd"
	ColKamer              = "Kamer"
	ColBed                = "Bed"
	ColIngangsklacht      = "Ingangsklacht"
	ColSpecialisme        = "Specialisme"
	ColTriage             = "Triage"
	ColVervoerder         = "Vervoerder"
	ColBestemming         = "Bestemming"
	ColGeboortedatum      = "Geboortedatum"
	ColOpnameAfdeling     = "OpnameAfdeling"
	ColOpnameSpecialisme  = "OpnameSpecialisme"
	ColHerkomst           = "Herkomst"
	ColOntslagbestemming  = "OntslagBestemming"
)

var columns = []string{
	ColID,
	ColMutatieID,
	ColLocatie,
	ColAangemaakt,
	ColBinnenkomstDatum,
	ColBinnenkomstTijd,
	ColAanvangTriageTijd,
	ColNaarKamerTijd,
	ColEersteContactTijd,
	ColAfdelingGebeldTijd,
	ColGereedOpnameTijd,
	ColVertrekTijd,
	ColKamer,
	ColBed,
	ColIngangsklacht,
	ColSpecialisme,
	ColTriage,
	ColVervoerder,
	ColBestemming,
	ColGeboortedatum,
	ColOpnameAfdeling,
	ColOpnameSpecialisme,
	ColHerkomst,
	ColOntslagbestemming,
}

type VisitorRecord struct {
	ID                 int
	MutatieID          int
	Locatie            string
	Aangemaakt         time.Time
	BinnenkomstDatum   string
	BinnenkomstTijd    string
	AanvangTriageTijd  string
	NaarKamerTijd      string
	EersteContactTijd  string
	AfdelingGebeldTijd string
	GereedOpnameTijd   string
	VertrekTijd        string
	Kamer              string
	Bed                string
	Ingangsklacht      string
	Specialisme        string
	Triage             string
	Vervoerder         string
	Bestemming         string
	Geboortedatum      time.Time
	OpnameAfdeling     string
	OpnameSpecialisme  string
	Herkomst           string
	Ontslagbestemming  string
}
