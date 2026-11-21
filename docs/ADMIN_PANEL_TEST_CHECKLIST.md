# Admin panel – grafikus teszt checklist & megjegyzések

Teszt környezet: lokális backend `http://127.0.0.1:8080`, admin token `local-dev-admin-token`, teszt DB `server/auth.db` (alice, bob, carol, dave). Tesztelés dátuma: 2026-11-21. Valós backend mód (`admin-real.js`), nem mock.

## 1. Előfeltételek / alapműködés
- [x] Admin token beállítása (Beállítás gomb) → "Valós backend aktív" jelzés megjelenik.
- [x] Fiókok száma: 4 regisztrált, 4 aktív, 0 online (szimulált), 0 letiltva a kezdeti állapotban.
- [x] Felhasználó táblázat: 4 sor (alice, bob, carol, dave), rang = Játékos, állapot = Online.
- [x] Kereső mező jelen van (nem tesztelt szűrés).
- [x] Audit napló kezdetben üres / csak új bejegyzések jelennek meg.
- [x] Rang mátrix táblázat megjelenik, szimulált rang = Tulajdonos oszlop kiemelve.
- [~] Érvénytelen token banner: nem tesztelve (token törlése → banner). Kódútvonal létezik (`admin-real.js`), de grafikusan nem ellenőriztük.

## 2. Ban / Unban (Játékos letiltása)
- [x] Ban modal megnyílik alice-nál.
- [x] Időtartam mező (perc, 0 = végleges) működik.
- [x] Indok mező (Minecraftban megjelenik) működik.
- [x] Hatókör (Csak Minecraft / Teljes platform / Mindkettő) választható.
- [x] Célpontok: Fiók / Minden HWID (30 nap) / IP-k (30 nap) checkboxok.
- [x] HWID előnézet: "1 eszköz" megjelenik.
- [x] IP előnézet: "127.0.0.1 — 4 fiók használta" megjelenik.
- [x] Ban megerősítés → alice állapota "Letiltva", gomb "Feloldás".
- [x] Unban (Feloldás) → alice visszaáll Aktív/Játékos állapotra.
- [x] Audit bejegyzés megjelenik: "admin · Játékos letiltása · alice · Teszt ban indok ...".

## 3. Kick (Kirúgás a szerverről)
- [x] Kick modal megnyílik alice-nál.
- [x] Időtartam / indok / hatókör / célpontok mezők megegyeznek a ban modallal.
- [x] HWID/IP előnézet megegyezik (1 eszköz, 127.0.0.1 — 4 fiók).
- [x] Megerősítés → státusz üzenet, modal bezárul.
- [~] Valós MC szerver kick nem tesztelhető lokálisan (nincs élő szerver); backend végpont hívása sikeres.

## 4. Mute (Némítás)
- [x] Mute modal megnyílik bob-nál.
- [x] Időtartam 10 perc beállítása működik.
- [x] Indok beírása működik.
- [x] Hatókör / célpontok választhatók.
- [x] Megerősítés → bob állapota "Online · Némítva", gomb "Némítás feloldása".
- [x] Audit bejegyzés: "admin · Némítás · bob · Teszt némítás 10 perc".

## 5. Unmute (Némítás feloldása)
- [x] "Némítás feloldása" gomb megjelenik bob-nál némítás után.
- [~] Klikk nem tesztelt (a flow logikailag azonos a ban feloldással); backend végpont létezik.

## 6. Warn (Figyelmeztetés)
- [x] Warn modal megnyílik dave-nál.
- [x] Időtartam / indok / hatókör / célpontok mezők működnek.
- [x] Megerősítés → státusz üzenet.
- [x] Audit bejegyzés: "admin · Figyelmeztetés · dave · Teszt figyelmeztetés".

## 7. Role váltás (Rang)
- [x] Rang legördülő (Játékos/Tulajdonos/Admin/Moderátor/Support) megjelenik.
- [x] dave rangjának módosítása (Moderátor kiválasztás) → állapot frissül (Admin lett a backend válasz alapján).
- [x] Staff számláló frissül (0 → 1 staff).
- [~] Audit bejegyzés role váltásról nem jelent meg a naplóban (lehet, hogy nem auditált művelet) — ellenőrizendő.

## 8. Jelszó visszaállítás (Password reset)
- [x] Jelszó visszaállítása modal megnyílik carol-nál.
- [x] Megerősítés → "Ideiglenes jelszó" dialog ideiglenes jelszóval (pl. `9LtSU@#P%FMQ`).
- [x] Utasítás megjelenik: belépés után kötelező új jelszót beállítani.
- [x] "Rendben" gomb bezárja a dialogot.
- [x] Státusz üzenet: "carol jelszava visszaállítva."
- [x] Audit bejegyzés: "admin · accounts.password-reset · carol · ideiglenes jelszó generálva".

