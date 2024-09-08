# 3. Domača naloga: Razširjanje sporočil

**Rok za oddajo: 17. 12. 2023**

Na podlagi [razlage](../07-razsirjanje-sporocil/Razsirjanje-sporocil.md), želimo poustvariti protokol za razširjanje sporočil med procesi.

Vaša naloga je napisati program v **go** za razširjanje sporočil med procesi in ga preizkusiti na **gruči Arnes**. Program naj preko ukazne vrstice prejme naslednje argumente (lahko tudi dodatne, če je potrebno):
- identifikator procesa `id`: celo število, ki identificira posamezen proces znotraj skupine,
- število vseh procesov v skupini `N`,
- število sporočil `M`, ki jih bo razširil glavni proces,
- število sporočil `K`, ki jih bo vsak proces ob prvem prejemu posredoval ostalim. S tem parametrom nastavljamo tudi metodo razširjanja sporočil. Če nastavimo `K==N-1`, naj se uporabi **nestrpno razširjanje**. Za `K<N-1` pa naj se uporabi **razširjanje z govoricami**.

Procesi naj za komunikacijo uporabljajo protokol **UDP**. Vsak proces naj ob prvem prejemu novega sporočila to izpiše na zaslon. Nadaljnje prejeme istega sporočila, naj proces ignorira. Razširjanje sporočila naj vedno začne proces z `id==0`.

Pri **nestrpnem razširjanju**, naj vsak proces posreduje sporočilo vsem ostalim v skupini. To naj stori samo ob **prvem** prejemu novega sporočila. Vsebina sporočil, ki se razširjajo, je lahko poljubna, vendar mora omogočati, da procesi med sabo ločijo različna sporočila. Komunikacijo vedno začne glavni proces (`id==0`), ki zaporedoma pošlje `M` sporočil. Med posamezna pošiljanja dodajte kratko pavzo (reda 100 ms). Pri razširjanju ni potrebno skrbeti za pravilen vrstni red prejema. 

![Potek komunikacije](./komunikacija.png)

Pri **razširjanju z govoricami** naj vsak proces ob prvem prejemu novega sporočila, le-tega posreduje naprej `K` naključno izbranim procesom. Pri naključnem izbiranju procesov poskrbite, da posamezen proces ne izbere istega procesa večkrat.

Pri pisanju programa se lahko zgledujete po kodi iz [prejšnjih vaj](../06-posredovanje-sporocil/Posredovanje-sporocil.md). Pri poslušanju za sporočila je priporočeno, da nastavite rok trajanja povezave s pomočjo funkcije [SetDeadline](https://pkg.go.dev/net#IPConn.SetDeadline) ali pa kako drugače poskrbite, da se proces zaključi in sprosti vrata, če po nekem času ne dobi sporočila. S tem se boste izognili težavam z zasedenostjo vrat v primeru, da pride do smrtnega objema, ko nek proces čaka na sporočilo, ki nikoli ne pride.

V procesih ni potrebno uporabiti principa preverjanja utripa za ugototavljanje, če so procesi prejemniki pripravljeni oziroma živi. Glavni proces naj kar takoj začne pošiljati sporočila. 

Za zaganjanje poljubnega števila procesov na gruči Arnes se lahko zgledujete po [skripti](../06-posredovanje-sporocil/koda/run_telefon.sh) iz prejšnjih vaj.

## Navodila za strukturo funkcij in zahteve za pravilno poganjanje testov

Da se bodo vaši programi pravilno izvajali in uspešno prestali avtomatsko testiranje, prosimo, da upoštevate naslednje zahteve pri implementaciji funkcij:

1. **Funkcija `receiveMessages`:**
   - Funkcija naj sprejema naslednje argumente:
     - `conn net.PacketConn`: povezava, ki jo proces uporablja za prejemanje UDP sporočil.
     - `id int`: identifikator procesa.
     - `N int`: število vseh procesov v skupini.
     - `K int`: število procesov, katerim se posreduje sporočilo (razširjanje).
     - `basePort int`: osnovni port, ki ga procesi uporabljajo za komunikacijo.
     - `receivedMessages map[string]bool`: mapa, ki sledi že prejetim sporočilom.
   - Implementacija mora ob prvem prejemu novega sporočila to izpisati na zaslon in nato posredovati naprej glede na metodo razširjanja (nestrpno razširjanje ali razširjanje z govoricami).

2. **Funkcija `sendInitialMessages`:**
   - Funkcija naj sprejema naslednje argumente:
     - `N int`: število vseh procesov v skupini.
     - `M int`: število sporočil, ki jih razširi glavni proces.
     - `K int`: število procesov, katerim se posreduje sporočilo (razširjanje).
     - `id int`: identifikator procesa.
     - `basePort int`: osnovni port, ki ga procesi uporabljajo za komunikacijo.
     - `receivedMessages map[string]bool`: mapa, ki sledi že prejetim sporočilom.
   - Implementacija mora poskrbeti za to, da proces z `id == 0` pošlje `M` sporočil ostalim procesom, s pavzo med pošiljanji.

3. **Funkcija `forwardMessage`:**
   - Funkcija naj sprejema naslednje argumente:
     - `N int`: število vseh procesov v skupini.
     - `K int`: število procesov, katerim se posreduje sporočilo (razširjanje).
     - `id int`: identifikator procesa.
     - `basePort int`: osnovni port, ki ga procesi uporabljajo za komunikacijo.
     - `message string`: vsebina sporočila.
   - Funkcija naj poskrbi za posredovanje sporočila vsem procesom (v primeru nestrpnega razširjanja) ali naključnim `K` procesom (v primeru razširjanja z govoricami).

4. **Funkcija `forwardToRandomProcesses`:**
   - Funkcija naj sprejema naslednje argumente:
     - `N int`: število vseh procesov v skupini.
     - `K int`: število procesov, katerim se posreduje sporočilo (razširjanje).
     - `id int`: identifikator procesa.
     - `basePort int`: osnovni port, ki ga procesi uporabljajo za komunikacijo.
     - `message string`: vsebina sporočila.
   - Funkcija mora izbrati `K` naključnih procesov (razen sebe) in jim posredovati sporočilo.

5. **Funkcija `sendMessage`:**
   - Funkcija naj sprejema naslednje argumente:
     - `address string`: naslov, na katerega pošljemo sporočilo (v obliki "localhost:port").
     - `message string`: vsebina sporočila.
   - Funkcija mora poslati sporočilo določenemu procesu prek UDP.

### Zahteve za izvajanje testov

Da bodo vaši programi uspešno prestali avtomatsko testiranje, poskrbite, da:

- Se vsako novo prejeto sporočilo izpiše na zaslon v obliki `fmt.Printf("Process %d received message: %s\n", id, message)`.
- Funkcije so poimenovane natančno tako, kot je navedeno zgoraj.
- Program deluje v skladu z navodili, tako da sledi pravilom nestrpnega razširjanja in razširjanja z govoricami.
- Uporabite mapo `receivedMessages` za sledenje že prejetim sporočilom in preprečevanje večkratnega pošiljanja istega sporočila.

Upoštevanje teh smernic bo zagotovilo, da bo vaša rešitev ustrezno delovala v vseh predvidenih scenarijih in prestala avtomatsko preverjanje.