package db

import (
	"bytes"
	"html/template"
	"time"
)

type Column struct {
	Name        string
	Source      string
	Type        string
	Description string
}

var (
	ColBezoeknummer      = Column{"SEHID", "seh_sehreg.SEHID", "NUMBER", "Uniek Bezoeknummer voor dit bezoek (niet voor de patiënt)"}
	ColMutatieID         = Column{"SEHMUTID", "seh_sehmut.SEHMUTID", "NUMBER", "Uniek ID voor deze mutatie"}
	ColLocatie           = Column{"Locatie", "seh_sehreg.LOCATIECOD", "STRING", "Code van de locatie"}
	ColAfdeling          = Column{"Afdeling", "", "STRING", "Naam van de afdeling, meestal 'seh'"}
	ColAangemeld         = Column{"Aangemaakt", "seh_sehreg.DATUM", "DATE", "Datum waarop dit record is aangemaakt"}
	ColBinnenkomstDatum  = Column{"BinnenkomstDatum", "seh_sehreg.AANKSDATUM", "STRING", "Datum waarop de patiënt is binnengekomen"}
	ColBinnenkomstTijd   = Column{"BinnenkomstTijd", "seh_sehreg.AANKSTIJD", "STRING", "Tijdstip waarop de patiënt is binnengekomen"}
	ColTriageTijd        = Column{"TriageTijd", "seh_sehreg.TRIAGETIJD", "STRING", "Tijdstip waarop de triage is afgerond"}
	ColNaarKamerTijd     = Column{"NaarKamerTijd", "seh_sehreg.ARTSBHTIJD", "STRING", "Tijdstip waarop de patiënt naar de behandelkamer is gegaan"}
	ColBijArtsTijd       = Column{"EersteContactTijd", "seh_sehreg.PATGEZT", "STRING", "Tijdstip waarop de patiënt voor het eerst contact heeft gehad met de behandelend arts"}
	ColArtsKlaarTijd     = Column{"ArtsKlaarTijd", "seh_sehref.ARTSKLAARTIJD", "STRING", "Tijdstip waarop de arts volledig klaar is met de behandeling van de patiënt"}
	ColGereedOpnameTijd  = Column{"GereedOpnameTijd", "opname_opname.INSCHRTIJD", "STRING", "Tijdstip waarop de patiënt is aangemerkt voor opname"}
	ColVertrekTijd       = Column{"VertrekTijd", "seh_sehreg.ARBEHETIJD", "STRING", "Tijdstip waarop de patiënt is vertrokken"}
	ColEindTijd          = Column{"EindTijd", "seh_sehreg.eindtijd", "STRING", "Tijdstip waarop het bezoek administratief is afgerond"}
	ColKamer             = Column{"Kamer", "seh_sehmut.BEHKAMERCO", "STRING", "Code van de behandelkamer"}
	ColBed               = Column{"Bed", "seh_sehmut.BEDNR", "STRING", "Bed nummer"}
	ColIngangsklacht     = Column{"Ingangsklacht", "", "STRING", "Ingangsklacht"}
	ColSpecialisme       = Column{"Specialisme", "seh_sehreg.SPECIALISM", "STRING", "Code van het specialisme waar de patiënt aan is toegewezen"}
	ColUrgentie          = Column{"Triage", "seh_sehreg.TRIANIVCOD", "STRING", "Triage code"}
	ColVervoerder        = Column{"Vervoerder", "seh_sehreg.VVCODE", "STRING", "Code van de vervoerder"}
	ColGeboortedatum     = Column{"Geboortedatum", "patient_patient.GEBDAT", "DATE", "Geboortedatum van de patient. Deze wordt automatisch omgezet naar een leeftijdscategorie voordat deze verstuurd wordt"}
	ColOpnameAfdeling    = Column{"OpnameAfdeling", "opname_opname.AFDELING", "STRING", "Afdeling waar de patiënt is opgenomen"}
	ColOpnameSpecialisme = Column{"OpnameSpecialisme", "opname_opname.SPECIALISM", "STRING", "Specialisme waar de patiënt is opgenomen"}
	ColHerkomst          = Column{"Herkomst", "seh_sehreg.VERVOERTYP", "STRING", "Code van de herkomst van de patiënt"}
	ColOntslagbestemming = Column{"OntslagBestemming", "seh_sehreg.BESTEMMING", "STRING", "Code van de ontslagbestemming"}
	ColVervallen         = Column{"Vervallen", "seh_sehreg.VERVALL", "NUMBER", "Is dit record vervallen?"}
	ColMutatieEindTijd   = Column{"MutatieEindTijd", "seh_sehmut.eindtijd", "STRING", "Eindtijd van deze mutatie"}
	ColMutatieStatus     = Column{"MutatieStatus", "seh_sehmut.status", "STRING", "Statuscode van deze mutatie"}
)

