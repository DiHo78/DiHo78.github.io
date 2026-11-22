---
_edit_last: "13892234"
_oembed_5c29eee2f5c8fa509013fb4b55a755d2: '{{unknown}}'
_oembed_654a2f74467325dcea07d85f74d08c7d: '{{unknown}}'
_oembed_7856db278f42e0ee4457e6e3beceb6e1: '{{unknown}}'
_oembed_027171c5f54afe30d5d4a4efd052bbcd: '{{unknown}}'
_oembed_bcd87703d0abdd7719c1d09f00e8e7f9: '{{unknown}}'
_oembed_d4b36e6ab557fe6586d7f1eee5dc4df2: '{{unknown}}'
_oembed_fd02ec7cdc6735fc92685eb26e16cde1: '{{unknown}}'
_publicize_done_10277294: "1"
_publicize_done_external:
  twitter:
    "1791958": https://twitter.com/DirkHondong/status/662706507142660096
_publicize_job_id: "16604611379"
_wpas_done_1791958: "1"
author: feleryan
categories:
  - sql-server
date: "2015-11-06T18:53:48+00:00"
geo_public: "0"
guid: https://dirkhondong.wordpress.com/?p=274
parent_post_id: null
post_id: "274"
publicize_twitter_user: DirkHondong
tags:
  - sqlserver
  - t-sql
title: Es muss nicht immer DROP&hellip;CREATE sein
url: /2015/11/06/es-muss-nicht-immer-dropcreate-sein/

---
Hallo zusammen,

dieser kleine Beitrag schwirrte schon lange in meinem Kopf herum und nun mache ich mich endlich mal daran, diesen auch umzusetzen. Anlass ist der letzte [Blog Post](http://gds-business-intelligence.de/de/2015/11/05/drop-if-exists-in-sql-server-2016-ctp-3/) von einem unserer #sqlpass\_de MVPs und Leiter der Regionalgruppe Ruhrgebiet: Frank Geisler ([B](http://gds-business-intelligence.de/de/2015/) \| [T](https://twitter.com/frankgeisler)). Frank hatte das, in der CTP 3 des SQL Servers 2016, neue DROP…IF EXISTS kurz vorgestellt.

Und dann ist mir wieder eingefallen, wie oft ich schon sogenannte Update Skripte gesehen habe, wo Funktionen, Prozeduren oder Views “überarbeitet” wurden. Die fingen dann nämlich so an:

```
   1: DROP PROCEDURE blabla
```

```
   2:
```

```
   3: CREATE PROCEDURE blabla
```

```
   4: AS
```

```
   5: ....
(das blabla stand natürlich nicht so in den Skripten)
```

An und für sich mag der Ansatz ja ok sein, auch wenn der schon nicht so richtig elegant ist. Was ist, wenn das entsprechende Objekt gar nicht vorhanden ist? Dann kommt sofort

```
Msg 3701, Level 11, State 5, Line 1
```

```
Cannot drop the procedure 'blabla', because it does not exist or you do not have permission.
```

Wie man es richtig macht, hat Frank in seinem BlogPost ja aufgezeigt, nämlich:

```
   1: IF OBJECT_ID('blabla','P') IS NOT NULL
```

```
   2:     DROP PROCEDURE blabla
```

oder dann demnächst in SQL Server 2016:

```
   1: DROP PROCEDURE IF EXISTS dbo.blabla
```

```
   2: GO
```

```
   3: CREATE PROCEDURE dbo.blabla
```

```
   4: AS
```

Nun aber Folgendes: Was ist denn, wenn man tatsächlich eine Datenbank hat, wo granular Berechtigungen vergeben worden sind und es ggf. kein dediziertes Schema gibt, auf dem ich die entsprechenden Berechtigungen vergeben kann? Wo also zum Beispiel ein bestimmtes Servicekonto nur eine Prozedur ausführen oder eine View abrufen soll? Mit dem DROP ist das Objekt weg und damit auch die granulare Berechtigung auf eben dieses.

Von daher hab ich nun schon einige Male folgende Idee mit auf den Weg gegeben (und es ist nichts Neues): Prüft, ob das entsprechende Objekt da ist oder nicht. Wenn nicht, dann einach einen Dummy anlegen und nachgelagert dann das neue Objekt mit ALTER begegnen. Wenn es schon da ist, dann greift direkt Euer ALTER Befehl. Also quasi so:

```
   1: IF OBJECT_ID('dbo.blabla','P') IS  NULL
```

```
   2:     EXEC ('CREATE PROCEDURE blabla
```

```
   3:     AS
```

```
   4:     SET NOCOUNT ON;')
```

```
   5: GO
```

```
   6: ALTER PROCEDURE blabla
```

```
   7: AS
```

```
   8: BEGIN
```

```
   9: PRINT 'So ist es doch besser....'
```

```
  10: END
```

Dies hat den Charme, dass Ihr nicht bereits gesetzte Berechtigungen verliert. Erspart in mancher heißen Testphase das eine oder andere Telefonat mit dem DBA eures Vertrauens.
