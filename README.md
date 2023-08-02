# Extranet Scraper

## Zweck
Extrahiert von der Webseite [Extranet Pastoralraum Aarau][Extranet] alle Dokumente und speichert sie in einer entsprechenden Dateistruktur ab.

## Benutzung

    % scraper --help
    Scraps the extranet of pastoralraum Aarau and downloads all the files
        -download string
    	    folder where to download all the files to (default "./downloads/")
    -password string
            password of extranet
    -top string
            top directory where to start scraping, like 'home', 'documents', 'personal', ... (default "home")
    -user string
            user name for extranet

## Copyright
Kreiskirche Aarau, 2023

Kontakt: <markus.greiner@pfarrei-aarau.ch>


[Extranet]: https://extranet.pastoralraum-aarau.ch/de/home