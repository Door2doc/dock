DROP TABLE IF EXISTS correct;

CREATE TABLE correct (
  id                 SERIAL,
  sehid              INTEGER,
  sehmutid           INTEGER,
  locatie            TEXT,      -- locatie code
  aangemaakt         TIMESTAMP, -- aanmaak datum
  binnenkomstdatum   TEXT,      -- aankomst datum
  binnenkomsttijd    TEXT,      -- aankomst tijd
  aanvangtriagetijd  TEXT,      -- triage tijd
  naarkamertijd      TEXT,      -- tijd begin behandeling
  eerstecontacttijd  TEXT,      -- tijd patient gezien
  afdelinggebeldtijd TEXT,      -- tijd gebeld
  gereedopnametijd   TEXT,      -- tijd opname
  vertrektijd        TEXT,      -- tijd vertrek
  kamer              TEXT,      -- behandelkamer
  bed                TEXT,      -- bed
  ingangsklacht      TEXT,      -- ingangsklacht
  specialisme        TEXT,      -- specialisme
  triage             TEXT,      -- triage
  vervoerder         TEXT,      -- vervoer
  bestemming         TEXT,      -- bestemming
  geboortedatum      TIMESTAMP, -- geboortedatum
  opnameafdeling     TEXT,
  opnamespecialisme  TEXT,
  herkomst           TEXT,
  ontslagbestemming  TEXT
);

INSERT INTO correct(sehid, sehmutid, locatie, aangemaakt, binnenkomstdatum, binnenkomsttijd, aanvangtriagetijd,
                    naarkamertijd, eerstecontacttijd, afdelinggebeldtijd, gereedopnametijd, vertrektijd, kamer, bed,
                    ingangsklacht, specialisme, triage, vervoerder, bestemming, geboortedatum, opnameafdeling,
                    opnamespecialisme, herkomst, ontslagbestemming)
VALUES (328996, 1091568, 'A', '2017-07-13 13:00:00', '2017-07-13 00:00:00.000', '23:18', NULL, '23:18', '02:40', NULL,
        '02:06', '04:34', '', '', 'Pneumonie', '04', NULL, '2', 'A', '1977-07-24 12:00:00', NULL, NULL, NULL, NULL);

INSERT INTO correct(sehid, sehmutid, locatie, aangemaakt, binnenkomstdatum, binnenkomsttijd, aanvangtriagetijd,
                    naarkamertijd, eerstecontacttijd, afdelinggebeldtijd, gereedopnametijd, vertrektijd, kamer, bed,
                    ingangsklacht, specialisme, triage, vervoerder, bestemming, geboortedatum, opnameafdeling,
                    opnamespecialisme, herkomst, ontslagbestemming)
VALUES (1, 2, 'locatie', '2018-07-04 12:04:00', 'binnenkomstdatum', 'binnenkomsttijd', 'aanvangtriagetijd',
        'naarkamertijd', 'eerstecontacttijd', 'afdelinggebeldtijd', 'gereedopnametijd', 'vertrektijd', 'kamer', 'bed',
        'ingangsklacht', 'specialisme', 'triage', 'vervoerder', 'bestemming', '1977-07-24 12:00:00', 'opnameafdeling',
        'opnamespecialisme', 'herkomst', 'ontslagbestemming');

DROP TABLE IF EXISTS seh_sehmut;
DROP TABLE IF EXISTS seh_sehreg;
DROP TABLE IF EXISTS opname_opname;
DROP TABLE IF EXISTS patient_patient;

CREATE TABLE patient_patient (
  patientnr INTEGER PRIMARY KEY,
  gebdat    TIMESTAMP
);
CREATE TABLE opname_opname (
  plannr     INTEGER PRIMARY KEY,
  inschrtijd TEXT,
  afdeling   TEXT,
  specialism TEXT
);
CREATE TABLE seh_sehreg (
  sehid         INTEGER PRIMARY KEY,
  locatiecod    TEXT,
  aanksdatum    TEXT,
  aankstijd     TEXT,
  triagetijd    TEXT,
  artsbhtijd    TEXT,
  patgezt       TEXT,
  arbehetijd    TEXT,
  artsklaartijd TEXT,
  eindtijd      TEXT,
  specialism    TEXT,
  vvcode        TEXT,
  vervoertyp    TEXT,
  trianivcod    TEXT,
  bestemming    TEXT,
  datum         TIMESTAMP,
  patientnr     INTEGER REFERENCES patient_patient (patientnr),
  opnameid      INTEGER REFERENCES opname_opname (plannr),
  vervall       INTEGER
);
CREATE TABLE seh_sehmut (
  sehmutid   INTEGER,
  behkamerco TEXT,
  bednr      TEXT,
  sehid      INTEGER REFERENCES seh_sehreg (sehid)
);


DROP TABLE IF EXISTS vrlijst_antwview;
DROP TABLE IF EXISTS vrlijst_vragen;
DROP TABLE IF EXISTS vrlijst_keuzelst;

CREATE TABLE vrlijst_keuzelst (
  lijstcode TEXT PRIMARY KEY,
  code      TEXT,
  omschr    TEXT
);
CREATE TABLE vrlijst_vragen (
  vraagid    INTEGER PRIMARY KEY,
  keuzelijst TEXT REFERENCES vrlijst_keuzelst (lijstcode)
);
CREATE TABLE vrlijst_antwview (
  lijstid  TEXT,
  objectid TEXT,
  antwoord TEXT,
  realvrid INTEGER REFERENCES vrlijst_vragen (vraagid)
);