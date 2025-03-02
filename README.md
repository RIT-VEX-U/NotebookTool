# Usage

```sh
go run . -entries /home/unknown/Clubs/VEX/Notebook/Entries -assets /home/unknown/Clubs/VEX/Notebook/Assets  -port 8989 -front-page frontpage.html
```

## Web Server

```sh
> go run . -assets <ASSET_PATH> -entries <ENTRIES_PATH>
Log Message
Log Message
Serving on http://localhost:XXXXX
```

## PDFs

```sh
> go run . -assets <ASSET_PATH> -entries <ENTRIES_PATH> -output <OUTPUT_FOLDER_FOR_PDFs>
```

# Structure

How it all works

1. Read all the entry files
2. analyze, sort, and order entries
3. Apply information to the template `page.tmpl.html`
4. Save that to temp/ directory
5. Save the css and supporting files to temp/
6. Start a web server that serves temp/
7. If enabled, use funny magic to tell chrome to save the page as a PDF

# Libraries and such

- [Goldmark](https://github.com/yuin/goldmark) Markdown Parser
  - A bunch of sub libraries for extensions
- [chromedp](https://github.com/chromedp/chromedp) Library to talk to google chrome with the chrome debug protocol
- [Go templates](https://pkg.go.dev/text/template) for taking the HTML from Goldmark and making the entire notebook from it
