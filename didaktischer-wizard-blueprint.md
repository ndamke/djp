# Didaktischer Wizard Blueprint

## Betriebsmodus (kurz)

- **Lokal = Bearbeiten:** Jahresplan wird mit Hugo + Go-Schreibdienst aktiv gepflegt (inkl. Drag-and-drop).
- **Extern = Anzeigen:** Auf dem Webserver liegt nur der statische Hugo-Build (`public/`) als Viewer.
- **Veroeffentlichung = Build + Deploy:** Lokale Aenderungen werden erst nach `hugo`-Build und Upload von `public/` extern sichtbar.
- **Modus-Schaltung:** Editor-Modus ist standardmaessig aktiv bei `hugo server` (`hugo.IsServer`), bei statischem Build automatisch aus.
- **Optionaler Override:** `params.editor_mode = true|false` kann den Modus explizit erzwingen.

## 1) Ordnerstruktur anlegen

- `content/bildungsgaenge/`
- `content/faecher/`
- `content/projekte/`
- `content/lernsituationen/`
- `content/jahresplaene/`
- `content/plan-eintraege/`
- `layouts/jahresplaene/`
- `layouts/lernsituationen/`
- `layouts/partials/`
- `static/anlagen/`

## 2) Zentrale Konventionen festlegen

- IDs sind stabil, klein, kebab-case (`ls-001`, `bg-it`, `proj-webshop`).
- Dateiname = `id` (z. B. `ls-001.md`).
- Nur Markdown-Links, keine Wikilinks (`[Text](/lernsituationen/ls-001/)`).
- `plan-eintraege` sind Datensaetze, nicht Seiten, und werden nicht oeffentlich gerendert.
- Ein Jahresplan hat immer 40 Wochen.

## 3) Minimal-Dateien (Content)

- `bildungsgaenge/*.md`: `id`, `type`, `title`, `dauer_jahre`, `draft`
- `faecher/*.md`: `id`, `type`, `title`, `sortierung`, `draft`
- `projekte/*.md`: `id`, `type`, `title`, `nummer`, `kurztext`, `sortierung`, `draft`
- `lernsituationen/*.md`: `id`, `type`, `title`, `nummer`, `kurztext`, `bildungsgang_id`, `ausbildungsjahr`, `fach_id`, `draft`
- `jahresplaene/*.md`: `id`, `type`, `title`, `bildungsgang_id`, `ausbildungsjahr`, `wochen: 40`, `draft`
- `plan-eintraege/*.md`: `id`, `type`, `jahresplan_id`, `lernsituation_id`, `bezug_typ`, `bezug_id`, `startwoche`, `endwoche`, `anzeige_nummer`, `anzeige_kurztext`, `draft`

## 4) Minimal-Dateien (Layouts)

- `layouts/jahresplaene/single.html`
  - laedt aktuellen Plan, Plan-Eintraege, Faecher und Projekte
  - ruft das Matrix-Partial auf
- `layouts/partials/jahresplan-matrix.html`
  - rendert Tabelle mit Kopfzeile `W1` bis `W40`
  - rendert Zeilen fuer Faecher und Projekte
- `layouts/partials/jahresplan-cell.html`
  - filtert passende Eintraege fuer Zeile und Woche
  - zeigt Link mit `anzeige_nummer` und `anzeige_kurztext`
- `layouts/lernsituationen/single.html`
  - rendert Detailansicht inklusive Anlagen
  - enthaelt optionalen Ruecklink zum Jahresplan

## 5) URL-Schema (empfohlen)

- Lernsituation: `/lernsituationen/<id>/`
- Jahresplan intern: `/jahresplaene/<id>/`
- Optional sprechend: `/jahresplan/<bildungsgang-id>/<ausbildungsjahr>/`
- Anlagen: `/anlagen/<lernsituation-id>/<datei>`

## 6) Renderfluss (fachlich)

1. Eine Lernsituation wird in `content/lernsituationen/*.md` angelegt.
2. Die Zuordnung zum Jahresplan erfolgt ueber `content/plan-eintraege/*.md`.
3. Die Jahresplan-Seite sammelt alle Eintraege mit passender `jahresplan_id`.
4. Die Matrix rendert fuer jede Zeile (Fach/Projekt) und Woche (1-40) den passenden Eintrag.
5. Klick auf eine belegte Zelle fuehrt zur Detailseite der verknuepften Lernsituation.

