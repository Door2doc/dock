WITH vragen AS
         (
             SELECT DISTINCT RIGHT(vav.objectid, 10) AS sehid,
                             vkl1.omschr

             FROM vrlijst_antwview vav
                      LEFT OUTER JOIN vrlijst_vragen vvr ON vav.realvrid = vvr.vraagid
                      LEFT OUTER JOIN vrlijst_keuzelst vkl1 ON vvr.keuzelijst = vkl1.lijstcode
                 AND SUBSTRING(vav.antwoord, 1, 10) = vkl1.code

             WHERE vkl1.lijstcode = 'CS00000991'
               AND vav.lijstid = 'CS00006105'
               AND vav.datum > GETDATE() - 2
         )

SELECT mut.sehid               AS sehid,
       mut.sehmutid            AS sehmutid,
       reg.locatiecod          AS locatie,
       'seh'                   AS afdeling,
       reg.datum + reg.regtijd AS aangemaakt,
       reg.aanksdatum          AS binnenkomstdatum,
       reg.aankstijd           AS binnenkomsttijd,
       reg.triagetijd          AS aanvangtriagetijd,
       reg.artsbhtijd          AS naarkamertijd,
       reg.patgezt             AS eerstecontacttijd,
       reg.artsklaartijd       AS artsklaartijd,
       oo.inschrtijd           AS gereedopnametijd,
       reg.arbehetijd          AS vertrektijd,
       reg.eindtijd            AS eindtijd,
       mut.behkamerco          AS kamer,
       mut.bednr               AS bed,
       vr.omschr               AS ingangsklacht,
       reg.specialism          AS specialisme,
       reg.trianivcod          AS triage,
       reg.vvcode              AS vervoerder,
       pp.gebdat               AS geboortedatum,
       oo.afdeling             AS opnameafdeling,
       oo.specialism           AS opnamespecialisme,
       reg.vervoertyp          AS herkomst,
       reg.bestemming          AS ontslagbestemming,
       reg.vervall             AS vervallen
FROM seh_sehmut mut
         WITH (NOLOCK)
         LEFT JOIN seh_sehreg reg ON mut.sehid = reg.sehid
         LEFT OUTER JOIN patient_patient pp ON reg.patientnr = pp.patientnr
         LEFT OUTER JOIN opname_opname oo ON reg.opnameid = oo.plannr
         LEFT OUTER JOIN vragen vr ON reg.sehid = vr.sehid
WHERE reg.datum >= getdate() - 2
ORDER BY mut.sehmutid DESC,
         reg.sehid DESC;