var VisitorColumns = []Column{
	ColBezoeknummer,
	ColMutatieID,
	ColLocatie,
	ColAfdeling,
	ColAangemeld,
	ColBinnenkomstDatum,
	ColBinnenkomstTijd,
	ColTriageTijd,
	ColNaarKamerTijd,
	ColBijArtsTijd,
	ColArtsKlaarTijd,
	ColGereedOpnameTijd,
	ColVertrekTijd,
	ColEindTijd,
	ColMutatieEindTijd,
	ColMutatieStatus,
	ColKamer,
	ColBed,
	ColIngangsklacht,
	ColSpecialisme,
	ColUrgentie,
	ColVervoerder,
	ColGeboortedatum,
	ColOpnameAfdeling,
	ColOpnameSpecialisme,
	ColHerkomst,
	ColOntslagbestemming,
	ColVervallen,
}

type VisitorRecord struct {
	Bezoeknummer      int
	MutatieID         int
	Locatie           string
	Afdeling          string
	Aangemeld         time.Time
	BinnenkomstDatum  string
	BinnenkomstTijd   string
	TriageTijd        string
	NaarKamerTijd     string
	BijArtsTijd       string
	ArtsKlaarTijd     string
	GereedOpnameTijd  string
	VertrekTijd       string
	EindTijd          string
	MutatieEindTijd   string
	Mutatiestatus     string
	Kamer             string
	Bed               string
	Ingangsklacht     string
	Specialisme       string
	Urgentie          string
	Vervoerder        string
	Geboortedatum     time.Time
	OpnameAfdeling    string
	OpnameSpecialisme string
	Herkomst          string
	Ontslagbestemming string
	Vervallen         bool
}

type VisitorRecords []VisitorRecord

func (v VisitorRecords) AsTable() template.HTML {
	var buf bytes.Buffer
	if err := tableTmpl.Execute(&buf, struct {
		Columns      []Column
		QueryResults []VisitorRecord
	}{VisitorColumns, v}); err != nil {
		panic(err)
	}

	return template.HTML(buf.String())
}

var tableTmpl = template.Must(template.New("table").Parse(`
<table class="table">
	<thead>
	<tr>
		{{ range $i, $row := .Columns }}
		<th>{{ $row.Name }}</th>
		{{ end }}
	</tr>
	</thead>
	<tbody>
	{{ range $index, $row := .QueryResults }}
		<tr>
			<td>{{ $row.Bezoeknummer }}</td>
			<td>{{ $row.MutatieID }}</td>
			<td>{{ $row.Locatie }}</td>
			<td>{{ $row.Afdeling }}</td>
			<td>{{ $row.Aangemeld }}</td>
			<td>{{ $row.BinnenkomstDatum }}</td>
			<td>{{ $row.BinnenkomstTijd }}</td>
			<td>{{ $row.TriageTijd }}</td>
			<td>{{ $row.NaarKamerTijd }}</td>
			<td>{{ $row.BijArtsTijd }}</td>
			<td>{{ $row.ArtsKlaarTijd }}</td>
			<td>{{ $row.GereedOpnameTijd }}</td>
			<td>{{ $row.VertrekTijd }}</td>
			<td>{{ $row.EindTijd }}</td>
			<td>{{ $row.MutatieEindTijd }}</td>
			<td>{{ $row.Mutatiestatus }}</td>
			<td>{{ $row.Kamer }}</td>
			<td>{{ $row.Bed }}</td>
			<td>{{ $row.Ingangsklacht }}</td>
			<td>{{ $row.Specialisme }}</td>
			<td>{{ $row.Urgentie }}</td>
			<td>{{ $row.Vervoerder }}</td>
			<td>{{ $row.Geboortedatum }}</td>
			<td>{{ $row.OpnameAfdeling }}</td>
			<td>{{ $row.OpnameSpecialisme }}</td>
			<td>{{ $row.Herkomst }}</td>
			<td>{{ $row.Ontslagbestemming }}</td>
			<td>{{ $row.Vervallen }}</td>
		</tr>
	{{ end }}
	</tbody>
</table>
`))
