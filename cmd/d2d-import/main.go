// d2d-import is the offline version of d2d-upload using CSV output of the specified query. It reads the query from
// stdin and sends it in batches to the cloud upload service.
package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/door2doc/d2d-uploader/pkg/uploader"
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
	test     = flag.Bool("test", true, "Use test mode")
	batch    = flag.Int("batch", 100, "Batch size")
)

type record struct {
	ID                string
	MutatieID         string
	Locatie           string
	Afdeling          string
	Aangemaakt        string
	BinnenkomstDatum  string
	BinnenkomstTijd   string
	AanvangTriageTijd string
	NaarKamerTijd     string
	EersteContactTijd string
	ArtsKlaarTijd     string
	GereedOpnameTijd  string
	VertrekTijd       string
	EindTijd          string
	MutatieEindTijd   string
	MutatieStatus     string
	Kamer             string
	Bed               string
	Ingangsklacht     string
	Specialisme       string
	Triage            string
	Vervoerder        string
	Geboortedatum     string
	OpnameAfdeling    string
	OpnameSpecialisme string
	Herkomst          string
	Ontslagbestemming string
	Vervallen         string
}

var requiredHeader = []string{
	"sehid", "sehmutid", "locatie", "afdeling", "aangemaakt", "binnenkomstdatum", "binnenkomsttijd", "triagetijd",
	"naarkamertijd", "eerstecontacttijd", "artsklaartijd", "gereedopnametijd", "vertrektijd", "eindtijd", "mutatieeindtijd",
	"mutatiestatus", "kamer", "bed",
	"ingangsklacht", "specialisme", "triage", "vervoerder", "geboortedatum", "opnameafdeling", "opnamespecialisme",
	"herkomst", "ontslagbestemming", "vervallen",
}

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
	r.Comma = ';'

	header, err := r.Read()
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(header, requiredHeader) {
		log.Printf("Want %v", requiredHeader)
		log.Printf("Got  %v", header)
		return errors.New("unsupported format")
	}
	for {
		csvRecords, err := readBatch(r, *batch)
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

		cfg := config.NewConfiguration()
		cfg.SetCredentials(*username, *password)

		u := uploader.Uploader{
			Configuration: cfg,
		}

		if *test {
			log.Println(buf)
			break
		}

		for {
			if err := u.UploadJSON(context.Background(), buf, true); err != nil {
				log.Println("Error", err)
				<-time.After(time.Second)
				continue
			}
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
		return fmt.Errorf("Ping failed: %s", res.Status)
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
			ID:                fields[0],
			MutatieID:         fields[1],
			Locatie:           fields[2],
			Afdeling:          fields[3],
			Aangemaakt:        fields[4],
			BinnenkomstDatum:  strings.TrimSuffix(fields[5], " 00:00:00"),
			BinnenkomstTijd:   fields[6],
			AanvangTriageTijd: fields[7],
			NaarKamerTijd:     fields[8],
			EersteContactTijd: fields[9],
			ArtsKlaarTijd:     fields[10],
			GereedOpnameTijd:  fields[11],
			VertrekTijd:       fields[12],
			EindTijd:          fields[13],
			MutatieEindTijd:   fields[14],
			MutatieStatus:     fields[15],
			Kamer:             fields[16],
			Bed:               fields[17],
			Ingangsklacht:     fields[18],
			Specialisme:       fields[19],
			Triage:            fields[20],
			Vervoerder:        fields[21],
			Geboortedatum:     fields[22],
			OpnameAfdeling:    fields[23],
			OpnameSpecialisme: fields[24],
			Herkomst:          fields[25],
			Ontslagbestemming: fields[26],
			Vervallen:         fields[27],
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
			Bezoeknummer:      id,
			MutatieID:         mutID,
			Locatie:           rec.Locatie,
			Afdeling:          "seh",
			Aangemeld:         aangemaakt,
			BinnenkomstDatum:  rec.BinnenkomstDatum,
			BinnenkomstTijd:   rec.BinnenkomstTijd,
			TriageTijd:        rec.AanvangTriageTijd,
			NaarKamerTijd:     rec.NaarKamerTijd,
			BijArtsTijd:       rec.EersteContactTijd,
			ArtsKlaarTijd:     rec.ArtsKlaarTijd,
			GereedOpnameTijd:  rec.GereedOpnameTijd,
			VertrekTijd:       rec.VertrekTijd,
			EindTijd:          rec.EindTijd,
			MutatieEindTijd:   rec.MutatieEindTijd,
			Mutatiestatus:     rec.MutatieStatus,
			Kamer:             rec.Kamer,
			Bed:               rec.Bed,
			Ingangsklacht:     rec.Ingangsklacht,
			Specialisme:       rec.Specialisme,
			Urgentie:          rec.Triage,
			Vervoerder:        rec.Vervoerder,
			Geboortedatum:     geb,
			OpnameAfdeling:    rec.OpnameAfdeling,
			OpnameSpecialisme: rec.OpnameSpecialisme,
			Herkomst:          rec.Herkomst,
			Ontslagbestemming: rec.Ontslagbestemming,
			Vervallen:         rec.Vervallen == "True",
		})
	}
	return res, nil
}
