SELECT r.sehid,
       '''' + CAST(r.klacht AS VARCHAR(100)) + ''''      AS redenvanbezoek,
       r.aanksdatum                                      AS datumbinnen,
       r.aankstijd                                       AS tijdbinnen,
       r.artsbhtijd                                      AS tijdbeginbehandeling,
       r.patgezt                                         AS tijdpatientgezien,
       o.inschrtijd                                      AS tijdopname,
       r.arbehetijd                                      AS tijdvertrek,
       r.specialism                                      AS specialisme,
       m.behkamerco                                      AS kamernummer,
       m.sehmutid                                        AS kamernrmutid,
       r.vervoertyp                                      AS herkomst,
       r.trianivcod                                      AS triagecode,
       '''' + CAST(r.bestemming AS VARCHAR(100)) + ''''  AS bestemming,
       r.locatiecod                                      AS locatiecode,
       e.vervoecode                                      AS vervoerder,
       (YEAR(GETDATE()) - DATEPART(YEAR, p.gebdat)) / 10 AS gebjaar,
       ISNULL(m.bednr, n '')                             AS bednummer,
       r.isrampreg                                       AS ramppatient
FROM seh_sehreg AS r
       WITH (NOLOCK)
       LEFT OUTER JOIN seh_seh_ext AS e ON r.sehid = e.objectid1
       LEFT OUTER JOIN opname_opname AS o ON r.opnameid = o.plannr
       LEFT OUTER JOIN patient_patient AS p ON p.patientnr = r.patientnr
       LEFT OUTER JOIN seh_sehmut AS m ON r.sehid = m.sehid
WHERE (r.locatiecod = 'A') AND (r.vervall <> 1) AND (r.ontslagdat > GETDATE() - 2)
   OR (r.locatiecod = 'A') AND (r.vervall <> 1) AND (r.ogevaldatu > GETDATE() - 2)