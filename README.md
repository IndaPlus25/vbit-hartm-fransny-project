# GoQuant Backtesting Engine

## Syfte
Detta projekt är en snabb, händelsestyrd (event-driven) backtesting-motor byggd från grunden för att utvärdera kvantitativa handelsstrategier. Syftet med plattformen är att kunna processa stora mängder historisk tick/bar-data för att hitta matematiska "edges" i marknaden (alfa). Motorn är designad för att vara strikt kronologisk, vilket helt eliminerar risker för s.k. *look-ahead bias*, och simulerar en verklig handelsmiljö så nära som möjligt.

## Varför Go?
Motorn är skriven i **Go (Golang)** av tre huvudsakliga anledningar:
1. **Prestanda:** Go är kompilerat och otroligt snabbt, vilket är ett absolut krav när man itererar över miljontals historiska prispunkter.
2. **Concurrency (Samtidighet):** Finansiell backtesting är "skamlöst parallelliserbart" (embarrassingly parallel). Genom Go:s *goroutines* kan vi exekvera flera tunga strategier på flera CPU-kärnor samtidigt över samma i-minnet-data, vilket kapar exekveringstiden drastiskt.
3. **Typsäkerhet:** Ett starkt typat system (via våra egna `structs`) förhindrar tysta fel som annars är vanliga i t.ex. Python när man hanterar komplexa finansiella datastrukturer.

## Varför Python?
1. Även om Go är överlägset när det gäller rå beräkningskraft och parallell exekvering, saknar språket det mogna ekosystem för data science som krävs för effektiv visualisering. Därför har vi valt en polyglot-arkitektur där Go agerar backend-motor och Python agerar presentationslager.
Valet av Python för resultathanteringen motiveras av följande:
2. Rätt verktyg för rätt jobb (Separation of Concerns): Genom att låta Go sköta det tunga lyftet (iterera över miljontals datapunkt och utvärdera logik) och låta Python sköta analysen, maximerar vi effektiviteten. Vi tvingar inte in Go i en roll det inte är byggt för (grafik), och vi tvingar inte Python att göra tunga loop-beräkningar.
3. Oslagbart ekosystem (Pandas & Matplotlib): Pythons bibliotek för datamanipulation (pandas) och grafritning (matplotlib / plotly) är branschstandard. Att aggregera trades, beräkna kumulativ vinst (Cumulative PnL) och rita upp en professionell graf tar bokstavligen 5 rader kod i Python. Att försöka bygga motsvarande grafer direkt i Go hade inneburit onödigt hög komplexitet.
4. Rapid Prototyping & EDA: När resultatet (vår 9-kolumners CSV-fil) väl är genererat, tillåter Python snabb Exploratory Data Analysis (EDA). Om vi vill lägga till en ny graf som visar "Vinst per veckodag" eller räkna ut Sharpe-kvoten, kan vi göra det omedelbart i ett Python-skript eller en Jupyter Notebook utan att behöva kompilera om själva handelsmotorn.
5. Kvantitativ branschstandard: Inom den finansiella sektorn och bland kvantitativa analytiker är Python det obestridda standardspråket för analys. Att bygga en pipeline som levererar färdigstädad data rakt in i Python speglar hur professionella trading-firmor bygger sina system (C++/Go för exekvering, Python för research).

## Arkitektur & Struktur
Projektet är uppdelat i tydliga domäner (Separation of Concerns):

* `data/` (Storage): Ansvarar för I/O. Läser in tunga Parquet-filer, omvandlar dem till Go-slices och **sorterar** dem kronologiskt för att garantera tidsintegritet innan simuleringen startar.
* `engine/`: Hjärtat i systemet. Innehåller en **Tillståndsmaskin (State Machine)** som itererar över data. Den håller koll på öppna positioner i minnet och loggar en komplett trade först när både *entry* och *exit* är slutförda. Stöder både långa (köp) och korta (blankning) positioner.
* `strategy/`: Modulärt gränssnitt (`interface`). Nya strategier kan pluggas in utan att motorn behöver skrivas om. Strategin får endast tillgång till historisk data fram till den aktuella sekunden.
* `python/`: Eftersom Go är bäst på beräkningar men Python är bäst på visualisering, exporteras resultatet (9 datakolumner) till en CSV-fil som sedan ritas upp via ett externt Python-skript (Pandas/Matplotlib).
* `types/`: Paketet types (ofta kallat domänlagret) är hjärtat i systemets dataflöde. Anledningen till att alla structs är isolerade i en egen mapp är för att undvika cykliska beroenden (circular dependencies) i Go. Genom att definiera typerna här, kan både data, engine och strategy importera dem fritt utan att trassla in sig i varandra.  

## Strategier & Teori
I den nuvarande versionen testas följande strategier:
* **SMA Cross:** En klassisk trendföljande algoritm. Går in i position när ett snabbare glidande medelvärde korsar ett långsammare.
* **FVG (Fair Value Gap):** Identifierar obalanser i orderboken (ineffektiv prissättning) baserat på tre-ljus-formationer, och handlar på teorin om att marknaden tenderar att "fylla" dessa gap.
* **Liquidity Sweep (MTF):** Bygger på marknadsstruktur. Strategin letar efter tillfällen där priset tillfälligt bryter tidigare toppar/bottnar för att "jaga stop-loss-ordrar" (samla likviditet) innan det vänder tillbaka i motsatt riktning.

## Förbättringspotential (Future Work)
Systemet är i dagsläget en "Trade-level backtester", vilket mäter den rena matematiska träffsäkerheten för 1 enhet per affär. För framtida versioner planeras följande:
1. **Portfolio-Level Engine:** Introducera ett dynamiskt startkapital och en riskmotor. Detta möjliggör *Position Sizing* (t.ex. att riskera exakt 1% av portföljen per trade baserat på avståndet till stop-loss).
2. **Avancerade Metriker:** Beräkna och exportera nyckeltal som Sharpe Ratio, Max Drawdown och Win/Loss-ratio direkt i Go-motorn.
3. **Live Trading Integration:** Koppla motorn till ett mäklar-API (t.ex. Interactive Brokers eller Binance) för att omvandla de asynkrona handelssignalerna till skarpa marknadsordrar.