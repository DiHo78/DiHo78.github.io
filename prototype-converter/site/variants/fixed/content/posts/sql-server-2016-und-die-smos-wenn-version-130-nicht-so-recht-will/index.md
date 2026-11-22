---
title: "SQL Server 2016 und die SMOs.... (wenn Version 130 nicht so recht will...)"
date: "2018-04-10T07:52:24+00:00"
url: "/sql-server-2016-und-die-smos-wenn-version-130-nicht-so-recht-will/"
lang: "de"
categories: ["Azure", "PowerShell", "SQL Server", "SQL Server Administration", "SSMS"]
tags: ["Azure VM", "Backup", "SMO"]
---
Hallo alle miteinander,

ich hab mich vor Kurzem ein wenig mit den SQL Server SMOs auseinandergesetzt. Die SMOs sind ja dazu gedacht einen SQL Server programmatisch zu managen. Weiterführende Infos finden sich [hier](https://docs.microsoft.com/de-de/sql/relational-databases/server-management-objects-smo/sql-server-management-objects-smo-programming-guide).

Nun, warum hab ich mich überhaupt näher mit den SMOs beschäftigt? Die Backup-Software mit der ich arbeite nutzt im Hintergrund für einige Operationen, wie zum Beispiel das Browsen einer Instanz und dem Auflisten der diversen Datenbanken, eben diese SMOs.

Bis dato hatte ich auch gar keine Probleme gehabt, bis ich auf einem SQL Server 2016 eine Datenbank offline gesetzt habe. Ein Log oder Fullbackup über die Software lief zwar weiterhin, aber um einiges länger als ich es gewohnt war. Zudem bekam ich auf einmal im SQL Server Log lauter Fehlermeldungen mit Fehlercode 18456 und der Nachricht "Failed to open explicitly specified database". Auf einem SQL Server 2014 jedoch lies sich dieses Verhalten nicht reproduzieren.

Das war der Grund meiner kleinen Recherche. Zuerst hatte ich den Hersteller der Backup Software angeschrieben.

Nach kurzem hin und her kam dann vom Hersteller die Aussage, dass die Entwicklungsabteilung das Problem nachstellen konnte. Darüber hinaus wurde aber auch gesagt: das Entwicklungsteam glaubt, dass es ein Microsoft Problem ist.

Ich hab dann einen PowerShell Dreizeiler erhalten, womit man das Verhalten nachstellen konnte:

```powershell
[void][reflection.aSSEMBLY]::LoadWithPartialName("Microsoft.SQLServer.Smo" )
$server = New-Object -TypeName Microsoft.SQLServer.Management.Smo.Server "<INSTANCE Name>"
$server.Databases.Item("<database name>")
```

[Gist 1c0053f](https://gist.github.com/DiHo78/1c0053ff6bfefc27fab0dd219624fc35)

Den oben stehenden Code habe ich dann kurz in Hinblick auf Instanzname und Datenbankname angepasst und direkt auf meinem SQL Server 2016 laufen lassen. Das Ergebnis waren lauter login failed Meldungen im SQLServer Log und keine Reaktion mehr von der PowerShell Session.

Dann dachte ich mir: ok, ggf. hat der Dirk eine "strubbelige" SQL Server Installation und ggf. irgendetwas ausgelassen. Also war mein nächster Schritt schnell einen SQL Server 2016 aus dem Azure Marketplace zu schnappen und bereitzustellen.

Sobald die Maschine bereit stand hab ich eine Datenbank namens OfflineDB angelegt und das obenstehende Script ausgeführt. Zu meiner Verwunderung lief es problemlos durch. Die Erklärung hatte ich aber auch schnell parat. Der Unterschied zu der Azure Marketplace VM und meinem Server war, dass die Azure VM inklusive einem SQL Server Management Studio daher kommt, in dem Fall ein SSMS 17.4. Somit kommen auch die Assemblies vom SQL Server 2017 mit auf das System:

![AssemblyPfad](assemblypfad.jpg)

Dann hab ich noch schnell überlegt, wie ich in einer PowerShell Session jetzt explizit ein Assembly ansprechen kann. Dies geht ganz einfach mit Add-Type -Path

Somit hab ich dann folgende Code Schnipsel ausprobiert:

```powershell
add-type -Path 'D:\Program Files\Microsoft SQL Server\130\SDK\Assemblies\Microsoft.SqlServer.Smo.dll'
$server = New-Object -TypeName Microsoft.SQLServer.Management.Smo.Server "smotest"
$server.Databases.Item("OfflineDB")
```

```powershell
add-type -Path 'D:\Program Files\Microsoft SQL Server\140\SDK\Assemblies\Microsoft.SqlServer.Smo.dll'
$server = New-Object -TypeName Microsoft.SQLServer.Management.Smo.Server "smotest"
$server.Databases.Item("OfflineDB")
```

[Gist 802b9d0](https://gist.github.com/DiHo78/802b9d03442d9743cc097a7c8272d401)

In dem oberen Beispiel spreche ich das 2016er Assembly an, im unteren das 2017er. Auch nochmal hier links (2016 / 130) und rechts (2017 / 140) zu sehen:

![SMOTest](smotest.jpg)

Im linken Fenster passierte dann auch nichts mehr und ich hab auch wieder "brav" meine failed logins produziert:

![LoginFailed](loginfailed.jpg)

Meine Interimslösung für meinen SQL Server 2016 war jetzt das SQL Server Management Studio 17.x zu installieren. Somit hab ich dann die aktuellen 2017er Assemblies auf dem Server gehabt.

Diese werden nun auch von der BackupSoftware angesprochen und jetzt läuft das Backup der Instanz mit einer offline Datenbank wieder problemlos.

Dies ist natürlich nicht der Weisheit letzter Schluß, denn eigentlich bin ich kein Fan davon, dass SSMS auf einem Server zu installieren und dies ist jetzt erst einmal auch nur ein Workaround. Microsoft hat jedoch etwas in Hinblick auf die Verteilung der SMOs geändert. Diese finden sich ab der Version 2017 nicht mehr in dem SQL Server Feature Pack wieder. Stattdessen wir nun ein NuGet Package angeboten. Sie auch hier: [https://docs.microsoft.com/de-de/sql/relational-databases/server-management-objects-smo/installing-smo](https://docs.microsoft.com/de-de/sql/relational-databases/server-management-objects-smo/installing-smo)

Wie ich das jetzt problemlos auf einem Windows Server 2012 R2 oder 2016 bekomme muss ich noch herausfinden.

Bis dahin bedanke ich mich für's Lesen.
