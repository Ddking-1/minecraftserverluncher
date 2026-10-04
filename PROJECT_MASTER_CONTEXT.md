\# 1. Projekt Áttekintés \& Technológiai Stack







\*\*Projekt Célja:\*\*



Egy egyedi, nagy teljesítményű és biztonságos hibrid Minecraft indítórendszer (Launcher) és a hozzá tartozó hitelesítési Backend fejlesztése. A rendszer fő célja a csalók és bannolt játékosok távoltartása (hardveres azonosítás és IP alapján), valamint egy biztonságos, munkamenet-alapú (session token) játékindítás biztosítása. A projekt tartalmaz továbbá egy Pterodactyl automatizációs réteget is, amely kizárólag adminisztratív és AI-tesztelési célokat szolgál.







\*\*Architektúra \& Technológiai Stack:\*\*



\*   \*\*Kliens (Launcher):\*\* C++ (C++17/20 szabvány), Qt 6 keretrendszer, CMake építőrendszer. (Felelős a UI-ért, HWID generálásért, Java detektálásért és a Minecraft háttérfolyamatként való indításáért).



\*   \*\*Backend (API \& Auth):\*\* Go (Golang). (Felelős a REST API végpontokért, bcrypt jelszókezelésért, Session tokenek kiadásáért, valamint a HWID/IP szintű ban ellenőrzésért az adatbázisból).



\*   \*\*Automatizáció \& Adminisztráció:\*\* Python 3 (`ptero\_control.py`). (A Pterodactyl API-n keresztül kezeli a tesztszervereket admin/AI célokra).



\*   \*\*Fejlesztői Környezet:\*\* Windows OS, OpenCode IDE (lokális AI ügynök autonóm terminál/fájlkezelési jogokkal), lokális LLM (LM Studio, 25k+ context, `q8\_0` KV cache kvantálás).







\---







\# 2. Működési Szabályok, Preferenciák \& Elvárások (STRICT RULES)







\*\*1. Kódolási és Architektúrális Szabályok:\*\*



\*   \*\*Egyszeri fájlolvasás:\*\* Egy adott prompt futása alatt minden fájlt szigorúan csak EGYSZER szabad beolvasni a token-keret kímélése érdekében. Tilos a fájlok közötti oda-vissza ugrálás.



\*   \*\*Autonóm Hibajavítás:\*\* Ha a CMake konfiguráció, a C++ fordítás (build) vagy a Go teszt során szintaktikai/fordítási hiba lép fel, az AI-nak kifejezett engedélye és kötelessége azt önállóan debuggolni és javítani, anélkül, hogy a felhasználótól kérdezne.



\*   \*\*Elszeparált Felelősségi Körök:\*\* A Pterodactyl vezérlés (szerver indítás/leállítás) SZIGORÚAN csak adminisztrátoroknak és AI tesztelésre van fenntartva. A normál C++ játékkliens nem indíthatja a szervert, csak a játékmotort.







\*\*2. Munkafolyamat (Workflow):\*\*



\*   \*\*Git-alapú CI/CD szimuláció:\*\* Minden új funkciót új ágon kezdünk (`git checkout -b feat/nev`). A fejlesztés és autonóm tesztelés után a folyamat automatikusan commitol, pushol a feature ágra, majd visszavált `master`-re, beolvasztja (merge) a változásokat, és feltolja a főágat a távoli (origin) tárolóba.



\*   \*\*Állapotjelentések:\*\* Az AI-nak minden egyes logikai lépésnél ki kell írnia: `\[ÁLLAPOT] - {Mit csinálok éppen}`.







\*\*3. Felügyelet Nélküli Futás (Off-grid mód) - KRITIKUS:\*\*



A hosszabb feladatokat a felhasználó felügyelet nélkül hagyja. Emiatt két szigorú szabály él:



\*   \*\*Előzetes Engedély-Trigger:\*\* A prompt legelején le kell futtatni egy `cmd /c "cd .. \&\& cd minecraftserverluncher"` (vagy hasonló mappa-kilépő) parancsot, hogy az OpenCode biztonsági ablaka azonnal felugorjon, és a felhasználó még a gép elhagyása előtt megadhassa a "Mindig engedje meg" jogosultságot.



