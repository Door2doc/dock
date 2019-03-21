USE upload;

DROP TABLE IF EXISTS correct;

CREATE TABLE correct (
  id                 INT NOT NULL PRIMARY KEY IDENTITY,
  sehid              INTEGER,
  sehmutid           INTEGER,
  locatie            TEXT,     -- locatie code
  aangemaakt         DATETIME, -- aanmaak datum
  binnenkomstdatum   TEXT,     -- aankomst datum
  binnenkomsttijd    TIME,     -- aankomst tijd
  aanvangtriagetijd  TEXT,     -- triage tijd
  naarkamertijd      TEXT,     -- tijd begin behandeling
  eerstecontacttijd  TEXT,     -- tijd patient gezien
  afdelinggebeldtijd TEXT,     -- tijd gebeld
  gereedopnametijd   TEXT,     -- tijd opname
  vertrektijd        TEXT,     -- tijd vertrek
  kamer              TEXT,     -- behandelkamer
  bed                TEXT,     -- bed
  ingangsklacht      TEXT,     -- ingangsklacht
  specialisme        TEXT,     -- specialisme
  triage             TEXT,     -- triage
  vervoerder         TEXT,     -- vervoer
  bestemming         TEXT,     -- bestemming
  geboortedatum      DATETIME, -- geboortedatum
  opnameafdeling     TEXT,
  opnamespecialisme  TEXT,
  herkomst           TEXT,
  ontslagbestemming  TEXT
);

INSERT INTO correct(sehid, sehmutid, locatie, aangemaakt, binnenkomstdatum, binnenkomsttijd, aanvangtriagetijd,
                    naarkamertijd, eerstecontacttijd, afdelinggebeldtijd, gereedopnametijd, vertrektijd, kamer, bed,
                    ingangsklacht, specialisme, triage, vervoerder, bestemming, geboortedatum, opnameafdeling,
                    opnamespecialisme, herkomst, ontslagbestemming)
VALUES (328996, 1091568, 'A', '2017-07-13 13:00:00', '2017-07-13', '23:18', NULL, '23:18', '02:40', NULL,
        '02:06', '04:34', '', '', 'Pneumonie', '04', NULL, '2', 'A', '1977-07-24 12:00:00', NULL, NULL, NULL, NULL);



