# BlockHost – Webová aplikace pro správu Minecraft serverů

**Autor:** Daniel Krejčí  
**Projekt:** Webová aplikace (BlockHost)  
**Typ dokumentu:** Rozšířená specifikace projektu a zadání  

---

## 1. Úvod a motivace

**BlockHost** je moderní webová aplikace určená pro centralizovanou správu, monitoring a automatizaci herních serverů Minecraft. 

Provozování herních serverů často vyžaduje manuální zásahy do konfiguračních souborů, ruční správu záloh a neefektivní alokaci systémových prostředků. Servery často běží nepřetržitě na maximální konfiguraci, i když je v noci nebo během dopoledne herní aktivita minimální. Cílem projektu BlockHost je nabídnout přehledné webové rozhraní, které nejen zjednodušuje každodenní správu serverů a hráčů, ale přináší i inteligentní dynamické řízení výkonu a nákladů.

---

## 2. Hlavní cíle projektu

- **Centralizace správy:** Poskytnout uživatelsky přívětivé rozhraní pro kontrolu stavu serverů bez nutnosti přímého přístupu do systémového terminálu.
- **Optimalizace nákladů a výkonu:** Umožnit definici různých výkonových profilů a jejich automatické přepínání v čase podle reálného vytížení.
- **Zvýšení spolehlivosti:** Nabídnout spolehlivý systém automatických i manuálních záloh a rychlé obnovy dat v případě potíží.
- **Flexibilita pro komunitní akce:** Podpora dočasných eventových profilů s navýšeným výkonem a automatickým návratem do standardního režimu po skončení akce.

---

## 3. Funkční specifikace

### 3.1 Správa instancí serverů
- **Základní ovládání:** Spuštění, zastavení, restartování a okamžité vypnutí (kill) serveru.
- **Interaktivní webová konzole:** Přímé zobrazení serverových logů v reálném čase a odesílání příkazů do serveru.
- **Monitoring stavu a systémových prostředků:**
  - Stav serveru (Online, Offline, Restarting, Error).
  - Využití procesoru (CPU), operační paměti (RAM) a diskového prostoru.
  - Aktuální počet připojených hráčů a hodnota TPS (Ticks Per Second).

### 3.2 Konfigurace a správa herního prostředí
- **Správa souboru `server.properties`:** Editace klíčových herních nastavení (herní mód, obtížnost, PvP, maximální počet hráčů, seed světa) přes formulářové prvky i přímý editor.
- **Správa verzí a jader:** Možnost volby serverového softwaru (Vanilla, Paper, Spigot, Purpur, Forge, Fabric).
- **Správa doplňků:** Přehled a správa nainstalovaných pluginů nebo modifikací.

### 3.3 Správa hráčů
- **Přehled aktivních hráčů:** Seznam online hráčů s jejich avatary/skiny, IP adresami a dobou připojení.
- **Administrativní zásahy:** Možnost přímo z rozhraní hráče vyhodit (kick), zabanovat (ban), omilostnit (unban) či udělit operátorská práva (OP).
- **Správa přístupových seznamů:** Správa Whitelistu a Banlistu.

### 3.4 Zálohování a obnova dat (Backup System)
- **Manuální zálohy:** Vytvoření okamžité zálohy světa a konfigurace na jedno kliknutí.
- **Plánované zálohy:** Automatické zálohování v pravidelných časových intervalech (např. denně ve 04:00).
- **Správa archivu:** Seznam vytvořených záloh s informací o datu, velikosti a stavu.
- **Obnova a export:** Obnova serveru z vybrané zálohy do původního stavu a možnost stažení záložního archivu (.zip / .tar.gz).

### 3.5 Výkonové profily a nákladový model
- **Definice profilů:** Nastavení hardwarových parametrů:
  - Alokovaná paměť RAM (např. 2 GB, 4 GB, 8 GB, 16 GB).
  - Počet přiřazených vCPU jader.
  - Priorita I/O operací a diskový prostor.
  - Cenová hladina za hodinu/měsíc provozu (fiktivní kredity nebo reálná měna).
- **Katalog profilů:** Např. *Eco* (nízká zátěž), *Standard* (běžný provoz), *Turbo* (vysoká zátěž / event).

