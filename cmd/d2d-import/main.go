// d2d-import is the offline version of d2d-upload using CSV output of the specified query. It reads the query from
// stdin and sends it in batches to the cloud upload service.
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/door2doc/d2d-uploader/pkg/uploader/config"
	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
	"github.com/door2doc/d2d-uploader/pkg/uploader/rest"
	"github.com/pkg/errors"
)

var (
	username = flag.String("username", "test", "Username for connecting to the upload service")
	password = flag.String("password", "", "Password for connecting to the upload service")
	server   = flag.String("server", config.Server, "Server to upload to")
	from     = flag.Int("from", 0, "Skip all mutaties < from")
)

type record struct {
	ID                 string
	MutatieID          string
	Locatie            string
	Aangemaakt         string
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
	Geboortedatum      string
	OpnameAfdeling     string
	OpnameSpecialisme  string
	Herkomst           string
	Ontslagbestemming  string
}

var requiredHeader = []string{
	"sehid", "sehmutid", "locatie", "aangemaakt", "binnenkomstdatum", "binnenkomsttijd", "aanvangtriagetijd",
	"naarkamertijd", "eerstecontacttijd", "afdelinggebeldtijd", "gereedopnametijd", "vertrektijd", "kamer", "bed",
	"ingangsklacht", "specialisme", "triage", "vervoerder", "geboortedatum", "opnameafdeling", "opnamespecialisme",
	"herkomst", "ontslagbestemming",
}

const batch = 100

var timezone *time.Location

func init() {
	var err error
	timezone, err = time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := ping(); err != nil {
		return err
	}

	r := csv.NewReader(os.Stdin)
	header, err := r.Read()
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(header, requiredHeader) {
		return errors.New("unsupported format")
	}
	for {
		csvRecords, err := readBatch(r, batch)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		visitorRecords, err := toVisitorRecords(csvRecords)
		if err != nil {
			return err
		}
		if len(visitorRecords) == 0 {
			continue
		}

		upload, err := rest.VisitorRecordsFromDB(visitorRecords, timezone)
		if err != nil {
			return err
		}

		buf := new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(upload); err != nil {
			return err
		}

		for {
			panic("fix this part")
			//if err := uploader.UploadJSON(context.Background(), *username, *password, buf); err != nil {
			//	log.Println("Error", err)
			//	<-time.After(time.Second)
			//	continue
			//}
			break
		}
		log.Printf("Upload %d--%d OK", visitorRecords[0].MutatieID, visitorRecords[len(visitorRecords)-1].MutatieID)
	}
	return nil
}

func ping() error {
	req, err := http.NewRequest(http.MethodGet, *server, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(*username, *password)
	req.URL.Path = config.PathPing
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("Ping failed")
	}
	return nil
}

// readBatch reads at most batchSize records from the csv reader. If an EOF is encountered, the length of the result
// might be smaller than the requested batch size. This function returns an EOF when there are no more results to
// consume.
func readBatch(r *csv.Reader, batchSize int) ([]record, error) {
	res := make([]record, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		fields, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		res = append(res, record{
			ID:                 fields[0],
			MutatieID:          fields[1],
			Locatie:            fields[2],
			Aangemaakt:         fields[3],
			BinnenkomstDatum:   strings.TrimSuffix(fields[4], " 00:00:00"),
			BinnenkomstTijd:    fields[5],
			AanvangTriageTijd:  fields[6],
			NaarKamerTijd:      fields[7],
			EersteContactTijd:  fields[8],
			AfdelingGebeldTijd: fields[9],
			GereedOpnameTijd:   fields[10],
			VertrekTijd:        fields[11],
			Kamer:              fields[12],
			Bed:                fields[13],
			Ingangsklacht:      fields[14],
			Specialisme:        fields[15],
			Triage:             fields[16],
			Vervoerder:         fields[17],
			Geboortedatum:      fields[18],
			OpnameAfdeling:     fields[19],
			OpnameSpecialisme:  fields[20],
			Herkomst:           fields[21],
			Ontslagbestemming:  fields[22],
		})
	}

	if len(res) == 0 {
		return nil, io.EOF
	}
	return res, nil
}

func toVisitorRecords(records []record) ([]db.VisitorRecord, error) {
	res := make([]db.VisitorRecord, 0, len(records))
	for _, rec := range records {
		if rec.Locatie == "" {
			log.Printf("Skipping mutatie %s, geen locatie", rec.MutatieID)
			continue
		}

		id, err := strconv.Atoi(rec.ID)
		if err != nil {
			return nil, err
		}
		mutID, err := strconv.Atoi(rec.MutatieID)
		if err != nil {
			return nil, err
		}
		if mutID < *from {
			continue
		}

		aangemaakt, err := time.ParseInLocation("2006-01-02 15:04:05", rec.Aangemaakt, timezone)
		if err != nil {
			return nil, err
		}
		geb, err := time.Parse("2006-01-02 15:04:05", rec.Geboortedatum)
		if err != nil {
			return nil, err
		}

		res = append(res, db.VisitorRecord{
			ID:                 id,
			MutatieID:          mutID,
			Locatie:            rec.Locatie,
			Aangemaakt:         aangemaakt,
			BinnenkomstDatum:   rec.BinnenkomstDatum,
			BinnenkomstTijd:    rec.BinnenkomstTijd,
			AanvangTriageTijd:  rec.AanvangTriageTijd,
			NaarKamerTijd:      rec.NaarKamerTijd,
			EersteContactTijd:  rec.EersteContactTijd,
			AfdelingGebeldTijd: rec.AfdelingGebeldTijd,
			GereedOpnameTijd:   rec.GereedOpnameTijd,
			VertrekTijd:        rec.VertrekTijd,
			Kamer:              rec.Kamer,
			Bed:                rec.Bed,
			Ingangsklacht:      rec.Ingangsklacht,
			Specialisme:        rec.Specialisme,
			Triage:             rec.Triage,
			Vervoerder:         rec.Vervoerder,
			Geboortedatum:      geb,
			OpnameAfdeling:     rec.OpnameAfdeling,
			OpnameSpecialisme:  rec.OpnameSpecialisme,
			Herkomst:           rec.Herkomst,
			Ontslagbestemming:  rec.Ontslagbestemming,
		})
	}
	return res, nil
}
