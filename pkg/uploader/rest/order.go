package rest

import (
	"time"

	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
)

type OrderRecord struct {
	Bezoeknummer int        `json:"bezoek_id"`
	Ordernummer  int        `json:"order_id"`
	Start        *time.Time `json:"dt_start"`
	Eind         *time.Time `json:"dt_eind"`
	Status       string     `json:"code_status"`
	Module       string     `json:"code_module"`
	Specialisme  string     `json:"code_specialisme"`
}

func fix(t *time.Time, loc *time.Location) *time.Time {
	if t == nil {
		return nil
	}

	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	res := time.Date(year, month, day, hour, min, sec, 0, loc)
	return &res
}

func (r *OrderRecord) fromRadiologie(order db.RadiologieOrder, loc *time.Location) error {
	r.Bezoeknummer = order.Bezoeknummer
	r.Ordernummer = order.Ordernummer
	r.Start = fix(order.Start, loc)
	r.Eind = fix(order.Eind, loc)
	r.Status = order.Status
	r.Module = order.Module
	return nil
}

func (r *OrderRecord) fromLab(order db.LabOrder, loc *time.Location) error {
	r.Bezoeknummer = order.Bezoeknummer
	r.Ordernummer = order.Ordernummer
	r.Start = fix(order.Start, loc)
	r.Eind = fix(order.Eind, loc)
	r.Status = order.Status
	return nil
}

func (r *OrderRecord) fromConsult(order db.ConsultOrder, loc *time.Location) error {
	r.Bezoeknummer = order.Bezoeknummer
	r.Ordernummer = order.Ordernummer
	r.Start = fix(order.Start, loc)
	r.Eind = fix(order.Eind, loc)
	r.Status = order.Status
	r.Specialisme = order.Specialisme
	return nil
}

func RadiologieRecordsFromDB(rs db.RadiologieOrders, loc *time.Location) ([]OrderRecord, error) {
	res := make([]OrderRecord, len(rs))
	for i := range rs {
		if err := res[i].fromRadiologie(rs[i], loc); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func LabRecordsFromDB(rs db.LabOrders, loc *time.Location) ([]OrderRecord, error) {
	res := make([]OrderRecord, len(rs))
	for i := range rs {
		if err := res[i].fromLab(rs[i], loc); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func ConsultRecordsFromDB(rs db.ConsultOrders, loc *time.Location) ([]OrderRecord, error) {
	res := make([]OrderRecord, len(rs))
	for i := range rs {
		if err := res[i].fromConsult(rs[i], loc); err != nil {
			return nil, err
		}
	}
	return res, nil
}
