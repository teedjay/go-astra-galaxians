# ASTRA GALAXIANS

Et Galaxians-inspirert arkadespill laget med Go og Ebitengine. Styr romskipet, skyt aliens og overlev stadig vanskeligere bølger med retro pixelgrafikk, chiptune og store partikkeleksplosjoner.

## Skjermbilder

Introen med pixel-aliens og retrotypografi:

![Introskjermen i ASTRA GALAXIANS](docs/screenshots/intro.png)

En oppsatt kampscene med kraftigste våpennivå: to kontinuerlige lasere, bevegelige knuter og store partikkeleksplosjoner. Begge bildene er tatt med spillets egen renderer.

![Kampscene med Knot Beams, aliens og partikkeleksplosjon](docs/screenshots/action.png)

## Spillet

Fire typer aliens kommer i bølger på 40, fordelt på ulike formasjoner. De flyr inn langs myke baner, roterer etter bevegelsesretningen og bryter etter hvert ut i stupangrep. Du har tre liv. En beseiret alien gir 100 poeng, eller 250 under angrep. Beste poengsum beholdes til programmet avsluttes.

Våpenet oppgraderes fra **blaster** til **målsøkende raketter**, **raske laserpulser** og til slutt **kontinuerlige lasere** som følger ulike mål og danner bevegelige knuter. Hovedkanonen beholdes på alle nivåer.

| Pickup | Effekt |
| --- | --- |
| Grønn glødende kapsel | Oppgraderer ett nivå og setter våpentiden til 10 sekunder. Ved maks nivå endres ikke tiden. |
| Rød glødende kapsel | Nedgraderer ett nivå og setter våpentiden til 10 sekunder. |
| Gul pulserende stjerne | Legger til 5 sekunder uten å endre våpennivået. |

Nedtellingen vises øverst til høyre når du har et oppgradert våpen. Når tiden løper ut, går våpenet tilbake til blaster. Nedtellingen pauses gjennom hele oppholdet mellom bølgene, inkludert musikken og ventetiden etterpå, og fortsetter når neste bølge starter. På laveste nivå slippes bare oppgraderinger; på høyeste nivå slippes bare nedgraderinger og stjerner. Avfyrte prosjektiler fullfører normalt ved nedgradering, og kontinuerlige lasere trekkes tilbake i kanonene.

Introen har en SID-inspirert chiptune som spilles én gang. Mellom bølgene spilles første del av en kortere låt. Den fader ut før gjentakelsen, etter omtrent 5,7 sekunder, fulgt av to sekunders pause. Under aktive bølger høres bare lydeffekter. Når siste liv går tapt, eksploderer skipet i flammer, røyk og fallende pixelbiter før «game over» vises. En liten, ufarlig pixelkatt går av og til langs bunnen av skjermen.

## Attract mode

Når introlåten er ferdig, blir introen stående i ti sekunder før den fader ut. Spillet viser deretter en automatisk demonstrasjon som starter på en tilfeldig bølge. Skipet styrer og skyter selv, viser ulike våpen og kan ikke miste liv i demonstrasjonen. Beste poengsum påvirkes ikke.

Demonstrasjonen har ingen lydeffekter eller mellomnivå-jingler. I stedet spilles den originale C64-inspirerte låten **Orbit Runner** på omtrent 102 sekunder, med tung bass, portamento, pulsbølger og raske arpeggioer. Mot slutten fader lyd og bilde ut, før introen starter på nytt med innfading og intromusikk. Et nytt tastetrykk eller museklikk under demonstrasjonen eller overgangen til den starter samme retur til introen. Dette gjelder også Esc og P. Enter eller mellomrom fra den vanlige introen starter fortsatt et vanlig spill.

## Kontroller

| Tast | Handling |
| --- | --- |
| Enter eller mellomrom | Start fra introen etter innfadingen. |
| Venstre/høyre pil eller A/D | Beveg skipet. |
| Hold mellomrom | Skyt med hovedkanon og tilgjengelige sidevåpen. |
| P | Pause eller fortsett. Nedtellingen pauses også. |
| Esc | Avslutt fra introen eller under spilling. |
| En tast eller museklikk på «game over» | Start utfading tilbake til introen. |

«Game over» går også automatisk tilbake til introen etter ti sekunder. Skipseksplosjonen fullføres før denne skjermen vises; tastetrykk hopper ikke over eksplosjonen.

## Bygge og kjøre

Du trenger **Go 1.27.1 eller nyere**, et grafisk skrivebord og nettverkstilgang for første nedlasting av avhengigheter. Kjør kommandoene fra prosjektmappen.

Start direkte:

```sh
go run .
```

Bygg og start på macOS eller Linux:

```sh
go build -o bin/astra-galaxians .
./bin/astra-galaxians
```

Bygg og start på Windows med PowerShell:

```powershell
go build -o bin/astra-galaxians.exe .
.\bin\astra-galaxians.exe
```

Go laster automatisk ned avhengighetene. På Linux trenger Ebitengine også systemets utviklingsbiblioteker for grafikk og lyd, blant annet X11, OpenGL og ALSA. Prosjektet er bygget og testet på macOS; Windows og Linux er ikke verifisert her.

## Teknisk informasjon

Spillet bruker **Ebitengine 2.10.0**, en logisk oppløsning på **640 × 800** og et vindu som kan endre størrelse. Alienbevegelser og laserbaner bruker kubiske Bézier-kurver. Pixelgrafikken genereres i Go, mens musikk og lydeffekter syntetiseres som stereo PCM ved 44,1 kHz. Partiklene har begrenset levetid, og skipets vrakbiter får tyngdekraft. Ingen eksterne bilde- eller lydfiler er nødvendige, og poengsum lagres ikke mellom oppstarter.

Prosjektet er organisert slik:

```text
main.go          Startpunkt for go run .
internal/game/   Spillogikk, grafikk, lyd og tester
bin/             Bygde programmer (ignoreres av Git)
docs/screenshots/ Skjermbilder brukt i README
```

I `internal/game/` ligger hovedløkken i `game.go`, oppstarten i `run.go`, våpnene i `weapons.go` og lyden i `sound.go`, `sid_music.go` og `attract_music.go`. Partikler, skipseksplosjon, skjermoverganger, attract mode og katten har egne filer. Testene ligger ved siden av koden de tester.

Kjør tester og statisk kontroll:

```sh
go test ./...
go vet ./...
```

Testene dekker blant annet formasjoner, bevegelsesbaner, våpen, pickups, tidsstyring, eksplosjoner, skjermoverganger og genererte lyddata.