## 7) Validierung vor Veroeffentlichung

- Alle Referenzen (`*_id`) zeigen auf existierende Inhalte.
- `1 <= startwoche <= endwoche <= 40`.
- `bezug_typ` ist `fach` oder `projekt`, und `bezug_id` passt dazu.
- Pro Kombination aus `bildungsgang_id` und `ausbildungsjahr` existiert genau ein Jahresplan.
- Produktive Inhalte haben `draft: false`.

## 8) Naechste sinnvolle Schritte

1. Beispielinhalte fuer einen Bildungsgang, zwei Faecher, ein Projekt und zwei Lernsituationen anlegen.
2. `jahresplaene/single.html` und die beiden Partials als lauffaehige Hugo-Templates erstellen.
3. Die Jahresplan-Matrix zuerst ohne Styling, danach mit horizontalem Scroll und klaren Zelllinks umsetzen.

## Blockbibliothek unter dem Raster (Neues Planungs-Feature)

### Zielbild

- Der Zeilenkopf eines Fachs ist klickbar und aktiviert die Zielzeile.
- Unter dem Raster erscheint eine Auswahlliste mit potenziellen Lernsituationen.
- Lernsituationen werden einzeln per Drag-and-drop aus der Liste in freie Wochen der aktivierten Zeile gelegt.
- Bereits eingeplante Lernsituationen verschwinden aus der Liste.

### Fachliche Regeln (MVP)

- Potenzielle Lernsituationen werden nach `bildungsgang_id` und `ausbildungsjahr` des aktuellen Jahresplans gefiltert.
- Potenzielle Lernsituationen werden zusaetzlich ueber `fach_id` auf die aktivierte Fachzeile gefiltert.
- Potenzielle Lernsituationen zeigen eine gespeicherte Standarddauer in Wochen.
- Neu eingeplante Bloecke werden direkt mit dieser Standarddauer gesetzt.
- Ueberlappungen in derselben Zeile sind weiterhin nicht erlaubt.

### Entfernen aus dem Planraster

- Jeder eingeplante Block kann direkt aus dem Raster entfernt werden.
- Entfernen setzt den zugehoerigen `plan-eintrag` auf `aktiv: false` (soft delete).
- Dadurch bleibt Historie erhalten und die Lernsituation ist wieder in der Auswahlliste verfuegbar.
- Soft delete wird ueber `aktiv: false` umgesetzt, damit die gespeicherte Dauer weiterhin fuer die Auswahlliste nutzbar bleibt.

### API-Erweiterung

- Neuer Endpunkt zum Anlegen: `POST /api/plan-eintraege`
- Neuer Endpunkt zum Entfernen: `POST /api/plan-eintraege/:id/remove`
- Bestehender Endpunkt zum Verschieben/Resize bleibt: `POST /api/plan-eintraege/:id/move`

#### Request `POST /api/plan-eintraege`

```json
{
  "jahresplan_id": "jp-bg-it-j1",
  "lernsituation_id": "ls-006",
  "startwoche": 27,
  "endwoche": 27,
  "bezug_typ": "fach",
  "bezug_id": "fach-deutsch"
}
```

#### Request `POST /api/plan-eintraege/:id/remove`

```json
{
  "expected_revision": "2026-04-22T20:05:48.617855Z"
}
```

## 9) Bearbeiten per Drag-and-drop (Go-Schreibdienst)

### Ziel

- Jahresplan-Bloecke sollen im Browser verschoben werden koennen.
- Aenderungen werden in `content/plan-eintraege/*.md` persistiert.
- Hugo bleibt die Render-Schicht, Go uebernimmt nur das Schreiben.

### Redaktions- und Betriebsmodell (lokal + extern)

- Es gibt zwei Nutzungsmodi fuer den Jahresplan:
  - **Lokal (Editor-Modus):** Bearbeitung mit Drag-and-drop ueber Hugo + Go-Schreibdienst.
  - **Extern (Viewer-Modus):** rein statische Anzeige auf dem Webserver aus `public/`.
- Die Quelldaten (`content/*.md`) bleiben im Redaktionssystem und werden nicht direkt extern bearbeitet.
- Deployment erfolgt als Publishing-Schritt:
  1. lokal bearbeiten und speichern
  2. mit `hugo` Build erzeugen
  3. `public/` auf den externen Webserver veroeffentlichen