### 3.6 Automatický plánovač (Scheduler)
- **Časové rozvrhy:** Definice pravidel pro přepínání profilů podle denní doby a dnů v týdnu.
  - *Příklad:* V noci a dopoledne (01:00–14:00) profil *Eco*, v odpoledních a večerních hodinách (14:00–01:00) profil *Standard*, o víkendech profil *Standard/Turbo*.
- **Plynulý přechod:** Upozornění hráčů ve hře před plánovaným restartem/změnou profilu a bezpečné uložení světa před aplikací změn.

### 3.7 Dočasné profily pro herní eventy
- **Jednorázové navýšení výkonu:** Možnost naplánovat dočasné přepnutí na vysoce výkonný profil na konkrétní datum a časové okno (např. sobota 18:00–22:00 pro komunitní turnaj).
- **Automatická expirace:** Po uplynutí stanoveného času se server automaticky vrátí k výchozímu výkonovému profilu nebo běžnému harmonogramu bez nutnosti ručního zásahu administrátora.

---

## 4. Uživatelské role a oprávnění

Aplikace bude podporovat systém uživatelských rolí (RBAC):
- **Administrátor (Owner):** Plný přístup ke konfiguraci systému, definici výkonových profilů a správě všech serverů.
- **Správce serveru (Manager):** Správa konkrétních serverů, konfigurace, plánování záloh a přepínání profilů.
- **Moderátor:** Přístup ke konzoli, správě hráčů (kick, ban, whitelist) a sledování stavu serveru bez možnosti měnit hardware a finanční nastavení.

---

## 5. Návrh architektury a technologií

### 5.1 Frontend
- **HTML5 & CSS3:** Moderní sémantická struktura a responzivní rozhraní přizpůsobené pro desktop i mobilní zařízení.
- **JavaScript / TypeScript:** Reaktivní klientská logika, dynamické aktualizace grafů a stavových indikátorů.
- **Stylování:** Čistý, přehledný dashboard s tmavým režimem (Dark Mode) inspirovaný moderními herními panely.

### 5.2 Backend & API
- **Webový server & REST API:** Zpracování klientských požadavků, autentizace uživatelů a business logika.
- **Komunikace s herním serverem:**
  - Využití protokolu **RCON** pro odesílání herních příkazů a správu hráčů.
  - Správa procesů na úrovni operačního systému nebo kontejnerizace (Docker API) pro spouštění/zastavování a dynamické limity paměti a CPU.

### 5.3 Databáze
- Relační databáze (např. SQLite / MySQL / PostgreSQL) pro ukládání:
  - Uživatelských účtů a oprávnění.
  - Nastavení serverových instancí.
  - Konfigurace výkonových profilů a naplánovaných úloh.
  - Historie záloh a systémových událostí (audit log).

---

## 6. Uživatelské rozhraní (UI koncept)

1. **Dashboard (Přehled):** Hlavní obrazovka se souhrnným stavem serveru, grafy využití CPU/RAM, rychlými tlačítky (Start/Stop/Restart) a seznamem online hráčů.
2. **Konzole:** Živý výpis logů s vyhledáváním a příkazovým řádkem.
3. **Profily & Plánovač:** Přehledný kalendář/časovač pro nastavení pravidel přepínání profilů a správa jednorázových eventů.
4. **Hráči:** Tabulka hráčů s možností okamžitých akcí (Whitelist, Ban, Kick, OP).
5. **Zálohy:** Seznam bodů obnovy s tlačítky pro vytvoření, stažení a obnovu.
6. **Nastavení:** Přehledné formuláře pro konfiguraci parametrů serveru.

---

## 7. Budoucí možnosti rozšíření

- **Discord integrace:** Odesílání notifikací na Discord server (upozornění na pád serveru, zahájení eventu, zálohování).
- **Platební brána / Kreditní systém:** Automatické dobíjení kreditů pro komunitní financování serveru.
- **Podpora pro více herních uzlů (Multi-node):** Možnost spravovat servery rozprostřené přes více fyzických nebo virtuálních strojů.
- **Automatická instalace módpacků:** Integrace s platformami jako Modrinth nebo CurseForge pro instalaci jedním kliknutím.
