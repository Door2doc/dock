WITH klacht AS
       (
         SELECT DISTINCT RIGHT(vav.objectid, 10) AS sehid,
                         vkl1.omschr
         FROM vrlijst_antwview vav
                LEFT OUTER JOIN vrlijst_vragen vvr ON vav.realvrid = vvr.vraagid
                LEFT OUTER JOIN vrlijst_keuzelst vkl1 ON vvr.keuzelijst = vkl1.lijstcode
           AND SUBSTRING(vav.antwoord, 1, 10) = vkl1.code
         WHERE vkl1.lijstcode = 'CS00000991'
           AND vav.lijstid = 'CS00006105'
       )
SELECT sehmut.sehmutid,
       seh.locatiecod,
       seh.sehid                                                   AS "Bezoeknummer SEH",
       klacht.omschr                                               AS ingangsklacht,
       seh.aanksdatum + seh.aankstijd                              AS datumtijdbinnen,
       seh.triagetijd                                              AS "TijdstipTrage",
       seh.artsbhtijd                                              AS begindatumtijdbehandeling,
       seh.artsklaartijd                                           AS patientgeziendatumtijdarts,
       oo.inschrtijd                                               AS datumtijdopname_opname,
       NULL                                                        AS "Tijdstip afdeling gebeld",
       seh.eindtijd                                                AS datumtijdvertrek,
       seh.specialism                                              AS specialisme,
       sehmut.behkamerco                                           AS kamernummer,
       sehmut.bednr                                                AS bednummer,
       seh.vvcode                                                  AS vervoerder,
       seh.vervoertyp                                              AS herkomst_code,
       shh.omschrijvi                                              AS herkomst_oms,
       seh.trianivcod                                              AS triageniveaucode,
       seh.bestemming                                              AS bestemming,
       shb.omschrijvi                                              AS bestemming_o,
       oo.afdeling                                                 AS opnameafdeling,
       oo.specialism                                               AS opnamespecialisme,
       cast(DATEDIFF(year, pat.gebdat, seh.datum) / 10 AS INTEGER) AS leeftijdsgroep,
       NULL                                                        AS "Diagnostische orders"
FROM seh_sehreg seh
       JOIN seh_sehmut sehmut ON seh.sehid = sehmut.sehid
       LEFT OUTER JOIN klacht ON seh.sehid = klacht.sehid
       LEFT OUTER JOIN patient_patient pat ON
  seh.patientnr = pat.patientnr
       LEFT OUTER JOIN opname_opname oo
                       ON seh.opnameid = oo.plannr
       LEFT OUTER JOIN seh_sehherk shh ON
  seh.vervoertyp = shh.herkmtcode
       LEFT OUTER JOIN seh_shbestem shb ON
  seh.bestemming = shb.code
WHERE seh.vervall = 0
  AND year(seh.datum) >= 2017
ORDER BY seh.sehid ASC,
         sehmut.sehmutid