DROP TABLE IF EXISTS correct;

CREATE TABLE correct (
    id                SERIAL,
    sehid             INTEGER,
    sehmutid          INTEGER,
    locatie           TEXT,      -- locatie code
    afdeling          TEXT,      -- afdeling code
    aangemaakt        TIMESTAMP, -- aanmaak datum
    binnenkomstdatum  TEXT,      -- aankomst datum
    binnenkomsttijd   TEXT,      -- aankomst tijd
    triagetijd        TEXT,      -- triage tijd
    naarkamertijd     TEXT,      -- tijd begin behandeling
    eerstecontacttijd TEXT,      -- tijd patient gezien
    artsklaartijd     TEXT,      -- tijd arts klaar
    gereedopnametijd  TEXT,      -- tijd opname
    vertrektijd       TEXT,      -- tijd vertrek
    eindtijd          TEXT,      -- eind tijd
    kamer             TEXT,      -- behandelkamer
    bed               TEXT,      -- bed
    ingangsklacht     TEXT,      -- ingangsklacht
    specialisme       TEXT,      -- specialisme
    triage            TEXT,      -- triage
    vervoerder        TEXT,      -- vervoer
    bestemming        TEXT,      -- bestemming
    geboortedatum     TIMESTAMP, -- geboortedatum
    opnameafdeling    TEXT,
    opnamespecialisme TEXT,
    herkomst          TEXT,
    ontslagbestemming TEXT,
    vervallen         INTEGER,
    mutatieeindtijd   TEXT,
    mutatiestatus     TEXT
);

INSERT INTO correct(sehid, sehmutid, locatie, afdeling, aangemaakt, binnenkomstdatum, binnenkomsttijd,
                    triagetijd,
                    naarkamertijd, eerstecontacttijd, artsklaartijd, gereedopnametijd, vertrektijd, eindtijd, kamer,
                    bed,
                    ingangsklacht, specialisme, triage, vervoerder, bestemming, geboortedatum, opnameafdeling,
                    opnamespecialisme, herkomst, ontslagbestemming, vervallen)
VALUES (328996, 1091568, 'A', 'seh', '2017-07-13 13:00:00', '2017-07-13', '23:18', NULL, '23:18', '02:40', NULL,
        '02:06', '04:34', '04:34', '', '', 'Pneumonie', '04', NULL, '2', 'A', '1977-07-24 12:00:00', NULL, NULL, NULL,
        NULL, 0);

INSERT INTO correct(sehid, sehmutid, locatie, afdeling, aangemaakt, binnenkomstdatum, binnenkomsttijd,
                    triagetijd,
                    naarkamertijd, eerstecontacttijd, artsklaartijd, gereedopnametijd, vertrektijd, eindtijd, kamer,
                    bed,
                    ingangsklacht, specialisme, triage, vervoerder, bestemming, geboortedatum, opnameafdeling,
                    opnamespecialisme, herkomst, ontslagbestemming, vervallen)
VALUES (1, 2, 'locatie', 'seh', '2018-07-04 12:04:00', 'binnenkomstdatum', 'binnenkomsttijd', 'aanvangtriagetijd',
        'naarkamertijd', 'eerstecontacttijd', 'artsklaartijd', 'gereedopnametijd', 'vertrektijd', 'eindtijd', 'kamer',
        'bed',
        'ingangsklacht', 'specialisme', 'triage', 'vervoerder', 'bestemming', '1977-07-24 12:00:00', 'opnameafdeling',
        'opnamespecialisme', 'herkomst', 'ontslagbestemming', 0);

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

DROP TABLE IF EXISTS correct_radiologie;
DROP TABLE IF EXISTS correct_lab;
DROP TABLE IF EXISTS correct_consult;

CREATE TABLE correct_radiologie (
    sehid          TEXT,
    ordernr        TEXT,
    status         TEXT,
    startdatumtijd TIMESTAMP,
    einddatumtijd  TIMESTAMP,
    module         TEXT
);
CREATE TABLE correct_lab (
    sehid          TEXT,
    ordernr        TEXT,
    status         TEXT,
    startdatumtijd TIMESTAMP,
    einddatumtijd  TIMESTAMP
);
CREATE TABLE correct_consult (
    sehid          TEXT,
    ordernr        TEXT,
    status         TEXT,
    startdatumtijd TIMESTAMP,
    einddatumtijd  TIMESTAMP,
    specialisme    TEXT
);
