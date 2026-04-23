# Didaktischer Wizard

Markdown-basierter Hugo-Prototyp fuer die Verwaltung von Lernsituationen und die Darstellung von Jahresplaenen.

## Ziel

- Lernsituationen als strukturierte Markdown-Dateien pflegen
- Stammdaten fuer Bildungsgange, Faecher und Projekte verwalten
- Jahresplaene als 40-Wochen-Raster rendern
- Aus dem Jahresplan direkt in die Detailansicht einer Lernsituation springen

## Lokale Entwicklung

### Hugo-Server starten

```bash
hugo server
```

Danach ist die Seite lokal erreichbar unter:

```text
http://127.0.0.1:1313/
```

Vorteile von `hugo server`:

- automatische Aktualisierung bei Datei-Aenderungen
- realistische Vorschau im Browser
- saubere Behandlung von Links und Assets

### Produktions-Build erzeugen

```bash
hugo
```

Der fertige statische Build liegt danach in:

```text
public/
```

Diesen Inhalt laedt man spaeter auf den Webserver hoch.

## Projektstruktur

```text
content/
  bildungsgaenge/
  faecher/
  projekte/
  lernsituationen/
  jahresplaene/
  plan-eintraege/
layouts/
  _default/
  jahresplan/
  lernsituation/
  partials/
static/
  anlagen/
```

## Content-Modell

### Bildungsgang

- liegt unter `content/bildungsgaenge/`
- beschreibt einen 2-, 3- oder 4-jaehrigen Bildungsgang

### Fach

- liegt unter `content/faecher/`
- liefert Stammdaten fuer Fachzeilen im Jahresplan

### Projekt

- liegt unter `content/projekte/`
- liefert Stammdaten fuer Projektzeilen im Jahresplan

### Lernsituation

- liegt unter `content/lernsituationen/`
- enthaelt Nummer, Kurztext, Rasterdaten und Anlagen

### Jahresplan

- liegt unter `content/jahresplaene/`
- repraesentiert einen Plan fuer genau einen Bildungsgang und ein Ausbildungsjahr

### Plan-Eintrag

- liegt unter `content/plan-eintraege/`
- verknuepft Lernsituation, Jahresplan und Fach oder Projekt mit einem Wochenbereich
- wird nicht als eigene Seite gerendert, sondern nur als Datenquelle verwendet

## Wichtige Konventionen

- Dateiname = `id`, z. B. `ls-001.md`
- IDs sind stabil und in `kebab-case`
- Links als normale Markdown-Links, keine Obsidian-Wikilinks
- Wochenbereich immer `1..40`
- pro Kombination aus `bildungsgang_id` und `ausbildungsjahr` genau ein Jahresplan

## Relevante Templates

- `layouts/jahresplan/single.html`: Jahresplan-Seite
- `layouts/partials/jahresplan-matrix.html`: Tabellenmatrix fuer 40 Wochen
- `layouts/partials/jahresplan-cell.html`: einzelne Jahresplan-Zelle
- `layouts/lernsituation/single.html`: Detailseite einer Lernsituation

## Beispiel-Workflow

1. Neue Lernsituation in `content/lernsituationen/` anlegen.
2. Neuen Planeintrag in `content/plan-eintraege/` anlegen.
3. Mit `hugo server` die Vorschau pruefen.
4. Mit `hugo` den statischen Build erzeugen.
5. Inhalt aus `public/` deployen.

## Hinweise

- `public/` ist Build-Output, nicht die primaere Arbeitsumgebung.
- Fuer lokale Vorschau immer `hugo server` verwenden.
- Anlagen koennen unter `static/anlagen/` abgelegt und dann per URL verlinkt werden.