\*   \*\*Automatikus Leállítás:\*\* A sikeres git merge és push után, a feladat legutolsó lépéseként az AI kiadja a `cmd /c "shutdown /s /f /t 180"` parancsot a PC 3 perces, kényszerített kikapcsolásához. Ezt TILOS futtatni, ha a kód/git push nem volt 100%-osan sikeres.







\---







\# 3. Eddigi Előrehaladás \& Meghozott Döntések







\*\*Kész modulok (1. és 2. Fázis lezárva):\*\*



\*   \*\*Go Backend Auth \& Ban System:\*\* A `/api/server/start` (csak admin) és a login végpontok élnek. A `database.go` és a `handlers.go` tartalmazza a `IsHWIDBanned` logikát. Bannolt HWID esetén a szerver 403 Forbidden státuszt ad.



\*   \*\*Pterodactyl Vezérlés:\*\* A `ptero\_control.py` script elkészült az admin/AI környezetkezeléshez.



\*   \*\*C++ Launcher Hitelesítés:\*\* A `BackendAuthClient` C++ osztály sikeresen csatolja a gép HWID-jét a login JSON payloadhoz, és fogadja a tokeneket (melyeket a `TokenStorage` / `SessionManager` kezel).



\*   \*\*C++ Qt UI \& Visszajelzések:\*\* A Login ablak UI össze van kötve a backenddel. Helytelen jelszó vagy 403-as HWID ban esetén vizuális hibaüzenet (pl. Qt `QMessageBox`) jelenik meg.



\*   \*\*Java Kereső Motor:\*\* A `JavaDetector` C++ osztály implementálva lett (ellenőrzi a JAVA\_HOME-ot és a Program Files mappákat JRE/JDK 17-21 verziókért).







\*\*Kulcsfontosságú Tervezési Döntések:\*\*



\*   \*\*Pterodactyl leválasztása a játékosokról:\*\* Döntés született, hogy a játékkliens nem kommunikál a Pterodactyl-lal. Ez egy kizárólag webalapú/admin funkció marad.



\*   \*\*VRAM Kímélés:\*\* Az AI modell KV cache-e `q8\_0` kvantáláson fut a nagy (25k) kontextusablak miatt, így az AI memóriakezelése optimalizált.







\---







\# 4. Jelenlegi Állapot \& Közvetlen Következő Lépés







\*\*Jelenlegi Állapot:\*\*



A projekt a \*\*3. Fázis (Minecraft Játékindító Motor)\*\* küszöbén áll (vagy épp annak legelején). A hitelesítés és a környezet (Java) ellenőrzése működik.







\*\*Közvetlen Következő Lépés:\*\*



A C++ kliensben meg kell írni a tényleges játékindító logikát (`MinecraftLauncher` osztály).



1\.  Fel kell használni a `JavaDetector` által visszaadott java.exe elérési utat.



2\.  Ki kell olvasni az érvényes Session Tokent és a felhasználó UUID/Név adatait.



3\.  Össze kell állítani a Minecraft indításához szükséges parancssori argumentumokat (pl. `--username`, `--uuid`, `--accessToken`, `--gameDir`).



4\.  A Qt `QProcess` osztályának segítségével el kell indítani a játékot önálló háttérfolyamatként, és a C++ UI "Play" gombjának feliratát/állapotát ennek megfelelően frissíteni (pl. "Játék fut...").







\---







\# 5. Master System Instruction (Új chathoz)







\*\*Kérlek, az új session indításakor ezt tekintsd az alapvető viselkedési mintádnak:\*\*







"Te egy Senior Full-Stack (C++ Qt \& Go) Szoftverarchitekt és Fejlesztő vagy. A feladatod a fentebb leírt egyedi Minecraft Launcher projekt folytatása és befejezése az OpenCode környezetben.



Kötelességed szigorúan betartani a Context Handover Documentben rögzített technológiai határokat (különösen a fájl-hozzáférési limiteket és az autonóm hibajavítást). Mielőtt bármilyen kódot írnál, röviden igazold vissza ezen dokumentum megértését, és várj a felhasználó első specifikus feladat-promptjára. 



Minden válaszodban és terminál-műveletedben priorizáld az önállóságot, a precíz C++17/Qt6 és Go idiómákat, valamint a felhasználó által kért 'Off-grid' automatizációs (Git merge + PC leállítás) lépéseket."





