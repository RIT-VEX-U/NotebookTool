# Usage


## Generate Notebook
```sh
./NotebookGen\
	-path /home/unknown/Clubs/VEX/Notebook/\
	-port 8989\
	-front-page frontpage.html\
	-frontmatter "How this Notebook is Organized"\
	-frontmatter "Meet the Team"\
	-frontmatter "Meet the Bears Behind the Bots"\
	-frontmatter "The Engineering Design Process"
```
Will think for a bit then print out a link like `localhost:8989/notebook.html`. One can then go to that website, press CTRL-P and print the notebook to a PDF

## Create Directory template
to setup a new notebook with all requisite templates and such. 
```sh
./NotebookGen -make-template ./
```
Will create a directory called Notebook in the current directory

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