## 9. Skin jóváhagyás
- [~] Nincs függőben lévő skin a teszt DB-ben → "Nincs függőben lévő skin" üres állapot megjelenik.
- [~] Elutasított skinek (blacklist) üres állapot megjelenik.
- [ ] Approve/Reject flow nem tesztelhető (nincs tesztadat). Backend endpoint létezik (`admin-real.js` + `admin_moderation.go`).

## 10. Egyéb
- [x] Világos/sötét mód kapcsoló jelen van.
- [x] Cookie/süti dialog megjelenik (zavaró overlay a teszt során).
- [ ] Fiók végleges törlése (delete) nem tesztelt (destruktív).
- [ ] Invalid token banner nem tesztelt.

---

## Megjegyzések – működés és fejlesztési javaslatok

### Működés (mit láttunk)
- A valós backend adapter (`admin-real.js`) stabilan működik: ban/mute/warn/kick/password-reset/role végpontok hívása sikeres, azonnali UI frissítés.
- Audit napló append-only: ban törlése után is megmarad a bejegyzés (immutábilis) — ez a legfontosabb követelmény teljesült.
- HWID/IP előnézet nagyon hasznos: 1 eszköz és 4 fiók ugyanazon IP (127.0.0.1) jól látható, segíti a döntést.
- Temp jelszó egyszeri megjelenítése biztonságos (nem tárolódik újra megjeleníthető formában).

### Használhatóság / UX javaslatok
1. **Cookie dialog zavaró**: a "Sütik és adatgyűjtés" modal minden oldalbetöltéskor/teszt során felugrik és elfedi a munkafelületet. Javaslat: teszt módban vagy admin oldalon ne jelenjen meg, vagy "csak szükségesek" automatikus elfogadása dev módban.
2. **Audit napló szűrés/keresés**: jelenleg csak utolsó bejegyzések listája. Javaslat: szűrés felhasználó/művelet szerint, export (CSV/JSON), lapozás.
3. **Role váltás auditálása**: a rangváltás nem jelent meg az audit naplóban. Javaslat: minden role change legyen auditált bejegyzés (ki, kinek, régi→új rang, indok).
4. **Kick modal "Időtartam" mező félrevezető**: kick esetén az időtartam nem releváns (kick = azonnali eltávolítás). Javaslat: kick modalból az időtartam mező elhagyása vagy "újracsatlakozás tiltása (perc)" átnevezése.
5. **Mute/Unmute gomb állapot**: "Némítás feloldása" csak némítás után jelenik meg — jó. De a némítás lejártát (10 perc) nem jelzi visszaszámláló; javaslat: hátralévő idő megjelenítése.
6. **Ban modal "végleges" (0 perc) vizuális jelzés**: javaslat piros figyelmeztetés, ha 0 = végleges ban.
7. **Skin előnézet (3D) nem tesztelhető** üres adatokkal; javaslat: demo skin feltöltés teszt módban.
8. **Nem tesztelt funkciók**: delete fiók, invalid token banner, skin approve/reject, unmute klikk, kereső szűrés. Ezeket külön tesztkörben kell lefuttatni.
9. **Role váltás eredmény**: Moderátor kiválasztásra az UI "Admin" lett — ellenőrizendő, hogy a backend mapping (moderator→admin?) szándékos-e vagy bug. Javaslat: konzisztens rangnevek frontend/backend között.

### Értékelés (1–5)
| Funkció | Működés | UX | Megjegyzés |
|---|---|---|---|
| Ban/Unban | 5 | 4 | Teljes flow, audit OK |
| Kick | 4 | 3 | Modal OK, élő szerver nélkül nem teljes |
| Mute | 5 | 4 | Időtartam + indok OK |
| Warn | 5 | 4 | OK |
| Password reset | 5 | 5 | Temp jelszó egyszeri, biztonságos |
| Role change | 3 | 3 | Eredmény mapping kérdéses |
| Skin approve | 1 | 2 | Nincs tesztadat |
| Audit log | 4 | 3 | Immutábilis, de szűrés hiányzik |
| Token kezelés | 4 | 4 | OK |

Összegzés: az admin panel valós backend módban működőképes, a fő moderációs flow-k (ban/mute/warn/kick/password reset) grafikusan ellenőrizve és működnek. Fő javítandó: cookie dialog zavarása, role-change audit és mapping, audit szűrés/export.