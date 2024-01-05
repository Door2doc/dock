---
title: Handleiding Dock - de door2doc Upload Service
author: door2doc 
---

# Overzicht

Dit document beschrijft de installatie van Dock -- een Windows service die op regelmatige basis 
data uitleest uit het ZIS en deze doorzet naar de door2doc Cloud omgeving.  

# Voorbereidingen

Voordat u aan de installatie kunt beginnen, dient over het volgende te beschikken:

## Server 
Een server waarop Dock geïnstalleerd kan worden. Dit hoeft geen dedicated server te zijn. Deze server
moet connectie kunnen maken met de ZIS database, en moet connectie kunnen maken met `https://integration.door2doc.net` 
(al dan niet via een proxy).

Op deze server dient een browser geïnstalleerd te zijn om de web interface te kunnen gebruiken. 

## Database gegevens
Om verbinding te maken met de database zijn er connectie gegevens nodig. Gebruikersnaam en wachtwoord zijn niet verplicht;
indien afwezig, wordt SQL Server Integrated Security gebruikt. 

## Door2doc Credentials
Voor de upload naar door2doc heeft de service een door2doc gebruikersnaam en wachtwoord nodig. Deze krijgt u door ons geleverd. 

## Queries
Dock voert op regelmatige basis queries uit om de data uit uw database te halen. Deze queries worden door
door2doc en uw dienst informatievoorziening opgesteld en aangeleverd. 

## Software
U kunt de laatste versie van deze software downloaden van 

`https://github.com/Door2doc/dock/releases`. 

De software wordt geleverd als `.exe` bestand.  

# Installatie

## Installatie 

Installatie vindt plaats door een dubbelklik op het `.exe` bestand:

![Installatie van Dock](install-software-1.png){height=50%}\   

Dit moet in enkele seconden afgerond zijn. U kunt de installatie verifiëren via Administrative Tools > Services. 
Hier is nu de service Door2doc Upload Service te zien in de status Running. 

![Services overzicht](running-service.png){height=50%}\ 

# Configuratie

De service dient nu geconfigureerd te worden. Dit gebeurt via een web interface op `http://localhost:17226`.

![Configuratiescherm](config-scherm.png){height=50%}\ 

Vul hier de gegevens in die u aan het begin verzameld hebt. Wanneer alle gegevens correct zijn ingevoerd, zal 
de service de volgende status geven. 

![Succes](configured-ok.png){height=50%}\ 

# Beheer

## Starten/Stoppen service

De service kan gepauzeerd worden via Administrative Tools > Services. 

## Logs 

Foutmeldingen worden naar de Windows event log gestuurd, en kunnen worden gemonitord via de Windows Event Viewer:

![Event viewer](eventviewer.png){height=50%}\

## Configuratie

Het configuratiebestand is te vinden in `C:\ProgramData\Door2doc`.
