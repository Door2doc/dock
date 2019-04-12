-- 12 april 2019
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
SELECT mut.sehid      AS sehid,
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
       klacht.omschr  AS ingangsklacht,
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
         LEFT OUTER JOIN klacht ON klacht.sehid = reg.sehid
WHERE reg.vervall = 0
  AND reg.datum >= getdate() - 1
ORDER BY mut.sehmutid DESC,
         reg.sehid DESC;

