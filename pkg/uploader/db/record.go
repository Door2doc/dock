package db

import "time"

type Column struct {
	Name        string
	Source      string
	Type        string
	Description string
}

var (
	ColID                 = Column{"SEHID", "seh_sehreg.SEHID", "NUMBER", "Uniek ID voor dit bezoek (niet voor de patiënt)"}
	ColMutatieID          = Column{"SEHMUTID", "seh_sehmut.SEHMUTID", "NUMBER", "Uniek ID voor deze mutatie"}
	ColLocatie            = Column{"Locatie", "seh_sehreg.LOCATIECOD", "STRING", "Code van de locatie"}
	ColAangemaakt         = Column{"Aangemaakt", "seh_sehreg.DATUM", "DATE", "Datum waarop dit record is aangemaakt"}
	ColBinnenkomstDatum   = Column{"BinnenkomstDatum", "seh_sehreg.AANKSDATUM", "STRING", "Datum waarop de patiënt is binnengekomen"}
	ColBinnenkomstTijd    = Column{"BinnenkomstTijd", "seh_sehreg.AANKSTIJD", "STRING", "Tijdstip waarop de patiënt is binnengekomen"}
	ColAanvangTriageTijd  = Column{"AanvangTriageTijd", "seh_sehreg.TRIAGETIJD", "STRING", "Tijdstip waarop de triage is begonnen"}
	ColNaarKamerTijd      = Column{"NaarKamerTijd", "seh_sehreg.ARTSBHTIJD", "STRING", "Tijdstip waarop de patiënt naar de behandelkamer is gegaan"}
	ColEersteContactTijd  = Column{"EersteContactTijd", "seh_sehreg.PATGEZT", "STRING", "Tijdstip waarop de patiënt voor het eerst contact heeft gehad met de behandelend arts"}
	ColAfdelingGebeldTijd = Column{"AfdelingGebeldTijd", "", "STRING", "Tijdstip waarop de opname afdeling is gebeld"}
	ColGereedOpnameTijd   = Column{"GereedOpnameTijd", "opname_opname.INSCHRTIJD", "STRING", "Tijdstip waarop de patiënt is aangemerkt voor opname"}
	ColVertrekTijd        = Column{"VertrekTijd", "seh_sehreg.ARBEHETIJD", "STRING", "Tijdstip waarop de patiënt is vertrokken"}
	ColKamer              = Column{"Kamer", "seh_sehmut.BEHKAMERCO", "STRING", "Code van de behandelkamer"}
	ColBed                = Column{"Bed", "seh_sehmut.BEDNR", "STRING", "Bed nummer"}
	ColIngangsklacht      = Column{"Ingangsklacht", "", "STRING", "Ingangsklacht"}
	ColSpecialisme        = Column{"Specialisme", "seh_sehreg.SPECIALISM", "STRING", "Code van het specialisme waar de patiënt aan is toegewezen"}
	ColTriage             = Column{"Triage", "seh_sehreg.TRIANIVCOD", "STRING", "Triage code"}
	ColVervoerder         = Column{"Vervoerder", "seh_sehreg.VVCODE", "STRING", "Code van de vervoerder"}
	ColGeboortedatum      = Column{"Geboortedatum", "patient_patient.GEBDAT", "DATE", "Geboortedatum van de patient. Deze wordt automatisch omgezet naar een leeftijdscategorie voordat deze verstuurd wordt"}
	ColOpnameAfdeling     = Column{"OpnameAfdeling", "opname_opname.AFDELING", "STRING", "Afdeling waar de patiënt is opgenomen"}
	ColOpnameSpecialisme  = Column{"OpnameSpecialisme", "opname_opname.SPECIALISM", "STRING", "Specialisme waar de patiënt is opgenomen"}
	ColHerkomst           = Column{"Herkomst", "seh_sehreg.VERVOERTYP", "STRING", "Code van de herkomst van de patiënt"}
	ColOntslagbestemming  = Column{"OntslagBestemming", "seh_sehreg.BESTEMMING", "STRING", "Code van de ontslagbestemming"}
)

var columns = []string{
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
	Geboortedatum      time.Time
	OpnameAfdeling     string
	OpnameSpecialisme  string
	Herkomst           string
	Ontslagbestemming  string
}
