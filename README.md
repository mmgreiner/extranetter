# Extranet Scraper

## Zweck
Extrahiert von der Webseite [Extranet Pastoralraum Aarau][Extranet] alle Dokumente und speichert sie in einer entsprechenden Dateistruktur ab.

## Benutzung

        % ./extranetter -help
        Downloads files from the extranet of Pastoralraum Aarau
        Copyright Pastoralraum Aarau, contact: markus.greiner@pfarrei-aarau.ch
        Options:
        -debug
                print debugging messages
        -download string
                folder where to download all the files to (default "./downloads/")
        -folder string
                folder inside top-folder which to download
        -password string
                password of extranet
        -silent
                print all files, but does not download them
        -top string
                top folder where to start scraping, like 'home', 'personal', 'sitzungen', 'archiv_sitzungen', 'budget', ... (default "home")
        -user string
                user name for extranet, example markus.greiner@pfarrei-aarau.ch

Typische Anwendungen:

- Lade alle Dokumente der Sitzung vom 24.8.2024: \
  `extranetter -user markus.greiner@pfarrei-aarau.ch -password XXX -folder 2023_08_24` \
  Der Name des Folders sieht man im [Extranet], wenn man auf den Link klickt. Alle Dokumente werden auf "downloads" runtergeladen.

- Lade alle Dokumente, die Personal betreffen: \
  `extranetter -user markus.greiner@pfarrei-aarau.ch -password XXX -top personal` 

- Lade alle Dokumente des [Extranet]s herunter in den Folder "Meine Dokumente". \
  `extranetter -user markus.greiner@pfarrei-aarau.ch -password XXX -download "Meine Dokumente"`


## Copyright
Kreiskirche Aarau, 2023

Kontakt: <markus.greiner@pfarrei-aarau.ch>


[Extranet]: https://extranet.pastoralraum-aarau.ch/de/home

## Installation

TBD

## Entwicklerhinweise

Dieses Werkzeug wurde mit [Go](https://go.dev) entwickelt. Der Web Scraper kommt von [Colly](http://go-colly.org).

Lizenz gemäss [MIT](LICENSE).
