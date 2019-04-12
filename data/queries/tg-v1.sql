-- 5 april 2019
SELECT TOP 100 mut.sehid      AS sehid,
               mut.sehmutid   AS sehmutid,
               reg.locatiecod AS locatie,
               reg.datum      AS aangemaakt,
               reg.aanksdatum AS binnenkomstdatum,
               reg.aankstijd  AS binnenkomsttijd,
               reg.triagetijd AS aanvangtriagetijd,
               reg.artsbhtijd AS naarkamertijd,
               reg.patgezt    AS eerstecontacttijd,
               NULL           AS afdelinggebeldtijd,
               oo.inschrtijd  AS gereedopnametijd,
               reg.arbehetijd AS vertrektijd,
               mut.behkamerco AS kamer,
               mut.bednr      AS bed,
               reg.klacht  AS ingangsklacht,
               reg.specialism AS specialisme,
               reg.trianivcod AS triage,
               reg.vvcode     AS vervoerder,
               pp.gebdat      AS geboortedatum,
               oo.afdeling    AS opnameafdeling,
               oo.specialism  AS opnamespecialisme,
               reg.vervoertyp AS herkomst,
               reg.bestemming AS ontslagbestemming
FROM seh_sehmut mut
         WITH (NOLOCK)
         LEFT JOIN seh_sehreg reg ON mut.sehid = reg.sehid
         LEFT OUTER JOIN patient_patient pp ON reg.patientnr = pp.patientnr
         LEFT OUTER JOIN opname_opname oo ON reg.opnameid = oo.plannr
WHERE reg.vervall = 0
  AND reg.datum >= getdate() - 1
ORDER BY mut.sehmutid DESC,
         reg.sehid DESC;