- Konsequenz: Externe Ansicht ist der letzte veroeffentlichte Stand, nicht zwingend der aktuelle lokale Arbeitsstand.

### Architektur

- Hugo-Frontend zeigt den Plan und sendet Aenderungen per HTTP.
- Go-Service laeuft lokal oder serverseitig als kleine API.
- Go-Service schreibt nur in `content/plan-eintraege/`.

### Ordner/Komponenten

- `tools/plan-write-service/`
- `tools/plan-write-service/cmd/server/main.go`
- `tools/plan-write-service/internal/api/`
- `tools/plan-write-service/internal/store/`
- `tools/plan-write-service/internal/validation/`

### API (MVP)

- `GET /api/health`
- `POST /api/plan-eintraege/:id/move`

#### Request `POST /api/plan-eintraege/:id/move`

```json
{
  "startwoche": 6,
  "endwoche": 8,
  "bezug_typ": "fach",
  "bezug_id": "fach-informatik",
  "expected_revision": "2026-04-21T10:31:00.000Z"
}
```

#### Response (Erfolg)

```json
{
  "ok": true,
  "id": "pe-0001",
  "updated": {
    "startwoche": 6,
    "endwoche": 8,
    "bezug_typ": "fach",
    "bezug_id": "fach-informatik"
  },
  "revision": "2026-04-21T10:35:43.221Z",
  "warnings": []
}
```

### Validierung

- `id`-Pattern: `^pe-[a-z0-9-]+$`
- `1 <= startwoche <= endwoche <= 40`
- `bezug_typ` ist `fach` oder `projekt`
- bei `bezug_typ` ist `bezug_id` Pflicht
- Referenzdatei fuer `bezug_id` muss existieren
- Zieldatei muss `type: plan_eintrag` haben

### Fehlercodes

- `400` ungueltiges Request-Format oder Feldwerte
- `404` Planeintrag oder Bezug nicht gefunden
- `409` `expected_revision` passt nicht
- `422` semantisch ungueltig
- `500` Schreib- oder Parserfehler

### Schreibstrategie

- Datei lesen
- Frontmatter parsen
- Felder patchen (`startwoche`, `endwoche`, optional Bezug)
- `revision` aktualisieren
- atomisch schreiben (temp + rename)

### Frontend-Integration

- Drag-Start: Planeintrag-ID und aktuelle Position merken.
- Drop: neue Wochen und neue Zeile berechnen.
- API-Aufruf `POST /move`.
- Bei Erfolg: UI-Refresh und neue Revision merken.
- Bei Fehler: Rueckmeldung mit Grund.

### Betriebsmodus

- Entwicklung: Hugo (`127.0.0.1:1313`) + Go-Service (`127.0.0.1:8787`)
- Produktion: Go-Service nur dann extern, wenn auch serverseitig redigiert werden soll
- CORS im MVP auf lokale Origins begrenzen

### Modus-Erkennung und UI-Verhalten

- Die Jahresplan-Seite setzt `editorMode` standardmaessig aus `hugo.IsServer`.
- Im Viewer-Modus werden Bearbeitungs-Elemente nicht gerendert:
  - keine klickbaren Zeilenkoepfe fuer Auswahl
  - keine Palette unter dem Raster
  - keine Remove-Buttons oder Resize-Handles
  - keine Drag-and-drop-Event-Logik
- Optionaler Hugo-Parameter:
  - `params.editor_mode: true` erzwingt Bearbeiten
  - `params.editor_mode: false` erzwingt schreibgeschuetzte Anzeige
- Optionaler Debug-Parameter:
  - `params.plan_debug: true` blendet den Debug-Block im Editor ein

### Lokale Ports und CORS

- Hugo-Frontend kann lokal auf `1313` oder ersatzweise `1314` laufen.
- Go-Schreibdienst laeuft auf `127.0.0.1:8787`.
- Fuer Browser-Requests muessen die Frontend-Origins im CORS-Allowlist des Go-Service stehen:
  - `http://127.0.0.1:1313`
  - `http://localhost:1313`
  - `http://127.0.0.1:1314`
  - `http://localhost:1314`

### Abgrenzung MVP

- nur `plan-eintraege` schreiben
- keine direkte Bearbeitung von `lernsituationen/*.md`
- keine Mehrbenutzer-Sperrlogik ausser `expected_revision`
