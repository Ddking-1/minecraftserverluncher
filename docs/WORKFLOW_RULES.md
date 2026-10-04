\# AI EGYÜTTMŰKÖDÉSI ÉS MUNKAFOLYAMAT SZABÁLYZAT (WORKFLOW\_RULES)



\## Szerepkörök

1\. \*\*Tervező (Architect - DeepSeek V3.2 / Qwen3 32B):\*\* Projektmenedzser, kód-felülvizsgáló (Reviewer), integrátor. Ő adja ki a feladatokat, és ellenőrzi a minőséget. Közvetlen kódolást nem végezhet!

2\. \*\*Programozó (Worker - DeepSeek-Coder-V2-Lite-Instruct):\*\* Kódoló, tesztelő. Ő hajtja végre a feladatokat a parancssoron (CLI) keresztül.



\## Alkalmazandó Skillek és Kérdezési Protokoll (Anti-Találgatás)

\- \*\*skills/ask-first/SKILL.md:\*\* Tilos a találgatás! Mielőtt bármilyen kódolási feladatot kiadna, a Tervező ellenőrzi, hogy a cél 100%-ig egyértelmű-e. Ha hiányzik egy technikai részlet, TILOS elkezdeni a kódolást — a Tervező köteles megállni és kérdezni.

\- Ha a Tervező kérdez, az `EMBERI CSELEKVÉS SZÜKSÉGES` mező értékét kötelező `Igen`-re állítani.



\## A Fejlesztési Ciklus (Minden feladatnál KÖTELEZŐ)

1\. \*\*Feladat Kijelölése \& Ellenőrzés:\*\* A Tervező értékeli a feladatot (ha nem tiszta, kérdez az `ask-first` szerint).

2\. \*\*Git Branch (Worker):\*\* A Tervező utasítja a Workert egy új branch létrehozására.

3\. \*\*Kódolás (Worker):\*\* A kódolás a `skills/ponytail/` és `skills/wednesday-skills-pack/` elvek alapján történik.

4\. \*\*Tesztelés (Worker):\*\* A Worker lefuttatja a teszteket/buildet, és csak a nyers kimenetet adja vissza.

5\. \*\*Review és Merge (Architect):\*\* 

&#x20;  - Siker esetén a Tervező mergel.

&#x20;  - Hiba esetén új promptot ad (max 8 iteráció).

6\. \*\*8-Iteration Retry Limit (Megállás):\*\* Ha a Worker 8 egymást követő próbálkozás után sem tudja megoldani a hibát, a Tervezőnek KÖTELEZŐ MEGÁLLNIA, és az Ember beavatkozását kérnie.



\## Kommunikációs Formátum (Kötelező Zárás)

A Tervező minden válaszát az alábbi blokkal kell zárnia:

\- \*\*JELENLEGI ÁLLAPOT:\*\* \[Sikeres/Tesztelve/Hibás/Kérdés van]

\- \*\*KÖVETKEZŐ LÉPÉS:\*\* \[Mi a következő teendő]

\- \*\*EMBERI CSELEKVÉS SZÜKSÉGES:\*\* \[Igen / Nem]

