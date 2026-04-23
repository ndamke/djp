# Plan Write Service

Kleiner Go-Schreibdienst fuer das lokale Bearbeiten von `content/plan-eintraege/*.md`.

## Ziel

- Drag-and-drop-Aenderungen aus dem Jahresplan entgegennehmen
- `startwoche`, `endwoche`, `bezug_typ` und `bezug_id` in Markdown-Dateien aktualisieren
- Hugo als Render-Schicht beibehalten

## MVP-Endpunkte

- `GET /api/health`
- `POST /api/plan-eintraege/:id/move`

## Start

Voraussetzung: `go` muss installiert sein. Auf macOS z. B. mit:

```bash
brew install go
```

Danach den Dienst starten:

```bash
cd tools/plan-write-service
go run ./cmd/server -data-root ../..
```

Standardadresse:

```text
127.0.0.1:8787
```

## Beispiel-Request

```bash
curl -X POST http://127.0.0.1:8787/api/plan-eintraege/pe-0001/move \
  -H 'Content-Type: application/json' \
  -d '{
    "startwoche": 6,
    "endwoche": 8,
    "bezug_typ": "fach",
    "bezug_id": "fach-informatik"
  }'
```

## Erster Testablauf

### 1. Beide lokalen Dienste starten

Terminal 1:

```bash
hugo server
```

Terminal 2:

```bash
cd tools/plan-write-service
go run ./cmd/server -data-root ../..
```

### 2. Health-Check pruefen

```bash
curl http://127.0.0.1:8787/api/health
```

Erwartet:

```json
{"ok":true,"service":"plan-write-service","version":"1.0.0"}
```

### 3. API-Smoke-Test mit Beispiel-Eintrag

Beispiel: `pe-0003` von Woche `8-10` auf `12-14` verschieben.

```bash
curl -X POST http://127.0.0.1:8787/api/plan-eintraege/pe-0003/move \
  -H 'Content-Type: application/json' \
  -d '{
    "startwoche": 12,
    "endwoche": 14,
    "bezug_typ": "fach",
    "bezug_id": "fach-mathematik"
  }'
```

Danach pruefen in:

```text
content/plan-eintraege/pe-0003.md
```

Erwartet:

- `startwoche: 12`
- `endwoche: 14`
- zusaetzlich neues Feld `revision: ...`

### 4. Ruecksetzen des Beispieldatensatzes

```bash
curl -X POST http://127.0.0.1:8787/api/plan-eintraege/pe-0003/move \
  -H 'Content-Type: application/json' \
  -d '{
    "startwoche": 8,
    "endwoche": 10,
    "bezug_typ": "fach",
    "bezug_id": "fach-mathematik"
  }'
```

### 5. Browser-Test fuer Drag-and-drop

Browser aufrufen:

```text
http://127.0.0.1:1313/jahresplaene/jp-bg-it-j1/
```

Testablauf:

- Block `LS1.2 Projektkosten kalkulieren` in der Zeile `Mathematik` greifen
- auf eine spaetere Woche in derselben Zeile ziehen, z. B. Start bei Woche `12`
- Statusmeldung beobachten
- nach erfolgreichem Speichern laedt die Seite neu

Danach pruefen:

- Block steht im Jahresplan an der neuen Position
- `content/plan-eintraege/pe-0003.md` hat die neuen Wochenwerte
- `revision` wurde aktualisiert

### 6. Typische Fehlerbilder

- `fetch failed` oder Statusfehler in der Seite:
  - Go-Service laeuft nicht oder auf falschem Port
- `405 METHOD_NOT_ALLOWED`:
  - Endpoint mit falscher HTTP-Methode aufgerufen
- `404 PLAN_ENTRY_NOT_FOUND`:
  - `pe-...` Datei existiert nicht
- `404 REFERENCE_NOT_FOUND`:
  - `bezug_id` zeigt auf kein vorhandenes Fach/Projekt
- `409 REVISION_MISMATCH`:
  - Planeintrag wurde parallel geaendert, Seite neu laden

## Hinweise

- Der Dienst ist fuer den lokalen Editor-Modus gedacht.
- Extern wird weiterhin nur der statische Hugo-Build aus `public/` veroeffentlicht.
- Der Dienst fuehrt keinen Hugo-Build aus; `hugo server` oder `hugo` muessen separat laufen.
