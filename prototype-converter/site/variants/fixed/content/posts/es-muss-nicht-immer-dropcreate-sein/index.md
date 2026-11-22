---
title: "Es muss nicht immer DROP…CREATE sein"
date: "2015-11-06T18:53:48+00:00"
url: "/es-muss-nicht-immer-dropcreate-sein/"
lang: "de"
categories: ["SQL Server"]
tags: ["SQLServer", "T-SQL"]
---
Hallo zusammen,

dieser kleine Beitrag schwirrte schon lange in meinem Kopf herum und nun mache ich mich endlich mal daran, diesen auch umzusetzen. Anlass ist der letzte [Blog Post](http://gds-business-intelligence.de/de/2015/11/05/drop-if-exists-in-sql-server-2016-ctp-3/) von einem unserer #sqlpass\_de MVPs und Leiter der Regionalgruppe Ruhrgebiet: Frank Geisler ([B](http://gds-business-intelligence.de/de/2015/) \| [T](https://twitter.com/frankgeisler)). Frank hatte das, in der CTP 3 des SQL Servers 2016, neue DROP…IF EXISTS kurz vorgestellt.

Und dann ist mir wieder eingefallen, wie oft ich schon sogenannte Update Skripte gesehen habe, wo Funktionen, Prozeduren oder Views “überarbeitet” wurden. Die fingen dann nämlich so an:

```sql
DROP PROCEDURE blabla

CREATE PROCEDURE blabla
AS
....
(das blabla stand natürlich nicht so in den Skripten)
```

An und für sich mag der Ansatz ja ok sein, auch wenn der schon nicht so richtig elegant ist. Was ist, wenn das entsprechende Objekt gar nicht vorhanden ist? Dann kommt sofort

```text
Msg 3701, Level 11, State 5, Line 1
Cannot drop the procedure 'blabla', because it does not exist or you do not have permission.
```

Wie man es richtig macht, hat Frank in seinem BlogPost ja aufgezeigt, nämlich:

```sql
IF OBJECT_ID('blabla','P') IS NOT NULL
    DROP PROCEDURE blabla
```

oder dann demnächst in SQL Server 2016:

```sql
DROP PROCEDURE IF EXISTS dbo.blabla
GO
CREATE PROCEDURE dbo.blabla
AS
```

Nun aber Folgendes: Was ist denn, wenn man tatsächlich eine Datenbank hat, wo granular Berechtigungen vergeben worden sind und es ggf. kein dediziertes Schema gibt, auf dem ich die entsprechenden Berechtigungen vergeben kann? Wo also zum Beispiel ein bestimmtes Servicekonto nur eine Prozedur ausführen oder eine View abrufen soll? Mit dem DROP ist das Objekt weg und damit auch die granulare Berechtigung auf eben dieses.

Von daher hab ich nun schon einige Male folgende Idee mit auf den Weg gegeben (und es ist nichts Neues): Prüft, ob das entsprechende Objekt da ist oder nicht. Wenn nicht, dann einach einen Dummy anlegen und nachgelagert dann das neue Objekt mit ALTER begegnen. Wenn es schon da ist, dann greift direkt Euer ALTER Befehl. Also quasi so:

```sql
IF OBJECT_ID('dbo.blabla','P') IS  NULL
    EXEC ('CREATE PROCEDURE blabla
    AS
    SET NOCOUNT ON;')
GO
ALTER PROCEDURE blabla
AS
BEGIN
PRINT 'So ist es doch besser....'
END
```

Dies hat den Charme, dass Ihr nicht bereits gesetzte Berechtigungen verliert. Erspart in mancher heißen Testphase das eine oder andere Telefonat mit dem DBA eures Vertrauens.
