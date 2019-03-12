CREATE TABLE correct (
  sehid       INTEGER,
  sehmutid    INTEGER,
  locatiecod  TEXT,      -- locatie code
  aanmaakdat  TEXT,      -- aanmaak datum
  aanmaaktijd TEXT,      -- aanmaak tijd
  aanksdatum  TEXT,      -- aankomst datum
  aankstijd   TEXT,      -- aankomst tijd
  triadatum   TEXT,      -- triage datum
  triatijd    TEXT,      -- triage tijd
  artsbhtijd  TEXT,      -- tijd begin behandeling
  patgezt     TEXT,      -- tijd patient gezien
  gebeld      TEXT,      -- tijd gebeld
  inschrtijd  TEXT,      -- tijd opname
  arbehetijd  TEXT,      -- tijd vertrek
  behkamerco  TEXT,      -- behandelkamer
  bednr       TEXT,      -- bed
  klacht      TEXT,      -- ingangsklacht
  specialism  TEXT,      -- specialisme
  trianivcod  TEXT,      -- triage
  vervoertyp  TEXT,      -- vervoer
  bestemming  TEXT,      -- bestemming
  gebdat      TIMESTAMP, -- geboortedatum
  opnameafd   TEXT,
  opnamespec  TEXT
);

INSERT INTO correct(sehid, sehmutid, locatiecod, aanmaakdat, aanmaaktijd, aanksdatum, aankstijd, triadatum, triatijd, artsbhtijd, patgezt, gebeld, inschrtijd, arbehetijd, behkamerco, bednr, klacht, specialism, trianivcod, vervoertyp, bestemming, gebdat, opnameafd, opnamespec)
VALUES (328996, 1091568, 'A', NULL, NULL, '2017-07-13 00:00:00.000', '23:18', NULL, NULL, '23:18', '02:40', NULL, '02:06', '04:34', '', '', 'Pneumonie', '04', NULL, '2', 'A', '1977-07-24 12:00:00', NULL, NULL);
